/**
 * Utility to extract user-friendly error messages from API calls, $fetch / FetchError,
 * standard Error objects, or custom backend response structures.
 */

export interface ErrorDetails {
  message: string
  details?: string
  code?: string
  statusCode?: number
}

/**
 * Extracts a descriptive error message from any error object.
 * Priority:
 * 1. Nested Go response structure: error.data.error.message (+ error.data.error.details)
 * 2. String error field: error.data.error
 * 3. Message field: error.data.message or error.data.msg
 * 4. Raw string response body: error.data
 * 5. HTTP status: error.statusCode + error.statusMessage
 * 6. Standard Error message: error.message
 * 7. Provided fallback string
 */
export const extractErrorMessage = (error: any, fallback = 'Terjadi kesalahan pada sistem. Tolong tunggu beberapa saat lalu coba lagi.'): string => {
  if (!error) return fallback
  if (typeof error === 'string') return error

  // 1. Nested Go backend structure: { success: false, error: { code, message, details } }
  if (error.data?.error && typeof error.data.error === 'object') {
    const msg = error.data.error.message
    const details = error.data.error.details
    if (msg && details) {
      return `${msg}: ${details}`
    }
    if (msg) return msg
    if (details) return details
  }

  // 2. String error property: { error: "Something went wrong" }
  if (typeof error.data?.error === 'string' && error.data.error.trim()) {
    return error.data.error.trim()
  }

  // 3. Message property: { message: "Validation error" }
  if (typeof error.data?.message === 'string' && error.data.message.trim()) {
    return error.data.message.trim()
  }

  // 4. Msg property: { msg: "Validation error" }
  if (typeof error.data?.msg === 'string' && error.data.msg.trim()) {
    return error.data.msg.trim()
  }

  // 5. Raw string response body
  if (typeof error.data === 'string' && error.data.trim()) {
    return error.data.trim()
  }

  // 6. HTTP Status Code + Status Message from FetchError
  if (error.statusCode && error.statusMessage) {
    return `${error.statusCode} ${error.statusMessage}`
  }
  if (error.statusMessage && typeof error.statusMessage === 'string') {
    return error.statusMessage
  }

  // 7. Standard JavaScript Error or FetchError.message (e.g. "Failed to fetch")
  if (error.message && typeof error.message === 'string' && error.message.trim()) {
    return error.message.trim()
  }

  return fallback
}

/**
 * Returns structured error information if available.
 */
export const getErrorMessageDetails = (error: any): ErrorDetails => {
  const message = extractErrorMessage(error)
  const details = error?.data?.error?.details || undefined
  const code = error?.data?.error?.code || undefined
  const statusCode = error?.statusCode || error?.status || undefined

  return {
    message,
    details,
    code,
    statusCode
  }
}

// ---------------------------------------------------------------------------
// User-facing (translated) error messages
// ---------------------------------------------------------------------------

/** Translation function with the same signature as `useI18n().t` */
export type TranslateFn = (key: string, params?: Record<string, string | number>) => string

export interface UserErrorMessageOptions {
  /**
   * i18n key used when the situation is unknown (no status, not a network
   * error, no known code), e.g. 'masterData.errors.fetchEmployees'.
   * Defaults to 'errors.unknown'.
   */
  fallbackKey?: string
}

/** The parts of a $fetch / FetchError the helpers below look at */
interface FetchErrorLike {
  statusCode?: unknown
  status?: unknown
  response?: { status?: unknown }
  message?: unknown
  cause?: { message?: unknown }
  data?: {
    error?: unknown
    code?: unknown
    message?: unknown
  }
}

const asErrorLike = (error: unknown): FetchErrorLike =>
  error && typeof error === 'object' ? error as FetchErrorLike : {}

/** HTTP status of a $fetch / FetchError, if the request got a response */
export const getErrorStatus = (error: unknown): number | undefined => {
  const e = asErrorLike(error)
  const status = e.statusCode ?? e.status ?? e.response?.status
  return typeof status === 'number' && status > 0 ? status : undefined
}

/**
 * Machine-readable error code from the backend.
 * Supports the Go shape `{ success: false, error: { code, message, details } }`
 * and a flat `{ code, message }` body.
 */
export const getErrorCode = (error: unknown): string | undefined => {
  const data = asErrorLike(error).data
  const nested = data?.error && typeof data.error === 'object' ? (data.error as { code?: unknown }).code : undefined
  const code = nested ?? data?.code
  return typeof code === 'string' && /^[A-Z][A-Z0-9_]*$/.test(code) ? code : undefined
}

const NETWORK_ERROR_PATTERN = /failed to fetch|networkerror|network request failed|load failed|fetch failed|err_network|err_connection|econnrefused|enotfound|timed? ?out|aborted/i

/**
 * True when the request never got an HTTP response (offline, DNS, CORS,
 * gateway unreachable, timeout).
 */
export const isNetworkError = (error: unknown): boolean => {
  if (!error || typeof error !== 'object' || getErrorStatus(error)) return false
  const e = asErrorLike(error)
  const message = `${String(e.message ?? '')} ${String(e.cause?.message ?? '')}`
  return NETWORK_ERROR_PATTERN.test(message)
}

// Things that betray an implementation detail rather than a sentence meant for a user
const TECHNICAL_MESSAGE_PATTERN = new RegExp([
  String.raw`Key: '`, 'Error:Field validation', String.raw`\bjson:`, 'cannot unmarshal', 'unexpected end of JSON',
  'SQLSTATE', String.raw`\bpq:`, String.raw`\bgorm\b`, 'record not found', 'duplicate key', 'violates',
  'invalid input syntax', 'uuid', 'nil pointer', 'runtime error', 'panic', 'strconv', 'goroutine',
  String.raw`\[(GET|POST|PUT|PATCH|DELETE)\]`, 'FetchError', 'TypeError', 'Failed to construct',
  String.raw`\bat \S+ \(`, '<html', 'Exception', 'undefined', 'null',
  // snake_case identifiers such as business_unit_id
  String.raw`\b[a-z]+_[a-z_]+\b`
].join('|'), 'i')

/**
 * Whether a backend message can be shown to a user as-is: a short,
 * single-line sentence without stack traces, JSON, SQL or field paths.
 */
export const isHumanReadableMessage = (message: unknown): message is string => {
  if (typeof message !== 'string') return false
  const text = message.trim()
  if (!text || text.length > 200 || /[\r\n]/.test(text)) return false
  if (/^[[{]/.test(text)) return false
  return !TECHNICAL_MESSAGE_PATTERN.test(text)
}

const backendMessage = (error: unknown): unknown => {
  const data = asErrorLike(error).data
  if (data?.error && typeof data.error === 'object') return (data.error as { message?: unknown }).message
  if (typeof data?.error === 'string') return data.error
  if (typeof data?.message === 'string') return data.message
  return undefined
}

/** `t` returns the key itself when a translation is missing */
const translateIfExists = (t: TranslateFn, key: string): string | undefined => {
  const value = t(key)
  return value && value !== key ? value : undefined
}

/**
 * Turns any error from $fetch into a readable, translated message.
 * Never returns raw technical text; log the original error with console.error.
 *
 * Priority:
 * 1. Backend error code with a translation at `errors.codes.<CODE>`
 * 2. Situation by HTTP status (network, 401, 403, 404, 400/422, 409, 429, 5xx);
 *    for 400/422 the backend message is used if it is human-readable
 * 3. `options.fallbackKey`, else `errors.unknown`
 */
export const getUserErrorMessage = (error: unknown, t: TranslateFn, options: UserErrorMessageOptions = {}): string => {
  const fallback = () => translateIfExists(t, options.fallbackKey ?? 'errors.unknown') ?? t('errors.unknown')

  if (!error) return fallback()

  const code = getErrorCode(error)
  if (code) {
    const byCode = translateIfExists(t, `errors.codes.${code}`)
    if (byCode) return byCode
  }

  const status = getErrorStatus(error)

  if (!status) {
    return isNetworkError(error) ? t('errors.network') : fallback()
  }

  if (status === 401) return t('errors.sessionExpired')
  if (status === 403) return t('errors.forbidden')
  if (status === 404) return t('errors.notFound')
  if (status === 400 || status === 422) {
    const message = backendMessage(error)
    return isHumanReadableMessage(message) ? message.trim() : t('errors.validation')
  }
  if (status === 409) return t('errors.conflict')
  if (status === 408) return t('errors.network')
  if (status === 429) return t('errors.tooManyRequests')
  if (status >= 500) return t('errors.server')

  return fallback()
}
