import { describe, it, expect } from 'vitest'
import { extractErrorMessage, getErrorMessageDetails } from '../../utils/error'

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
