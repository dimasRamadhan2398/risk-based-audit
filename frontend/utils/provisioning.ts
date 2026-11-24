/**
 * Tenant provisioning job (Site Generator): types, SSE parsing, event reducer, validation,
 * request-body builder, shell escaping and the stream/poll watcher.
 *
 * Shape and behaviour follow docs/provisioning-job-contract.md. Everything here is plain TS
 * (no Nuxt imports) so it can be unit tested; composables/useProvisioning.ts wires it to $fetch.
 */

export type ProvisioningMode = 'local' | 'production'
export type JobStatus = 'queued' | 'running' | 'succeeded' | 'failed'
export type StepStatus = 'pending' | 'running' | 'done' | 'skipped' | 'failed'

export interface ProvisioningStep {
  key: string
  index: number
  title: string
  status: StepStatus
  message?: string | null
  started_at?: string | null
  finished_at?: string | null
}

export interface ProvisioningResult {
  slug?: string
  frontend_url?: string
  api_url?: string
  admin_username?: string
  admin_email?: string
  /** True only when both URLs are https. False in production = TLS not configured (allow_no_tls fallback). */
  tls?: boolean
  /** Returned once, only to the job's creator, then null. */
  admin_initial_password?: string | null
}

export interface ProvisioningJob {
  id: string
  status: JobStatus
  slug: string
  mode: ProvisioningMode
  dry_run: boolean
  /** Production only: step 9 is `skipped` (plain HTTP) instead of failing when DNS/certbot is not ready. */
  allow_no_tls?: boolean
  /** Seconds step 8 waited for health; absent when the script default applied. */
  health_timeout_seconds?: number | null
  created_by?: string
  created_at: string
  started_at?: string | null
  finished_at?: string | null
  current_step?: string | null
  steps: ProvisioningStep[]
  error?: string | null
  result?: ProvisioningResult | null
}

/** POST /jobs body. Only these fields are sent (contract). */
export interface CreateJobBody {
  client_name: string
  slug: string
  mode: ProvisioningMode
  domain?: string
  fe_port?: number
  api_port?: number
  admin_name?: string
  admin_email: string
  with_demo_data: boolean
  provision_email: boolean
  dry_run: boolean
  /** Sent only when true, and only in production mode. */
  allow_no_tls?: boolean
  /** Sent only when set (HEALTH_TIMEOUT_MIN..MAX); the script default applies otherwise. */
  health_timeout_seconds?: number
}

export interface NextPorts {
  fe_port: number
  api_port: number
}

export interface LogLine {
  seq: number
  line: string
  stream: 'stdout' | 'stderr'
  ts?: string
}

/** The 9 steps in fixed order (contract "Steps"). */
export const PROVISIONING_STEPS: ReadonlyArray<{ key: string, index: number, title: string }> = [
  { key: 'generate', index: 1, title: 'Generate tenant stack' },
  { key: 'databases', index: 2, title: 'Create databases' },
  { key: 'build', index: 3, title: 'Build images' },
  { key: 'migrate', index: 4, title: 'Migrate and seed' },
  { key: 'storage', index: 5, title: 'Create storage' },
  { key: 'email', index: 6, title: 'Provision email domain' },
  { key: 'start', index: 7, title: 'Start containers' },
  { key: 'health', index: 8, title: 'Wait for services to be healthy' },
  { key: 'ingress', index: 9, title: 'Configure nginx and TLS' }
]

/** Databases onboard-tenant.sh creates (analytics-service opens no per-tenant DB). */
export const TENANT_DB_SERVICES = ['auth', 'audit', 'master', 'risk'] as const

export const SLUG_PATTERN = /^[a-z][a-z0-9]{1,30}$/
export const CLIENT_NAME_MIN = 2
export const CLIENT_NAME_MAX = 120
export const PORT_MIN = 1024
export const PORT_MAX = 65535
export const DEFAULT_BASE_DOMAIN = 'auditsphere.app'
export const MAX_LOG_LINES = 2000
export const ADMIN_NAME_MIN = 2
export const ADMIN_NAME_MAX = 120
/** health_timeout_seconds bounds (contract); onboard-tenant.sh defaults to 120 when it is omitted. */
export const HEALTH_TIMEOUT_MIN = 30
export const HEALTH_TIMEOUT_MAX = 900
export const HEALTH_TIMEOUT_DEFAULT = 120

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const HOSTNAME_PATTERN = /^(?=.{1,253}$)(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$/

export const isValidSlug = (slug: string) => SLUG_PATTERN.test(slug)
export const isValidEmail = (email: string) => EMAIL_PATTERN.test(email.trim())
export const isValidHostname = (host: string) => HOSTNAME_PATTERN.test(host.trim().toLowerCase())
export const isValidClientName = (name: string) => {
  const n = name.trim().length
  return n >= CLIENT_NAME_MIN && n <= CLIENT_NAME_MAX
}
export const isValidPort = (port: unknown): port is number =>
  typeof port === 'number' && Number.isInteger(port) && port >= PORT_MIN && port <= PORT_MAX
/** admin_name is optional: empty is valid, otherwise ADMIN_NAME_MIN..MAX characters after trimming. */
export const isValidAdminName = (name: string | null | undefined) => {
  const n = (name || '').trim().length
  return n === 0 || (n >= ADMIN_NAME_MIN && n <= ADMIN_NAME_MAX)
}
export const isValidHealthTimeout = (s: unknown): s is number =>
  typeof s === 'number' && Number.isInteger(s) && s >= HEALTH_TIMEOUT_MIN && s <= HEALTH_TIMEOUT_MAX
/** Empty (null/undefined/NaN, e.g. a cleared number input) means "use the script default". */
export const isHealthTimeoutUnset = (s: unknown) =>
  s === null || s === undefined || s === '' || (typeof s === 'number' && Number.isNaN(s))

/**
 * The job ran with allow_no_tls and step 9 was skipped: the tenant is live on plain HTTP
 * ("TLS NOT CONFIGURED ..."), with http:// result URLs.
 */
export const isNoTlsFallback = (job: Pick<ProvisioningJob, 'mode' | 'allow_no_tls' | 'steps' | 'result'> | null | undefined) => {
  // Local mode always skips ingress and serves http://localhost; that is not a fallback
  if (!job || job.mode !== 'production') return false
  // result.tls is authoritative when present; the checks below cover older job records without it
  if (typeof job.result?.tls === 'boolean') return !job.result.tls
  const ingress = (job.steps || []).find(s => s.key === 'ingress')
  if (ingress?.status !== 'skipped') return false
  return !!job.allow_no_tls || /NOT CONFIGURED/.test(ingress.message || '')
    || /^http:\/\//.test(job.result?.frontend_url || '')
}

/** Best-effort slug suggestion from a client name (still validated by SLUG_PATTERN). */
export const suggestSlug = (name: string) =>
  name.toLowerCase().replace(/[^a-z0-9]/g, '').replace(/^[0-9]+/, '').slice(0, 31)

export const tenantDatabases = (slug: string) => TENANT_DB_SERVICES.map(s => `rb_audit_${s}_${slug}`)

/** Evidence storage silo, as onboard-tenant.sh computes it. */
export const tenantStoragePath = (slug: string, mode: ProvisioningMode) =>
  mode === 'local' ? `backend/audit-service/uploads/${slug}` : `/var/data/auditsphere/storage/${slug}`

/** Tenant domain for production, or null in local mode. */
export const tenantDomain = (f: Pick<ProvisionFormInput, 'slug' | 'mode' | 'baseDomain'>): string | null =>
  f.mode === 'production' ? `${f.slug.trim()}.${f.baseDomain.trim().toLowerCase()}` : null

/**
 * A step reported as skipped although something went wrong: email is never fatal ("Email provisioning
 * FAILED ...") and ingress can be skipped with --allow-no-tls ("TLS NOT CONFIGURED ...").
 */
export const isStepWarning = (step: Pick<ProvisioningStep, 'status' | 'message'>) =>
  step.status === 'skipped' && /\bFAILED\b|NOT CONFIGURED/.test(step.message || '')

export const isTerminalJobStatus = (status: string | undefined | null): status is 'succeeded' | 'failed' =>
  status === 'succeeded' || status === 'failed'

export interface ProvisionFormInput {
  clientName: string
  slug: string
  mode: ProvisioningMode
  /** Parent domain (production). The script requires domain = <slug>.<base>, API = api-<slug>.<base>. */
  baseDomain: string
  autoPorts: boolean
  fePort: number | null
  apiPort: number | null
  adminName: string
  adminEmail: string
  withDemoData: boolean
  provisionEmail: boolean
  dryRun: boolean
  /** Production only; ignored in local mode. Optional so older callers default to false. */
  allowNoTls?: boolean
  /** Empty/null = script default (HEALTH_TIMEOUT_DEFAULT). */
  healthTimeoutSeconds?: number | null
}

/** Builds the POST /jobs body: contract fields only, optional ones omitted when not applicable. */
export const buildCreateJobBody = (f: ProvisionFormInput): CreateJobBody => {
  const body: CreateJobBody = {
    client_name: f.clientName.trim(),
    slug: f.slug.trim(),
    mode: f.mode,
    admin_email: f.adminEmail.trim(),
    with_demo_data: f.withDemoData,
    // Email and ingress are skipped in local mode anyway
    provision_email: f.mode === 'production' && f.provisionEmail,
    dry_run: f.dryRun
  }
  const adminName = f.adminName.trim()
  if (adminName) body.admin_name = adminName
  const domain = tenantDomain(f)
  if (domain) body.domain = domain
  if (!f.autoPorts && isValidPort(f.fePort) && isValidPort(f.apiPort)) {
    body.fe_port = f.fePort
    body.api_port = f.apiPort
  }
  if (f.mode === 'production' && f.allowNoTls) body.allow_no_tls = true
  if (isValidHealthTimeout(f.healthTimeoutSeconds)) body.health_timeout_seconds = f.healthTimeoutSeconds
  return body
}

// ---------------------------------------------------------------------------
// Shell escaping (for the secondary "Show CLI command" option)
// ---------------------------------------------------------------------------

/** POSIX single-quote escaping: wraps in '...' and turns every ' into '\''. Safe for any string. */
export const shellQuote = (arg: string | number) => `'${String(arg).replace(/'/g, `'\\''`)}'`

/**
 * The equivalent manual command, in the contract's form: options, then `--`, then <slug> <client_name>
 * (so a client name starting with "-" is never read as a flag). Every argument is single-quoted.
 */
export const buildOnboardCommand = (f: ProvisionFormInput): string => {
  const args: string[] = []
  if (f.dryRun) args.push('--dry-run')
  const domain = tenantDomain(f)
  if (domain) args.push('--domain', domain)
  else args.push('--local')
  if (!f.autoPorts && isValidPort(f.fePort) && isValidPort(f.apiPort)) {
    args.push('--fe-port', String(f.fePort), '--api-port', String(f.apiPort))
  }
  if (f.adminName.trim()) args.push('--admin-name', f.adminName.trim())
  args.push('--admin-email', f.adminEmail.trim())
  args.push(f.withDemoData ? '--with-demo-data' : '--empty-data')
  if (f.mode === 'local' || !f.provisionEmail) args.push('--skip-email')
  if (isValidHealthTimeout(f.healthTimeoutSeconds)) args.push('--health-timeout', String(f.healthTimeoutSeconds))
  if (f.mode === 'production' && f.allowNoTls) args.push('--allow-no-tls')
  args.push('--', f.slug.trim(), f.clientName.trim())
  return ['./scripts/onboard-tenant.sh', ...args.map(shellQuote)].join(' ')
}

// ---------------------------------------------------------------------------
// Server-Sent Events parsing
// ---------------------------------------------------------------------------

export interface SseEvent {
  id?: string
  event: string
  data: string
}

/**
 * Incremental SSE parser (WHATWG event-stream rules, minus `retry`). Feed it decoded text chunks
 * in any split; it calls onEvent once per complete event (blank-line terminated).
 */
export const createSseParser = (onEvent: (evt: SseEvent) => void) => {
  let buffer = ''
  let id: string | undefined
  let event = ''
  let data: string[] = []
  let hasData = false

  const dispatch = () => {
    if (hasData) onEvent({ id, event: event || 'message', data: data.join('\n') })
    event = ''
    data = []
    hasData = false
    // id persists across events per spec, but we pass it only with the event that set it
    id = undefined
  }

  const processLine = (line: string) => {
    if (line === '') return dispatch()
    if (line.startsWith(':')) return
    const colon = line.indexOf(':')
    const field = colon === -1 ? line : line.slice(0, colon)
    let value = colon === -1 ? '' : line.slice(colon + 1)
    if (value.startsWith(' ')) value = value.slice(1)
    if (field === 'event') event = value
    else if (field === 'data') {
      data.push(value)
      hasData = true
    } else if (field === 'id') id = value
  }

  return {
    push(chunk: string) {
      buffer += chunk
      let nl: number
      // Lines end in \n, \r\n or \r
      while ((nl = buffer.search(/\r\n|\r|\n/)) !== -1) {
        // A trailing lone \r may be the first half of \r\n: wait for more input
        if (buffer[nl] === '\r' && nl === buffer.length - 1) break
        const line = buffer.slice(0, nl)
        const sepLen = buffer.startsWith('\r\n', nl) ? 2 : 1
        buffer = buffer.slice(nl + sepLen)
        processLine(line)
      }
    },
    /** End of stream: an unterminated event is discarded, as the spec says. */
    end() {
      // A held-back trailing \r is a line end after all
      if (buffer.endsWith('\r')) processLine(buffer.slice(0, -1))
      buffer = ''
      event = ''
      data = []
      hasData = false
      id = undefined
    }
  }
}

// ---------------------------------------------------------------------------
// Job state + event reducer
// ---------------------------------------------------------------------------

/** Ensures all 9 steps exist (pending when the backend has not reported them yet), in order. */
export const normalizeSteps = (steps: ProvisioningStep[] | null | undefined): ProvisioningStep[] => {
  const byKey = new Map((steps || []).map(s => [s.key, s]))
  const known = PROVISIONING_STEPS.map(def => ({ ...def, status: 'pending' as StepStatus, ...(byKey.get(def.key) || {}) }))
  const extra = (steps || []).filter(s => !PROVISIONING_STEPS.some(d => d.key === s.key))
  return [...known, ...extra].sort((a, b) => a.index - b.index)
}

export interface WatcherState {
  job: ProvisioningJob | null
  logs: LogLine[]
  lastSeq: number
  connection: 'idle' | 'connecting' | 'streaming' | 'reconnecting' | 'polling' | 'closed'
  /** Captured from the first GET that carries it; the backend clears it after one read. */
  initialPassword: string | null
  error: { kind: ProvisioningErrorKind, status?: number, message?: string } | null
}

export const createWatcherState = (): WatcherState => ({
  job: null,
  logs: [],
  lastSeq: 0,
  connection: 'idle',
  initialPassword: null,
  error: null
})

type JsonObject = Record<string, unknown>

/** Parses an event's data; only JSON objects count. */
const parseJsonObject = (s: string): JsonObject | undefined => {
  try {
    const v: unknown = JSON.parse(s)
    return v && typeof v === 'object' && !Array.isArray(v) ? v as JsonObject : undefined
  } catch {
    return undefined
  }
}

/** Merges a GET /jobs/{id} snapshot into the state, keeping a password already captured. */
export const mergeJobSnapshot = (state: WatcherState, job: ProvisioningJob) => {
  const pw = job.result?.admin_initial_password
  if (pw) state.initialPassword = pw
  state.job = { ...job, steps: normalizeSteps(job.steps) }
}

/** Applies one SSE event to the state. Returns true when the event was a terminal status. */
export const applyProvisioningEvent = (state: WatcherState, evt: SseEvent): boolean => {
  if (evt.id !== undefined) {
    const seq = Number(evt.id)
    if (Number.isFinite(seq) && seq > state.lastSeq) state.lastSeq = seq
  }
  const data = parseJsonObject(evt.data)
  if (!data) return false

  const job = state.job
  switch (evt.event) {
    case 'progress': {
      if (!job || typeof data.key !== 'string') return false
      const step = data as unknown as ProvisioningStep
      const steps = normalizeSteps(job.steps)
      const i = steps.findIndex(s => s.key === step.key)
      if (i === -1) steps.push(step)
      else steps[i] = { ...steps[i]!, ...step }
      job.steps = steps.sort((a, b) => a.index - b.index)
      if (step.status === 'running') job.current_step = step.key
      if (job.status === 'queued') job.status = 'running'
      return false
    }
    case 'log': {
      state.logs.push({
        seq: Number(evt.id) || state.lastSeq,
        line: String(data.line ?? ''),
        stream: data.stream === 'stderr' ? 'stderr' : 'stdout',
        ts: typeof data.ts === 'string' ? data.ts : undefined
      })
      if (state.logs.length > MAX_LOG_LINES) state.logs.splice(0, state.logs.length - MAX_LOG_LINES)
      return false
    }
    case 'status': {
      if (!job || typeof data.status !== 'string') return false
      job.status = data.status as JobStatus
      job.error = typeof data.error === 'string' ? data.error : null
      return isTerminalJobStatus(data.status)
    }
    case 'result': {
      if (!job) return false
      // The stream never carries the password; keep whatever a GET already returned
      const { admin_initial_password: _pw, ...rest } = data as ProvisioningResult
      job.result = { ...(job.result || {}), ...rest }
      return false
    }
    default:
      return false
  }
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

export type ProvisioningErrorKind = 'unreachable' | 'forbidden' | 'notFound' | 'conflict' | 'validation' | 'network' | 'unknown'

/** What ofetch's FetchError (and plain errors) may carry. */
interface HttpErrorLike {
  statusCode?: number
  status?: number
  response?: { status?: number }
  data?: { message?: unknown, error?: unknown }
}

export const provisioningErrorStatus = (e: unknown): number | undefined => {
  const err = (e ?? {}) as HttpErrorLike
  return err.statusCode ?? err.status ?? err.response?.status ?? undefined
}

export const provisioningErrorMessage = (e: unknown): string | undefined => {
  const d = ((e ?? {}) as HttpErrorLike).data
  const nested = d?.error && typeof d.error === 'object' ? (d.error as { message?: unknown }).message : undefined
  const msg = nested ?? d?.message ?? d?.error
  return typeof msg === 'string' && msg ? msg : undefined
}

export const classifyProvisioningError = (e: unknown): ProvisioningErrorKind => {
  const s = provisioningErrorStatus(e)
  if (s === 502 || s === 503 || s === 504) return 'unreachable'
  if (s === 401 || s === 403) return 'forbidden'
  if (s === 404) return 'notFound'
  if (s === 409) return 'conflict'
  if (s === 400 || s === 422) return 'validation'
  if (s === undefined) return 'network'
  return 'unknown'
}

// ---------------------------------------------------------------------------
// Time
// ---------------------------------------------------------------------------

const toMs = (iso?: string | null) => {
  if (!iso) return NaN
  return Date.parse(iso)
}

/** Elapsed ms for a step: finished - started, or now - started while running. null when unknown. */
export const stepElapsedMs = (step: Pick<ProvisioningStep, 'status' | 'started_at' | 'finished_at'>, now: number): number | null => {
  const start = toMs(step.started_at)
  if (Number.isNaN(start)) return null
  const end = step.status === 'running' ? now : toMs(step.finished_at)
  if (Number.isNaN(end)) return step.status === 'running' ? Math.max(0, now - start) : null
  return Math.max(0, end - start)
}

/** 4s, 1m 05s, 1h 02m. */
export const formatDuration = (ms: number | null): string => {
  if (ms === null || !Number.isFinite(ms)) return ''
  const total = Math.floor(ms / 1000)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h) return `${h}h ${String(m).padStart(2, '0')}m`
  if (m) return `${m}m ${String(s).padStart(2, '0')}s`
  return `${s}s`
}

// ---------------------------------------------------------------------------
// Watcher: SSE stream with resume, falling back to polling
// ---------------------------------------------------------------------------

export interface WatcherDeps {
  getJob: (id: string) => Promise<ProvisioningJob>
  /** Opens GET /jobs/{id}/events?since={since}; resolves to the response body stream. */
  openStream: (id: string, since: number, signal: AbortSignal) => Promise<ReadableStream<Uint8Array>>
  sleep?: (ms: number, signal: AbortSignal) => Promise<void>
  pollIntervalMs?: number
  /** Consecutive failed connections before switching to polling. */
  maxStreamFailures?: number
  /** While polling, try the stream again this often (it also replays missed log lines). */
  streamRetryMs?: number
}

const defaultSleep = (ms: number, signal: AbortSignal) => new Promise<void>((resolve) => {
  if (signal.aborted) return resolve()
  const t = setTimeout(resolve, ms)
  signal.addEventListener('abort', () => {
    clearTimeout(t)
    resolve()
  }, { once: true })
})

export const createJobWatcher = (jobId: string, deps: WatcherDeps, state: WatcherState = createWatcherState()) => {
  const sleep = deps.sleep || defaultSleep
  const pollIntervalMs = deps.pollIntervalMs ?? 2000
  const maxStreamFailures = deps.maxStreamFailures ?? 3
  const streamRetryMs = deps.streamRetryMs ?? 15000
  const controller = new AbortController()
  let terminalSeen = false
  let running: Promise<void> | null = null

  const setError = (e: unknown) => {
    state.error = { kind: classifyProvisioningError(e), status: provisioningErrorStatus(e), message: provisioningErrorMessage(e) }
  }

  const fetchSnapshot = async (): Promise<boolean> => {
    try {
      mergeJobSnapshot(state, await deps.getJob(jobId))
      state.error = null
      return true
    } catch (e) {
      setError(e)
      return false
    }
  }

  /** One connection. 'events' = stream ended after delivering events, 'empty' = ended with none. */
  const streamOnce = async (): Promise<'events' | 'empty' | 'error' | 'aborted'> => {
    let received = 0
    let reader: ReadableStreamDefaultReader<Uint8Array> | null = null
    const parser = createSseParser((evt) => {
      received++
      if (applyProvisioningEvent(state, evt)) terminalSeen = true
    })
    try {
      state.connection = state.lastSeq > 0 || state.connection === 'reconnecting' ? 'reconnecting' : 'connecting'
      const body = await deps.openStream(jobId, state.lastSeq, controller.signal)
      state.connection = 'streaming'
      state.error = null
      reader = body.getReader()
      const decoder = new TextDecoder()
      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        parser.push(decoder.decode(value, { stream: true }))
        if (terminalSeen) break
      }
      parser.push(decoder.decode())
      parser.end()
      return received > 0 ? 'events' : 'empty'
    } catch (e) {
      if (controller.signal.aborted) return 'aborted'
      setError(e)
      return received > 0 ? 'events' : 'error'
    } finally {
      try {
        reader?.cancel().catch(() => {})
      } catch {
        // already released
      }
    }
  }

  const finalize = async () => {
    // One GET after the end: final steps, and the password (the stream never carries it)
    await fetchSnapshot()
    state.connection = 'closed'
  }

  const run = async () => {
    const ok = await fetchSnapshot()
    if (!ok && (state.error?.kind === 'notFound' || state.error?.kind === 'forbidden')) {
      state.connection = 'closed'
      return
    }

    let failures = 0
    let mode: 'stream' | 'poll' = 'stream'
    let sincePollStart = 0

    while (!controller.signal.aborted) {
      if (mode === 'stream') {
        const outcome = await streamOnce()
        if (controller.signal.aborted || outcome === 'aborted') break
        if (terminalSeen) {
          await finalize()
          return
        }
        if (outcome === 'events') failures = 0
        else failures++
        if (failures >= maxStreamFailures) {
          mode = 'poll'
          sincePollStart = 0
          continue
        }
        state.connection = 'reconnecting'
        await sleep(Math.min(1000 * Math.max(1, failures), 5000), controller.signal)
        continue
      }

      state.connection = 'polling'
      await sleep(pollIntervalMs, controller.signal)
      if (controller.signal.aborted) break
      await fetchSnapshot()
      if (isTerminalJobStatus(state.job?.status)) {
        state.connection = 'closed'
        return
      }
      sincePollStart += pollIntervalMs
      if (sincePollStart >= streamRetryMs) {
        // One more stream attempt; back to polling right away if it fails again
        mode = 'stream'
        failures = maxStreamFailures - 1
        sincePollStart = 0
      }
    }
    if (state.connection !== 'closed') state.connection = 'closed'
  }

  return {
    state,
    start() {
      if (!running) running = run()
      return running
    },
    stop() {
      controller.abort()
      if (state.connection !== 'closed') state.connection = 'closed'
    }
  }
}
