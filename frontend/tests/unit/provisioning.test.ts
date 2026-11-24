/**
 * Site Generator provisioning job (docs/provisioning-job-contract.md): SSE parsing, the event reducer,
 * the stream/poll watcher, the POST /jobs body, validation and shell escaping of the CLI command.
 */
import { describe, it, expect } from 'vitest'
import { execFileSync } from 'node:child_process'
import {
  applyProvisioningEvent,
  buildCreateJobBody,
  buildOnboardCommand,
  classifyProvisioningError,
  createJobWatcher,
  createSseParser,
  createWatcherState,
  formatDuration,
  isHealthTimeoutUnset,
  isNoTlsFallback,
  isStepWarning,
  isValidAdminName,
  isValidHealthTimeout,
  isValidSlug,
  mergeJobSnapshot,
  normalizeSteps,
  shellQuote,
  stepElapsedMs,
  tenantDatabases,
  tenantStoragePath,
  type ProvisionFormInput,
  type ProvisioningJob,
  type SseEvent
} from '../../utils/provisioning'
import { isPlatformAdminPath, PLATFORM_ADMIN_ROLES } from '../../utils/viewerPolicy'

const collect = (chunks: string[]) => {
  const events: SseEvent[] = []
  const p = createSseParser(e => events.push(e))
  chunks.forEach(c => p.push(c))
  p.end()
  return events
}

const baseJob = (over: Partial<ProvisioningJob> = {}): ProvisioningJob => ({
  id: 'job-1',
  status: 'running',
  slug: 'bai',
  mode: 'local',
  dry_run: false,
  created_at: '2026-11-24T10:00:00Z',
  steps: [],
  ...over
})

const streamOf = (...chunks: string[]) => {
  const enc = new TextEncoder()
  return new ReadableStream<Uint8Array>({
    start(c) {
      chunks.forEach(ch => c.enqueue(enc.encode(ch)))
      c.close()
    }
  })
}

const sse = (id: number, event: string, data: unknown) => `id: ${id}\nevent: ${event}\ndata: ${JSON.stringify(data)}\n\n`

describe('createSseParser', () => {
  it('parses id, event and data, with chunks split anywhere', () => {
    const text = sse(1, 'progress', { key: 'generate', status: 'running' }) + sse(2, 'log', { line: 'hello' })
    const whole = collect([text])
    const split = collect(text.split(''))
    expect(whole).toEqual(split)
    expect(whole).toEqual([
      { id: '1', event: 'progress', data: '{"key":"generate","status":"running"}' },
      { id: '2', event: 'log', data: '{"line":"hello"}' }
    ])
  })

  it('handles CRLF and CR line endings, including a CRLF split across chunks', () => {
    expect(collect(['id: 5\r', '\nevent: status\r\ndata: {}\r\n\r\n'])).toEqual([{ id: '5', event: 'status', data: '{}' }])
    expect(collect(['event: log\rdata: x\r\r'])).toEqual([{ id: undefined, event: 'log', data: 'x' }])
  })

  it('joins multi-line data, ignores comments, defaults the event name and drops an unterminated event', () => {
    expect(collect([': keep-alive\n\ndata: a\ndata: b\n\ndata: lost'])).toEqual([
      { id: undefined, event: 'message', data: 'a\nb' }
    ])
  })

  it('keeps a value that has no space after the colon and ignores unknown fields', () => {
    expect(collect(['event:log\nretry: 10\ndata:{"line":"x"}\n\n'])).toEqual([{ id: undefined, event: 'log', data: '{"line":"x"}' }])
  })
})

describe('applyProvisioningEvent', () => {
  it('upserts steps, tracks the highest seq, caps nothing below the limit and flags terminal status', () => {
    const state = createWatcherState()
    mergeJobSnapshot(state, baseJob({ status: 'queued' }))
    expect(state.job!.steps).toHaveLength(9)
    expect(state.job!.steps.every(s => s.status === 'pending')).toBe(true)

    applyProvisioningEvent(state, { id: '3', event: 'progress', data: JSON.stringify({ key: 'build', index: 3, title: 'Build images', status: 'running', message: 'Building 6 images', started_at: '2026-11-24T10:00:05Z' }) })
    expect(state.lastSeq).toBe(3)
    expect(state.job!.status).toBe('running')
    expect(state.job!.current_step).toBe('build')
    expect(state.job!.steps.find(s => s.key === 'build')).toMatchObject({ status: 'running', message: 'Building 6 images' })

    applyProvisioningEvent(state, { id: '4', event: 'log', data: JSON.stringify({ line: 'oops', stream: 'stderr', ts: 't' }) })
    expect(state.logs).toEqual([{ seq: 4, line: 'oops', stream: 'stderr', ts: 't' }])

    // An older id never moves lastSeq back
    applyProvisioningEvent(state, { id: '2', event: 'log', data: JSON.stringify({ line: 'late' }) })
    expect(state.lastSeq).toBe(4)

    expect(applyProvisioningEvent(state, { id: '9', event: 'status', data: JSON.stringify({ status: 'failed', error: 'build failed' }) })).toBe(true)
    expect(state.job).toMatchObject({ status: 'failed', error: 'build failed' })
  })

  it('never takes the password from a result event and keeps one captured from GET', () => {
    const state = createWatcherState()
    mergeJobSnapshot(state, baseJob({ status: 'succeeded', result: { admin_initial_password: 's3cret' } }))
    expect(state.initialPassword).toBe('s3cret')
    applyProvisioningEvent(state, { event: 'result', data: JSON.stringify({ frontend_url: 'https://bai.example', admin_initial_password: 'nope' }) })
    expect(state.job!.result).toMatchObject({ frontend_url: 'https://bai.example', admin_initial_password: 's3cret' })
    // A later GET returns null (already read): the captured value stays
    mergeJobSnapshot(state, baseJob({ status: 'succeeded', result: { admin_initial_password: null } }))
    expect(state.initialPassword).toBe('s3cret')
  })

  it('ignores malformed data and unknown events', () => {
    const state = createWatcherState()
    mergeJobSnapshot(state, baseJob())
    expect(applyProvisioningEvent(state, { event: 'progress', data: 'not json' })).toBe(false)
    expect(applyProvisioningEvent(state, { event: 'whatever', data: '{}' })).toBe(false)
    expect(state.logs).toHaveLength(0)
  })

  it('normalizeSteps keeps the fixed order and fills missing steps', () => {
    const steps = normalizeSteps([{ key: 'ingress', index: 9, title: 'x', status: 'skipped', message: 'local mode' }])
    expect(steps.map(s => s.key)).toEqual(['generate', 'databases', 'build', 'migrate', 'storage', 'email', 'start', 'health', 'ingress'])
    expect(steps[8]).toMatchObject({ status: 'skipped', message: 'local mode' })
  })
})

describe('createJobWatcher', () => {
  const noSleep = () => Promise.resolve()

  it('resumes the stream with ?since= after a drop and fetches the final job (password) once at the end', async () => {
    const sinces: number[] = []
    const streams = [
      () => streamOf(sse(1, 'progress', { key: 'generate', index: 1, status: 'running' }), sse(2, 'log', { line: 'one' })),
      () => streamOf(sse(3, 'progress', { key: 'generate', index: 1, status: 'done' }), sse(4, 'status', { status: 'succeeded', error: null }))
    ]
    let gets = 0
    const watcher = createJobWatcher('job-1', {
      getJob: async () => {
        gets++
        return gets === 1 ? baseJob() : baseJob({ status: 'succeeded', result: { admin_initial_password: 'pw-once' } })
      },
      openStream: async (_id, since) => {
        sinces.push(since)
        return streams[sinces.length - 1]!()
      },
      sleep: noSleep
    })
    await watcher.start()
    expect(sinces).toEqual([0, 2])
    expect(gets).toBe(2)
    expect(watcher.state.initialPassword).toBe('pw-once')
    expect(watcher.state.job!.status).toBe('succeeded')
    expect(watcher.state.logs.map(l => l.line)).toEqual(['one'])
    expect(watcher.state.connection).toBe('closed')
  })

  it('falls back to polling after repeated stream failures and stops when the job ends', async () => {
    let polls = 0
    const sleeps: number[] = []
    const watcher = createJobWatcher('job-1', {
      getJob: async () => {
        polls++
        return polls < 3 ? baseJob() : baseJob({ status: 'failed', error: 'provisioner restarted' })
      },
      openStream: async () => { throw Object.assign(new Error('buffered'), { statusCode: 502 }) },
      sleep: async (ms) => { sleeps.push(ms) },
      maxStreamFailures: 2,
      pollIntervalMs: 2000
    })
    await watcher.start()
    expect(sleeps).toContain(2000)
    expect(watcher.state.job).toMatchObject({ status: 'failed', error: 'provisioner restarted' })
    expect(watcher.state.connection).toBe('closed')
  })

  it('stops without streaming when the job does not exist', async () => {
    let opened = false
    const watcher = createJobWatcher('missing', {
      getJob: async () => { throw Object.assign(new Error('nf'), { statusCode: 404 }) },
      openStream: async () => {
        opened = true
        return streamOf()
      },
      sleep: noSleep
    })
    await watcher.start()
    expect(opened).toBe(false)
    expect(watcher.state.error?.kind).toBe('notFound')
  })
})

describe('isStepWarning', () => {
  it('flags skipped steps whose message reports a failure', () => {
    expect(isStepWarning({ status: 'skipped', message: 'Email provisioning FAILED: 401 from Resend' })).toBe(true)
    expect(isStepWarning({ status: 'skipped', message: 'TLS NOT CONFIGURED: certbot missing' })).toBe(true)
    expect(isStepWarning({ status: 'skipped', message: 'Local environment: skipped' })).toBe(false)
    expect(isStepWarning({ status: 'done', message: 'FAILED' })).toBe(false)
  })
})

describe('errors and time', () => {
  it('classifies HTTP errors', () => {
    expect(classifyProvisioningError({ statusCode: 503 })).toBe('unreachable')
    expect(classifyProvisioningError({ response: { status: 403 } })).toBe('forbidden')
    expect(classifyProvisioningError({ status: 409 })).toBe('conflict')
    expect(classifyProvisioningError({ statusCode: 422 })).toBe('validation')
    expect(classifyProvisioningError(new TypeError('Failed to fetch'))).toBe('network')
  })

  it('computes elapsed time for running and finished steps', () => {
    const now = Date.parse('2026-11-24T10:01:05Z')
    expect(stepElapsedMs({ status: 'running', started_at: '2026-11-24T10:00:00Z' }, now)).toBe(65000)
    expect(stepElapsedMs({ status: 'done', started_at: '2026-11-24T10:00:00Z', finished_at: '2026-11-24T10:00:24Z' }, now)).toBe(24000)
    expect(stepElapsedMs({ status: 'pending' }, now)).toBeNull()
    expect(formatDuration(65000)).toBe('1m 05s')
    expect(formatDuration(4000)).toBe('4s')
    expect(formatDuration(null)).toBe('')
  })
})

describe('form → request and display helpers', () => {
  const form: ProvisionFormInput = {
    clientName: '  PT Bank Arta Indonesia ',
    slug: 'bai',
    mode: 'production',
    baseDomain: 'Example.com',
    autoPorts: false,
    fePort: 3015,
    apiPort: 8095,
    adminName: 'Siti CAE',
    adminEmail: 'cae@bai.co.id',
    withDemoData: false,
    provisionEmail: true,
    dryRun: true
  }

  it('sends exactly the contract fields', () => {
    expect(buildCreateJobBody(form)).toEqual({
      client_name: 'PT Bank Arta Indonesia',
      slug: 'bai',
      mode: 'production',
      domain: 'bai.example.com',
      fe_port: 3015,
      api_port: 8095,
      admin_name: 'Siti CAE',
      admin_email: 'cae@bai.co.id',
      with_demo_data: false,
      provision_email: true,
      dry_run: true
    })
  })

  it('omits ports when automatic and domain/email in local mode', () => {
    const body = buildCreateJobBody({ ...form, mode: 'local', autoPorts: true, dryRun: false })
    expect(body).not.toHaveProperty('fe_port')
    expect(body).not.toHaveProperty('api_port')
    expect(body).not.toHaveProperty('domain')
    expect(body.provision_email).toBe(false)
    expect(body.dry_run).toBe(false)
  })

  it('validates the slug with ^[a-z][a-z0-9]{1,30}$', () => {
    for (const ok of ['bai', 'b1', 'a'.repeat(31)]) expect(isValidSlug(ok)).toBe(true)
    for (const bad of ['', 'b', '1bai', 'bai-x', 'Bai', 'b ai', 'a'.repeat(32), 'bai;rm']) expect(isValidSlug(bad)).toBe(false)
  })

  it('lists the 4 tenant databases and the script storage paths', () => {
    expect(tenantDatabases('bai')).toEqual(['rb_audit_auth_bai', 'rb_audit_audit_bai', 'rb_audit_master_bai', 'rb_audit_risk_bai'])
    expect(tenantStoragePath('bai', 'production')).toBe('/var/data/auditsphere/storage/bai')
    expect(tenantStoragePath('bai', 'local')).toBe('backend/audit-service/uploads/bai')
  })

  it('guards the Site Generator path for platform admins only', () => {
    expect(isPlatformAdminPath('/site-generator')).toBe(true)
    expect(isPlatformAdminPath('/site-generator/?job=1')).toBe(true)
    expect(isPlatformAdminPath('/settings')).toBe(false)
    // Mirrors ProvisioningRoles in auth-service routes/registry.go (RequireRoles "ADMIN")
    expect(PLATFORM_ADMIN_ROLES).toEqual(['ADMIN'])
  })

  it('sends admin_name only when filled in, and validates its length when it is', () => {
    expect(buildCreateJobBody({ ...form, adminName: '   ' })).not.toHaveProperty('admin_name')
    expect(buildCreateJobBody({ ...form, adminName: ' Siti ' }).admin_name).toBe('Siti')
    expect(isValidAdminName('')).toBe(true)
    expect(isValidAdminName('   ')).toBe(true)
    expect(isValidAdminName('Si')).toBe(true)
    expect(isValidAdminName('a'.repeat(120))).toBe(true)
    expect(isValidAdminName('S')).toBe(false)
    expect(isValidAdminName('a'.repeat(121))).toBe(false)
  })

  it('sends allow_no_tls only when checked in production mode', () => {
    expect(buildCreateJobBody(form)).not.toHaveProperty('allow_no_tls')
    expect(buildCreateJobBody({ ...form, allowNoTls: false })).not.toHaveProperty('allow_no_tls')
    expect(buildCreateJobBody({ ...form, allowNoTls: true }).allow_no_tls).toBe(true)
    expect(buildCreateJobBody({ ...form, mode: 'local', allowNoTls: true })).not.toHaveProperty('allow_no_tls')
  })

  it('sends health_timeout_seconds only when set to a whole number in 30..900', () => {
    expect(buildCreateJobBody(form)).not.toHaveProperty('health_timeout_seconds')
    expect(buildCreateJobBody({ ...form, healthTimeoutSeconds: null })).not.toHaveProperty('health_timeout_seconds')
    expect(buildCreateJobBody({ ...form, healthTimeoutSeconds: 300 }).health_timeout_seconds).toBe(300)
    expect(buildCreateJobBody({ ...form, healthTimeoutSeconds: 30 }).health_timeout_seconds).toBe(30)
    expect(buildCreateJobBody({ ...form, healthTimeoutSeconds: 900 }).health_timeout_seconds).toBe(900)
    for (const bad of [29, 901, 120.5, Number.NaN, '' as unknown as number]) {
      expect(buildCreateJobBody({ ...form, healthTimeoutSeconds: bad })).not.toHaveProperty('health_timeout_seconds')
      expect(isValidHealthTimeout(bad)).toBe(false)
    }
    // A cleared number input counts as "use the default", not as an error
    for (const unset of [null, undefined, '', Number.NaN]) expect(isHealthTimeoutUnset(unset)).toBe(true)
    expect(isHealthTimeoutUnset(0)).toBe(false)
  })

  it('detects a production job that finished on plain HTTP (allow_no_tls, step 9 skipped)', () => {
    const job = (over: Partial<ProvisioningJob>, ingress: { status: 'done' | 'skipped', message?: string }): ProvisioningJob => ({
      id: 'j', status: 'succeeded', slug: 'bai', mode: 'production', dry_run: false, created_at: '2026-01-01T00:00:00Z',
      steps: normalizeSteps([{ key: 'ingress', index: 9, title: 'Configure nginx and TLS', ...ingress }]),
      result: { frontend_url: 'https://bai.example.com' },
      ...over
    })
    const noTls = 'TLS NOT CONFIGURED, plain HTTP only: DNS not ready. Fix it, then run: certbot --nginx -d bai.example.com'
    expect(isNoTlsFallback(job({ allow_no_tls: true, result: { frontend_url: 'http://bai.example.com' } }, { status: 'skipped', message: noTls }))).toBe(true)
    // Older job records without the flag are still recognised from the step message
    expect(isNoTlsFallback(job({}, { status: 'skipped', message: noTls }))).toBe(true)
    expect(isNoTlsFallback(job({ allow_no_tls: true }, { status: 'done' }))).toBe(false)
    // result.tls (true only when both URLs are https) wins over the heuristics
    expect(isNoTlsFallback(job({ result: { tls: false, frontend_url: 'http://bai.example.com' } }, { status: 'skipped', message: noTls }))).toBe(true)
    expect(isNoTlsFallback(job({ allow_no_tls: true, result: { tls: false } }, { status: 'done' }))).toBe(true)
    expect(isNoTlsFallback(job({ allow_no_tls: true, result: { tls: true, frontend_url: 'https://bai.example.com' } }, { status: 'skipped', message: noTls }))).toBe(false)
    expect(isNoTlsFallback(job({ mode: 'local', result: { tls: false } }, { status: 'skipped' }))).toBe(false)
    // Local mode always skips ingress and serves http://localhost: not a fallback
    expect(isNoTlsFallback(job({ mode: 'local', result: { frontend_url: 'http://bai.localhost:3015' } }, { status: 'skipped', message: 'local mode' }))).toBe(false)
    expect(isNoTlsFallback(null)).toBe(false)
  })
})

describe('shellQuote / buildOnboardCommand', () => {
  it('single-quotes every argument and escapes embedded quotes', () => {
    expect(shellQuote('bai')).toBe(`'bai'`)
    expect(shellQuote(`O'Brien & Sons`)).toBe(`'O'\\''Brien & Sons'`)
    expect(shellQuote('')).toBe(`''`)
    expect(shellQuote('$(rm -rf /) `id` "x" \\ ;|')).toBe(`'$(rm -rf /) \`id\` "x" \\ ;|'`)
  })

  it('round-trips hostile strings through a real POSIX shell unchanged', () => {
    let sh = ''
    try {
      execFileSync('sh', ['-c', 'true'])
      sh = 'sh'
    } catch {
      // no POSIX shell on this machine
    }
    if (!sh) return
    for (const s of [`O'Brien`, `a'b'c`, '$(whoami)', '`id`', 'a\nb', `'`, `''`, '"; rm -rf ~; "', 'tab\there']) {
      const out = execFileSync(sh, ['-c', `printf %s ${shellQuote(s)}`]).toString()
      expect(out).toBe(s)
    }
  })

  it('builds the command from the form, quoting the client name', () => {
    const cmd = buildOnboardCommand({
      clientName: `PT O'Neil "Tbk"`, slug: 'oneil', mode: 'local', baseDomain: '', autoPorts: false,
      fePort: 3015, apiPort: 8095, adminName: 'x', adminEmail: 'x@y.z', withDemoData: true, provisionEmail: false, dryRun: true
    })
    expect(cmd).toBe(`./scripts/onboard-tenant.sh '--dry-run' '--local' '--fe-port' '3015' '--api-port' '8095' '--admin-name' 'x' '--admin-email' 'x@y.z' '--with-demo-data' '--skip-email' '--' 'oneil' 'PT O'\\''Neil "Tbk"'`)
    const auto = buildOnboardCommand({
      clientName: '-Acme', slug: 'acme', mode: 'production', baseDomain: 'auditsphere.app', autoPorts: true,
      fePort: 3015, apiPort: 8095, adminName: 'x', adminEmail: 'x@y.z', withDemoData: false, provisionEmail: true, dryRun: false
    })
    // A client name starting with "-" comes after "--", so it can never be read as a flag
    expect(auto).toBe(`./scripts/onboard-tenant.sh '--domain' 'acme.auditsphere.app' '--admin-name' 'x' '--admin-email' 'x@y.z' '--empty-data' '--' 'acme' '-Acme'`)
  })

  it('adds --health-timeout <n> and --allow-no-tls (production only), and drops an empty admin name', () => {
    const base: ProvisionFormInput = {
      clientName: 'Acme', slug: 'acme', mode: 'production', baseDomain: 'auditsphere.app', autoPorts: true,
      fePort: null, apiPort: null, adminName: '', adminEmail: 'x@y.z', withDemoData: false, provisionEmail: true, dryRun: false
    }
    expect(buildOnboardCommand({ ...base, allowNoTls: true, healthTimeoutSeconds: 300 }))
      .toBe(`./scripts/onboard-tenant.sh '--domain' 'acme.auditsphere.app' '--admin-email' 'x@y.z' '--empty-data' '--health-timeout' '300' '--allow-no-tls' '--' 'acme' 'Acme'`)
    // Local mode never gets --allow-no-tls; an out-of-range timeout is left to the script default
    const local = buildOnboardCommand({ ...base, mode: 'local', allowNoTls: true, healthTimeoutSeconds: 5 })
    expect(local).not.toContain('--allow-no-tls')
    expect(local).not.toContain('--health-timeout')
    expect(buildOnboardCommand(base)).not.toMatch(/--allow-no-tls|--health-timeout|--admin-name/)
  })
})
