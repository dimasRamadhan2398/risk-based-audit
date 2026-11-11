/**
 * MFA setup helpers for Settings > Security (components/settings/mfa.vue).
 */

export const TOTP_LENGTH = 6

/** Digits only, capped at the TOTP length (pasted "123 456" becomes "123456"). */
export const sanitizeTotp = (value: unknown): string =>
  String(value ?? '').replace(/\D/g, '').slice(0, TOTP_LENGTH)

/**
 * True when auth-service says MFA is already on (POST /mfa/enroll or /mfa/verify). Seen as
 * `{ code: 'MFA_ALREADY_ENABLED', message }` from the MFA controller and `{ error: { code, message } }`
 * from the shared response helper; the message is checked too in case the code changes.
 */
interface ErrorBody {
  code?: unknown
  message?: unknown
  error?: unknown
}

/** Error code from either backend shape: `{ code }` (MFA controller) or `{ error: { code } }` (shared helper). */
export function mfaErrorCode(err: unknown): string {
  const data = (err as { data?: ErrorBody } | null | undefined)?.data
  const nested = data?.error && typeof data.error === 'object' ? (data.error as ErrorBody) : undefined
  const code = typeof data?.code === 'string' ? data.code : typeof nested?.code === 'string' ? nested.code : ''
  return code.toUpperCase()
}

export function isMfaAlreadyEnabledError(err: unknown): boolean {
  if (mfaErrorCode(err) === 'MFA_ALREADY_ENABLED') return true
  const e = err as { data?: ErrorBody, message?: unknown } | null | undefined
  const data = e?.data
  const nested = data?.error && typeof data.error === 'object' ? (data.error as ErrorBody) : undefined
  const messages = [data?.message, nested?.message, typeof data?.error === 'string' ? data.error : '', e?.message]
  return messages.some(m => typeof m === 'string' && /already\s+enabled/i.test(m))
}

/** 409 MFA_SETUP_CHANGED: a newer setup replaced the secret, so the QR code on screen is stale. */
export const isMfaSetupChangedError = (err: unknown): boolean => mfaErrorCode(err) === 'MFA_SETUP_CHANGED'

/**
 * Wraps an async action so overlapping calls (Enter + click, or the auto-submit on the last digit
 * plus a click) share the first run instead of sending a second request.
 */
export function singleFlight<T>(fn: () => Promise<T>): (() => Promise<T>) & { readonly running: boolean } {
  let current: Promise<T> | null = null
  const run = (() => {
    if (current) return current
    current = (async () => {
      try {
        return await fn()
      } finally {
        current = null
      }
    })()
    return current
  }) as (() => Promise<T>) & { readonly running: boolean }
  Object.defineProperty(run, 'running', { get: () => current !== null })
  return run
}
