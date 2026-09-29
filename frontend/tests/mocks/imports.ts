// @ts-nocheck
import { vi } from 'vitest'

export const useToast = () => ({
  add: vi.fn(),
  remove: vi.fn(),
  clear: vi.fn()
})

export const useAppToast = () => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn()
})

export const useRuntimeConfig = () => ({
  public: {
    auditServiceBaseUrl: 'http://localhost:8002/api/v1',
    riskServiceBaseUrl: 'http://localhost:8004/api/v1',
    apiBase: 'http://localhost:8000/api/v1'
  }
})

export default {
  useToast,
  useAppToast,
  useRuntimeConfig
}
