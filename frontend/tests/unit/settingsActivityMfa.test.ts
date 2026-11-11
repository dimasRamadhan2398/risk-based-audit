// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useAuthStore } from '~/stores/auth'
import { ZERO_UUID, describeDevice, normalizeTrustedDevices } from '~/utils/trustedDevices'
import { isMfaAlreadyEnabledError, isMfaSetupChangedError, sanitizeTotp, singleFlight } from '~/utils/mfa'

global.$fetch = vi.fn()

const zeroRow = { id: ZERO_UUID, device_fingerprint: '', device_name: '', device_type: '', user_agent: '', ip_address: '', created_at: '0001-01-01T00:00:00Z' }
const macRow = {
  id: 'd1',
  device_fingerprint: 'abc12345',
  device_name: 'Chrome on macOS',
  device_type: 'desktop',
  user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36',
  ip_address: '10.0.0.5, 172.16.0.1',
  created_at: '2026-11-01T08:00:00Z'
}
const phoneRow = {
  id: 'd2',
  device_fingerprint: 'ffff0000',
  device_name: 'Safari on iOS',
  device_type: 'mobile',
  user_agent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1',
  ip_address: '10.0.0.9',
  created_at: '2026-11-03T08:00:00Z'
}

describe('normalizeTrustedDevices', () => {
  it('unwraps the { data: { data: [] } } envelope and drops the zero-value rows the service prepends', () => {
    const list = normalizeTrustedDevices({ success: true, message: 'Success', data: { data: [zeroRow, zeroRow, macRow, phoneRow] } })
    expect(list.map(d => d.id)).toEqual(['d2', 'd1'])
  })

  it('maps snake_case fields and keeps the first X-Forwarded-For address', () => {
    const [d] = normalizeTrustedDevices([macRow])
    expect(d).toMatchObject({ id: 'd1', fingerprint: 'abc12345', name: 'Chrome on macOS', type: 'desktop', ipAddress: '10.0.0.5' })
  })

  it('returns [] for empty or unexpected payloads', () => {
    expect(normalizeTrustedDevices(null)).toEqual([])
    expect(normalizeTrustedDevices({ data: null })).toEqual([])
    expect(normalizeTrustedDevices({ data: { data: 'nope' } })).toEqual([])
  })

  it('dedupes by id', () => {
    expect(normalizeTrustedDevices([macRow, { ...macRow }])).toHaveLength(1)
  })

  it('reads is_trusted and last_seen_at, and sorts by last activity', () => {
    const sessionOnly = { ...phoneRow, is_trusted: false, last_seen_at: '2026-11-03T09:00:00Z' }
    const trusted = { ...macRow, is_trusted: true, last_seen_at: '2026-11-04T10:00:00Z' }
    const list = normalizeTrustedDevices({ data: { data: [sessionOnly, trusted] } })
    expect(list.map(d => [d.id, d.trusted, d.lastSeenAt])).toEqual([
      ['d1', true, '2026-11-04T10:00:00Z'],
      ['d2', false, '2026-11-03T09:00:00Z']
    ])
  })

  it('treats rows from the old API (no is_trusted) as trusted unless session_only', () => {
    expect(normalizeTrustedDevices([macRow])[0].trusted).toBe(true)
    expect(normalizeTrustedDevices([{ ...macRow, session_only: true }])[0].trusted).toBe(false)
  })
})

describe('describeDevice', () => {
  it('reads browser, OS and form factor from the user agent', () => {
    expect(describeDevice({ userAgent: macRow.user_agent })).toEqual({ browser: 'Chrome', os: 'macOS', type: 'desktop' })
    expect(describeDevice({ userAgent: phoneRow.user_agent })).toEqual({ browser: 'Safari', os: 'iOS', type: 'mobile' })
  })

  it('falls back to the stored "<Browser> on <OS>" name and type', () => {
    expect(describeDevice({ name: 'Firefox on Linux', type: 'tablet' })).toEqual({ browser: 'Firefox', os: 'Linux', type: 'tablet' })
    expect(describeDevice({ name: 'Unknown Browser on Unknown OS' })).toEqual({ browser: '', os: '', type: 'desktop' })
  })
})

describe('MFA helpers', () => {
  it('recognises "already enabled" from both backend error shapes', () => {
    expect(isMfaAlreadyEnabledError({ data: { code: 'MFA_ALREADY_ENABLED', message: 'MFA is already enabled' } })).toBe(true)
    expect(isMfaAlreadyEnabledError({ data: { success: false, error: { code: 'MFA_ALREADY_ENABLED' } } })).toBe(true)
    expect(isMfaAlreadyEnabledError({ data: { message: 'MFA is already enabled' } })).toBe(true)
    expect(isMfaAlreadyEnabledError({ data: { code: 'INVALID_CODE', message: 'Invalid verification code' } })).toBe(false)
    expect(isMfaAlreadyEnabledError(null)).toBe(false)
  })

  it('recognises the 409 MFA_SETUP_CHANGED error', () => {
    expect(isMfaSetupChangedError({ status: 409, data: { code: 'MFA_SETUP_CHANGED', message: 'MFA setup was restarted' } })).toBe(true)
    expect(isMfaSetupChangedError({ data: { error: { code: 'MFA_SETUP_CHANGED' } } })).toBe(true)
    expect(isMfaSetupChangedError({ data: { code: 'MFA_SETUP_CHANGED' } }) && isMfaAlreadyEnabledError({ data: { code: 'MFA_SETUP_CHANGED' } })).toBe(false)
  })

  it('sanitizes pasted codes to 6 digits', () => {
    expect(sanitizeTotp('123 456')).toBe('123456')
    expect(sanitizeTotp('12a3-4567')).toBe('123456')
    expect(sanitizeTotp(undefined)).toBe('')
  })

  it('singleFlight shares one in-flight run between overlapping triggers', async () => {
    let resolve
    const fn = vi.fn(() => new Promise((r) => {
      resolve = r
    }))
    const run = singleFlight(fn)
    const a = run() // e.g. auto-submit on the 6th digit
    const b = run() // e.g. Enter / click right after
    expect(fn).toHaveBeenCalledTimes(1)
    expect(run.running).toBe(true)
    resolve('ok')
    expect(await a).toBe('ok')
    expect(await b).toBe('ok')
    expect(run.running).toBe(false)
    run()
    expect(fn).toHaveBeenCalledTimes(2)
  })

  it('singleFlight releases the lock after a failure', async () => {
    const fn = vi.fn().mockRejectedValueOnce(new Error('boom')).mockResolvedValueOnce('ok')
    const run = singleFlight(fn)
    await expect(run()).rejects.toThrow('boom')
    expect(await run()).toBe('ok')
  })
})

describe('auth store trusted devices', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('stores the normalized list and sends no hand-rolled Authorization header', async () => {
    const store = useAuthStore()
    store.token = 'tok'
    vi.mocked($fetch).mockResolvedValueOnce({ success: true, data: { data: [zeroRow, macRow] } })
    const list = await store.fetchTrustedDevices()
    expect(list.map(d => d.id)).toEqual(['d1'])
    expect(store.trustedDevices).toEqual(list)
    const [url, opts] = vi.mocked($fetch).mock.calls[0]
    expect(url).toMatch(/\/devices$/)
    expect(opts?.headers?.Authorization).toBeUndefined()
  })

  it('throws on failure and keeps the previous list', async () => {
    const store = useAuthStore()
    store.token = 'tok'
    store.trustedDevices = normalizeTrustedDevices([macRow])
    vi.mocked($fetch).mockRejectedValueOnce(new Error('500'))
    await expect(store.fetchTrustedDevices()).rejects.toThrow()
    expect(store.trustedDevices.map(d => d.id)).toEqual(['d1'])
  })
})

describe('settings activity / MFA copy', () => {
  const locale = (l: string) => JSON.parse(readFileSync(resolve(__dirname, `../../locales/${l}/common.json`), 'utf-8'))
  const source = (f: string) => readFileSync(resolve(__dirname, `../../components/settings/${f}`), 'utf-8')
  const keysIn = (src: string, prefix: string) =>
    [...new Set([...src.matchAll(new RegExp(`['"]settings\\.${prefix}\\.([A-Za-z0-9_]+)['"]`, 'g'))].map(m => m[1]))]

  it.each([['activity.vue', 'activity'], ['mfa.vue', 'mfa']])('every key used by %s exists in en and id', (file, prefix) => {
    const keys = keysIn(source(file), prefix)
    expect(keys.length).toBeGreaterThan(5)
    for (const l of ['en', 'id']) {
      const section = locale(l).settings[prefix]
      const missing = keys.filter(k => typeof section[k] !== 'string')
      expect(missing, `${l} missing`).toEqual([])
    }
  })

  it('the MFA form guards every submit path and sends this device with the verify call', () => {
    const src = source('mfa.vue')
    expect(src).toMatch(/const handleVerifySetup = singleFlight\(/)
    expect(src).toMatch(/const handleSetup = singleFlight\(/)
    expect(src).toMatch(/@submit\.prevent="handleVerifySetup"/)
    expect(src).toMatch(/:disabled="verifying \|\| !canSubmitCode"/)
    expect(src).toMatch(/device_fingerprint: device\.deviceFingerprint/)
    expect(src).not.toMatch(/Authorization: `Bearer/)
  })

  it('device tiles use the secondary colour, not primary/blue', () => {
    const src = source('activity.vue')
    expect(src).toMatch(/bg-secondary-500\/10/)
    expect(src).not.toMatch(/cyan-|blue-|color="primary"/)
  })

  it('the activity page no longer shows the database storage verification block', () => {
    const src = source('activity.vue')
    expect(src).not.toMatch(/verificationTitle|registeredYes|checkingDb|dbStatus|verifyDatabaseRegistration/)
  })
})
