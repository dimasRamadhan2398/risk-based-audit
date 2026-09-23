// @ts-nocheck
import { vi } from 'vitest'

// Mock Nuxt auto-imports
global.defineNuxtRouteMiddleware = vi.fn((middleware) => middleware)
global.definePageMeta = vi.fn()
global.navigateTo = vi.fn()
global.useRouter = vi.fn(() => ({
  push: vi.fn(),
  replace: vi.fn(),
  go: vi.fn(),
  back: vi.fn()
}))
global.useRoute = vi.fn(() => ({
  query: {},
  params: {},
  path: '/',
  fullPath: '/'
}))
global.useCookie = vi.fn(() => ({
  value: null
}))
global.useRuntimeConfig = vi.fn(() => ({
  public: {
    apiBase: 'http://localhost:8080/api/v1'
  }
}))
global.ref = (val: any) => ({ value: val })
global.computed = (fn: any) => ({ value: fn() })

export {}
