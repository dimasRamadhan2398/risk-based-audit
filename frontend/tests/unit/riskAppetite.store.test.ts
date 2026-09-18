// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useRiskAppetiteStore } from '~/stores/risk-appetite'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      riskServiceBaseUrl: 'http://localhost:8004/api/v1'
    }
  })
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

global.$fetch = vi.fn()

describe('Risk Appetite Store - Status Handling', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('should send SUBMITTED status when creating statement with SUBMITTED', async () => {
    const store = useRiskAppetiteStore()
    const mockPostResponse = {
      id: 'RA-test-1',
      statement: 'Test statement submitted',
      threshold_limit: 10,
      status: 'SUBMITTED'
    }

    vi.mocked($fetch)
      .mockResolvedValueOnce(mockPostResponse) // POST response
      .mockResolvedValueOnce([mockPostResponse]) // fetchStatements GET response

    await store.createStatement({
      statement: 'Test statement submitted',
      threshold_limit: 10,
      status: 'SUBMITTED'
    })

    expect($fetch).toHaveBeenCalledWith(
      'http://localhost:8004/api/v1/risk-appetite',
      expect.objectContaining({
        method: 'POST',
        body: {
          statement: 'Test statement submitted',
          threshold_limit: 10,
          status: 'SUBMITTED'
        }
      })
    )
  })

  it('should send APPROVED status when creating statement with APPROVED', async () => {
    const store = useRiskAppetiteStore()
    const mockPostResponse = {
      id: 'RA-test-2',
      statement: 'Test statement approved',
      threshold_limit: 5,
      status: 'APPROVED'
    }

    vi.mocked($fetch)
      .mockResolvedValueOnce(mockPostResponse)
      .mockResolvedValueOnce([mockPostResponse])

    await store.createStatement({
      statement: 'Test statement approved',
      threshold_limit: 5,
      status: 'APPROVED'
    })

    expect($fetch).toHaveBeenCalledWith(
      'http://localhost:8004/api/v1/risk-appetite',
      expect.objectContaining({
        method: 'POST',
        body: {
          statement: 'Test statement approved',
          threshold_limit: 5,
          status: 'APPROVED'
        }
      })
    )
  })

  it('should default to DRAFT status if status is omitted when creating statement', async () => {
    const store = useRiskAppetiteStore()
    const mockPostResponse = {
      id: 'RA-test-3',
      statement: 'Test statement draft',
      threshold_limit: 8,
      status: 'DRAFT'
    }

    vi.mocked($fetch)
      .mockResolvedValueOnce(mockPostResponse)
      .mockResolvedValueOnce([mockPostResponse])

    await store.createStatement({
      statement: 'Test statement draft',
      threshold_limit: 8
    })

    expect($fetch).toHaveBeenCalledWith(
      'http://localhost:8004/api/v1/risk-appetite',
      expect.objectContaining({
        method: 'POST',
        body: {
          statement: 'Test statement draft',
          threshold_limit: 8,
          status: 'DRAFT'
        }
      })
    )
  })

  it('should properly update statement with new status', async () => {
    const store = useRiskAppetiteStore()
    const mockPutResponse = {
      id: 'RA-test-1',
      statement: 'Updated statement',
      threshold_limit: 12,
      status: 'APPROVED'
    }

    vi.mocked($fetch)
      .mockResolvedValueOnce(mockPutResponse)
      .mockResolvedValueOnce([mockPutResponse])

    await store.updateStatement('RA-test-1', {
      statement: 'Updated statement',
      threshold_limit: 12,
      status: 'APPROVED'
    })

    expect($fetch).toHaveBeenCalledWith(
      'http://localhost:8004/api/v1/risk-appetite/RA-test-1',
      expect.objectContaining({
        method: 'PUT',
        body: {
          statement: 'Updated statement',
          threshold_limit: 12,
          status: 'APPROVED'
        }
      })
    )
  })
})
