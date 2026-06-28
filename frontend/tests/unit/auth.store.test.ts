// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '~/stores/auth'

// Mock $fetch
vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      apiBase: 'http://localhost:3001/api'
    }
  }),
  useCookie: vi.fn(() => ({
    value: null
  })),
  navigateTo: vi.fn()
}))

global.$fetch = vi.fn()

describe('Auth Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('should initialize with default state', () => {
    const store = useAuthStore()

    expect(store.user).toBeNull()
    expect(store.token).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('should login successfully', async () => {
    const store = useAuthStore()
    const mockResponse = {
      user: { id: '1', email: 'test@example.com', fullName: 'Test User' },
      token: 'mock-token'
    }

    vi.mocked($fetch).mockResolvedValueOnce(mockResponse)

    await store.login({ email: 'test@example.com', password: 'password' })

    expect(store.user).toEqual(mockResponse.user)
    expect(store.token).toBe(mockResponse.token)
    expect(store.isAuthenticated).toBe(true)
  })

  it('should logout successfully', async () => {
    const store = useAuthStore()
    store.user = { id: '1', email: 'test@example.com', fullName: 'Test User' } as any
    store.token = 'mock-token'
    store.isAuthenticated = true

    await store.logout()

    expect(store.user).toBeNull()
    expect(store.token).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })
})
