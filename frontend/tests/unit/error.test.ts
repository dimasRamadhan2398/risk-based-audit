import { describe, it, expect } from 'vitest'
import {
  extractErrorMessage,
  getErrorMessageDetails,
  getUserErrorMessage,
  getErrorCode,
  getErrorStatus,
  isHumanReadableMessage,
  isNetworkError,
  parseBlobErrorBody,
  type TranslateFn
} from '../../utils/error'
import en from '../../locales/en/common.json'
import id from '../../locales/id/common.json'

describe('extractErrorMessage', () => {
  it('returns fallback for null or undefined', () => {
    expect(extractErrorMessage(null, 'Fallback')).toBe('Fallback')
    expect(extractErrorMessage(undefined, 'Fallback')).toBe('Fallback')
  })

  it('returns string errors as is', () => {
    expect(extractErrorMessage('Direct error string')).toBe('Direct error string')
  })

  it('extracts nested Go response error message and details', () => {
    const error = {
      data: {
        success: false,
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Database connection failed',
          details: 'connection refused at port 5432'
        }
      }
    }
    expect(extractErrorMessage(error)).toBe('Database connection failed: connection refused at port 5432')
  })

  it('extracts nested Go response error message without details', () => {
    const error = {
      data: {
        success: false,
        error: {
          code: 'UNAUTHORIZED',
          message: 'Token expired'
        }
      }
    }
    expect(extractErrorMessage(error)).toBe('Token expired')
  })

  it('extracts string error field', () => {
    const error = {
      data: {
        error: 'Invalid request payload'
      }
    }
    expect(extractErrorMessage(error)).toBe('Invalid request payload')
  })

  it('extracts message field from API/gateway', () => {
    const error = {
      data: {
        message: 'An invalid response was received from upstream server'
      }
    }
    expect(extractErrorMessage(error)).toBe('An invalid response was received from upstream server')
  })

  it('extracts msg field', () => {
    const error = {
      data: {
        msg: 'Validation failed'
      }
    }
    expect(extractErrorMessage(error)).toBe('Validation failed')
  })

  it('extracts raw string data', () => {
    const error = {
      data: 'Raw plain text error'
    }
    expect(extractErrorMessage(error)).toBe('Raw plain text error')
  })

  it('extracts statusCode and statusMessage from FetchError', () => {
    const error = {
      statusCode: 504,
      statusMessage: 'Gateway Timeout'
    }
    expect(extractErrorMessage(error)).toBe('504 Gateway Timeout')
  })

  it('extracts Error.message', () => {
    const error = new Error('Failed to fetch')
    expect(extractErrorMessage(error)).toBe('Failed to fetch')
  })

  it('extracts structured error details', () => {
    const error = {
      statusCode: 400,
      data: {
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Title is required',
          details: 'Field title must not be empty'
        }
      }
    }
    const details = getErrorMessageDetails(error)
    expect(details.message).toBe('Title is required: Field title must not be empty')
    expect(details.code).toBe('VALIDATION_ERROR')
    expect(details.details).toBe('Field title must not be empty')
    expect(details.statusCode).toBe(400)
  })
})

// Same lookup rules as useI18n().t: returns the key itself when missing
const makeT = (dict: Record<string, unknown>): TranslateFn => (key, params) => {
  let value: unknown = dict
  for (const k of key.split('.')) {
    if (value && typeof value === 'object' && k in value) value = (value as Record<string, unknown>)[k]
    else return key
  }
  if (typeof value !== 'string') return key
  return params ? value.replace(/\{(\w+)\}/g, (m, p) => params[p]?.toString() ?? m) : value
}

const tEn = makeT(en)
const tId = makeT(id)

const httpError = (status: number, data?: unknown) => ({
  name: 'FetchError',
  message: `[GET] "/api/v1/employees": ${status}`,
  statusCode: status,
  data
})

describe('getUserErrorMessage', () => {
  it('maps a relative-URL TypeError to the context fallback, never the raw text', () => {
    const error = new TypeError('Failed to construct \'URL\': Invalid URL')
    const msg = getUserErrorMessage(error, tEn, { fallbackKey: 'masterData.errors.fetchEmployees' })
    expect(msg).toBe(en.masterData.errors.fetchEmployees)
    expect(msg).not.toMatch(/construct|URL/)
  })

  it('uses errors.unknown when no fallbackKey is given or it is missing', () => {
    expect(getUserErrorMessage(new Error('boom'), tEn)).toBe(en.errors.unknown)
    expect(getUserErrorMessage(null, tEn)).toBe(en.errors.unknown)
    expect(getUserErrorMessage(new Error('boom'), tEn, { fallbackKey: 'does.not.exist' })).toBe(en.errors.unknown)
  })

  it('maps network failures (no response) to errors.network', () => {
    expect(isNetworkError(new TypeError('Failed to fetch'))).toBe(true)
    expect(getUserErrorMessage({ name: 'FetchError', message: '[GET] "/api/v1/employees": <no response> Failed to fetch' }, tEn))
      .toBe(en.errors.network)
    expect(getUserErrorMessage(httpError(408), tEn)).toBe(en.errors.network)
  })

  it.each([
    [401, 'sessionExpired'],
    [403, 'forbidden'],
    [404, 'notFound'],
    [409, 'conflict'],
    [429, 'tooManyRequests'],
    [500, 'server'],
    [502, 'server'],
    [503, 'server'],
    [504, 'server']
  ] as const)('maps HTTP %i to errors.%s', (status, key) => {
    expect(getUserErrorMessage(httpError(status), tEn)).toBe(en.errors[key])
    expect(getUserErrorMessage(httpError(status), tId)).toBe(id.errors[key])
  })

  it('reads the status from `status` and `response.status` too', () => {
    expect(getUserErrorMessage({ status: 403 }, tEn)).toBe(en.errors.forbidden)
    expect(getUserErrorMessage({ response: { status: 500 } }, tEn)).toBe(en.errors.server)
  })

  it('never leaks technical details of a 5xx', () => {
    const error = httpError(500, {
      success: false,
      error: { code: 'DATABASE_ERROR', message: 'Failed to fetch employees', details: 'pq: relation "employees" does not exist' }
    })
    expect(getUserErrorMessage(error, tEn)).toBe(en.errors.server)
  })

  it('shows a human-readable backend message for 400/422', () => {
    const error = httpError(400, { success: false, error: { code: 'VALIDATION_ERROR', message: 'Employee email is required' } })
    expect(getUserErrorMessage(error, tEn)).toBe('Employee email is required')
    expect(getUserErrorMessage(httpError(422, { message: 'Title is required' }), tEn)).toBe('Title is required')
  })

  it('replaces technical 400 messages with errors.validation', () => {
    const binding = httpError(400, {
      success: false,
      error: { code: 'BAD_REQUEST', message: 'Key: \'Employee.Email\' Error:Field validation for \'Email\' failed on the \'email\' tag' }
    })
    expect(getUserErrorMessage(binding, tEn)).toBe(en.errors.validation)
    expect(getUserErrorMessage(httpError(400, { error: 'json: cannot unmarshal string into Go struct field' }), tEn))
      .toBe(en.errors.validation)
    expect(getUserErrorMessage(httpError(400), tId)).toBe(id.errors.validation)
  })

  it('prefers a translated errors.codes.<CODE> over the status message (nested shape)', () => {
    const error = httpError(409, { success: false, error: { code: 'EMPLOYEE_CODE_ALREADY_EXISTS', message: 'Employee code already exists' } })
    expect(getUserErrorMessage(error, tEn)).toBe(en.errors.codes.EMPLOYEE_CODE_ALREADY_EXISTS)
    expect(getUserErrorMessage(error, tId)).toBe(id.errors.codes.EMPLOYEE_CODE_ALREADY_EXISTS)
  })

  it('supports a flat { code, message } body', () => {
    const error = httpError(404, { code: 'EMPLOYEE_NOT_FOUND', message: 'Employee not found' })
    expect(getErrorCode(error)).toBe('EMPLOYEE_NOT_FOUND')
    expect(getUserErrorMessage(error, tId)).toBe(id.errors.codes.EMPLOYEE_NOT_FOUND)
  })

  it('falls back to the status message for unknown codes', () => {
    const error = httpError(404, { success: false, error: { code: 'SOMETHING_NEW', message: 'x' } })
    expect(getUserErrorMessage(error, tEn)).toBe(en.errors.notFound)
    expect(getUserErrorMessage(httpError(500, { error: { code: 'INTERNAL_ERROR' } }), tEn)).toBe(en.errors.server)
  })

  it('ignores codes that are not UPPER_SNAKE identifiers', () => {
    expect(getErrorCode(httpError(500, { code: 'errors.codes.x' }))).toBeUndefined()
    expect(getErrorCode(httpError(500, { code: 42 }))).toBeUndefined()
  })
})

describe('isHumanReadableMessage', () => {
  it('accepts short plain sentences', () => {
    expect(isHumanReadableMessage('Employee code is required')).toBe(true)
  })

  it.each([
    '',
    '   ',
    undefined,
    '{"error":"x"}',
    'line one\nline two',
    'business_unit_id is required',
    'SQLSTATE 23505 duplicate key value violates unique constraint',
    'TypeError: Failed to construct \'URL\': Invalid URL',
    '[GET] "/api/v1/employees": 500 Internal Server Error',
    'x'.repeat(201)
  ])('rejects %j', (message) => {
    expect(isHumanReadableMessage(message)).toBe(false)
  })
})

describe('error i18n keys', () => {
  const collect = (obj: unknown, prefix = ''): string[] =>
    Object.entries(obj as Record<string, unknown>).flatMap(([k, v]) =>
      v && typeof v === 'object' ? collect(v, `${prefix}${k}.`) : [`${prefix}${k}`])

  it.each(['errors', 'masterData'] as const)('%s has identical keys in en and id', (section) => {
    expect(collect(id[section]).sort()).toEqual(collect(en[section]).sort())
  })
})

describe('backend error codes', () => {
  // Codes master-service emits (backend/master-service/pkg/errors/codes.go) that
  // need their own message rather than the status-based one
  const codes = [
    'EMPLOYEE_NOT_FOUND', 'EMPLOYEE_CODE_ALREADY_EXISTS', 'EMPLOYEE_EMAIL_ALREADY_EXISTS',
    'DEPARTMENT_NOT_FOUND', 'DEPARTMENT_CODE_ALREADY_EXISTS', 'DEPARTMENT_NAME_ALREADY_EXISTS',
    'COMPANY_NOT_FOUND', 'COMPANY_CODE_ALREADY_EXISTS', 'COMPANY_TAX_ID_ALREADY_EXISTS',
    'BUSINESS_UNIT_NOT_FOUND', 'BUSINESS_UNIT_COMPANY_MISMATCH', 'PIC_NOT_FOUND', 'PIC_COMPANY_MISMATCH',
    'VALIDATION_FAILED', 'INVALID_REQUEST_BODY', 'INVALID_ID', 'VALUE_TOO_LONG', 'REFERENCE_NOT_FOUND', 'FOREIGN_KEY_VIOLATION',
    'DUPLICATE_ENTRY'
  ]

  it.each(codes)('%s is translated in en and id', (code) => {
    expect(tEn(`errors.codes.${code}`)).not.toBe(`errors.codes.${code}`)
    expect(tId(`errors.codes.${code}`)).not.toBe(`errors.codes.${code}`)
  })

  it('translates a top-level code (new backend shape) on a delete conflict', () => {
    const error = httpError(409, {
      success: false,
      code: 'FOREIGN_KEY_VIOLATION',
      error: { code: 'FOREIGN_KEY_VIOLATION', message: 'This record cannot be deleted because it is still used by other records.' }
    })
    expect(getUserErrorMessage(error, tId)).toBe(id.errors.codes.FOREIGN_KEY_VIOLATION)
  })
})

describe('parseBlobErrorBody', () => {
  // Shape of the ofetch FetchError for `$fetch(url, { responseType: 'blob' })`:
  // the error body is read as a Blob, and over HTTP/2 statusText is empty
  const blobFetchError = (status: number, body: string, type = 'application/json') => ({
    name: 'FetchError',
    message: `[GET] "https://auditsphere.app/api/v1/audit-charters/d8dda52a-3597-4a43-b52b-b36e06a473ce/download": ${status} `,
    statusCode: status,
    statusMessage: '',
    data: new Blob([body], { type })
  })

  const charterFileMissing = () => blobFetchError(404, JSON.stringify({
    success: false,
    error: { code: 'NOT_FOUND', message: 'Audit charter file was not found on the server disk.' }
  }))

  it('a Blob body hides the backend message from extractErrorMessage (the bug)', () => {
    const msg = extractErrorMessage(charterFileMissing(), 'Gagal mengunduh file Audit Charter.')
    expect(msg).toContain('[GET]')
    expect(msg).not.toContain('not found on the server disk')
  })

  it('parses a JSON error body and keeps the status', async () => {
    const parsed = await parseBlobErrorBody(charterFileMissing())
    expect(getErrorStatus(parsed)).toBe(404)
    expect(getErrorCode(parsed)).toBe('NOT_FOUND')
    expect(extractErrorMessage(parsed)).toBe('Audit charter file was not found on the server disk.')
  })

  it('keeps a non-JSON body as text', async () => {
    const parsed = await parseBlobErrorBody(blobFetchError(502, '<html>Bad Gateway</html>', 'text/html'))
    expect(getErrorStatus(parsed)).toBe(502)
    expect((parsed as { data: unknown }).data).toBe('<html>Bad Gateway</html>')
    expect(getUserErrorMessage(parsed, tId)).toBe(id.errors.server)
  })

  it('returns errors without a Blob body unchanged', async () => {
    const plain = httpError(404, { error: { code: 'NOT_FOUND', message: 'Resource not found' } })
    expect(await parseBlobErrorBody(plain)).toBe(plain)
    expect(await parseBlobErrorBody(null)).toBe(null)
    const network = new TypeError('Failed to fetch')
    expect(await parseBlobErrorBody(network)).toBe(network)
  })

  it('has the audit charter download messages in en and id', () => {
    for (const key of ['downloadTitle', 'fileMissing', 'download']) {
      expect(tEn(`auditCharter.errors.${key}`)).not.toBe(`auditCharter.errors.${key}`)
      expect(tId(`auditCharter.errors.${key}`)).not.toBe(`auditCharter.errors.${key}`)
    }
    expect(id.auditCharter.errors.fileMissing).toMatch(/unggah ulang/)
    expect(en.auditCharter.errors.fileMissing).toMatch(/re-upload/)
  })
})
