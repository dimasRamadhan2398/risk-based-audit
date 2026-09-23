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
