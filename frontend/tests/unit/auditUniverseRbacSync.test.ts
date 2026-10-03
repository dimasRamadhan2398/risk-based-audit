// @ts-nocheck
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import fs from 'fs'
import path from 'path'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { UserRole } from '~/types/auth'

// Mirrors the mocking pattern used in tests/unit/auth.store.test.ts
global.$fetch = vi.fn()

describe('Audit Universe RBAC sync after login (bug repro)', () => {
  const originalUseCookie = global.useCookie

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    global.useCookie = originalUseCookie
  })

  afterEach(() => {
    global.useCookie = originalUseCookie
  })

  it('grants canEditAuditUniverse to an AUDIT_MANAGER immediately after a successful login', async () => {
    const store = useAuthStore()

    // Mirrors the real auth-service response shape consumed by _persistSession
    // (backend/auth-service/models/dto.go UserInfo.Roles is []string, populated from
    // role.Name — confirmed via backend/auth-service/services/auth/auth_service.go
    // completeLogin()). Response is wrapped in {success, message, data}.
    vi.mocked($fetch).mockResolvedValueOnce({
      data: {
        user: {
          id: 'u-1',
          email: 'manager@auditsphere.app',
          full_name: 'Manager User',
          roles: [UserRole.AUDIT_MANAGER],
        },
        token: 'real-token',
      },
    })

    await store.login({ username: 'manager', password: 'secret' })

    expect(store.isAuthenticated).toBe(true)
    expect(store.user?.roles).toEqual([UserRole.AUDIT_MANAGER])

    const rbac = useRbac()
    expect(rbac.hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE])).toBe(true)
    expect(rbac.canEditAuditUniverse.value).toBe(true)

    // Mirrors the actual gates in pages/risk-profile/audit-universe/index.vue:
    //   line 72/91: :disabled="!canEditAuditUniverse"  (row checkboxes)
    //   line 115/137: v-if="canEditAuditUniverse"       (Edit / Add UI)
    const checkboxDisabled = !rbac.canEditAuditUniverse.value
    const editUiVisible = rbac.canEditAuditUniverse.value
    expect(checkboxDisabled).toBe(false)
    expect(editUiVisible).toBe(true)
  })

  it('BUG: fetchUser() cookie-rehydration leaves a logged-in manager with no edit access when roles were dropped from the persisted user', async () => {
    const store = useAuthStore()

    // This is the exact shape stores/auth.ts _persistSession() would have written to the
    // `auth-user` cookie if the login response's data.user.roles was undefined: the
    // conditional spread at stores/auth.ts:254
    //   ...(data.user?.roles !== undefined ? { roles: data.user.roles } : {})
    // simply omits `roles` from the persisted user object instead of defaulting to [].
    // fetchUser() (stores/auth.ts:306-334) then parses this cookie verbatim on reload/
    // hydration and sets isAuthenticated = true with this roles-less user.
    const persistedUserWithoutRoles = {
      id: 'u-1',
      email: 'manager@auditsphere.app',
      fullName: 'Manager User',
      mustChangePassword: false,
    }

    global.useCookie = vi.fn((name: string) => {
      if (name === 'auth-token') return { value: 'real-token' }
      if (name === 'auth-user') return { value: JSON.stringify(persistedUserWithoutRoles) }
      return { value: null }
    })

    await store.fetchUser()

    // The store's own bookkeeping says the user IS logged in.
    expect(store.isAuthenticated).toBe(true)
    expect(store.user).toBeTruthy()
    expect(store.user?.email).toBe('manager@auditsphere.app')

    const rbac = useRbac()

    // THE BUG: despite isAuthenticated === true, canEditAuditUniverse stays false because
    // `roles` never made it into the rehydrated user object. The Audit Universe page then
    // renders as if the user were unauthenticated: Edit UI hidden, row checkboxes disabled.
    expect(rbac.hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE])).toBe(false)
    expect(rbac.canEditAuditUniverse.value).toBe(false)
  })

  it('hasAnyRole normalizes case, ruling out role-string casing as the cause when roles ARE present', () => {
    const store = useAuthStore()

    // Seed data in this repo is inconsistent in casing: backend/auth-service seeder.go
    // SeedRoles() creates UPPERCASE_SNAKE names ("ADMIN"), while seed_admin.sql inserts
    // lowercase ("admin"). hasAnyRole()/normalize() must tolerate either.
    store.user = {
      id: 'u-2',
      email: 'admin@auditsphere.app',
      fullName: 'Admin User',
      roles: ['ADMIN'],
    } as any
    store.isAuthenticated = true

    const rbac = useRbac()
    expect(rbac.hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE])).toBe(true)
    expect(rbac.canEditAuditUniverse.value).toBe(true)
  })

  it('the audit-universe page gates row checkboxes and Edit UI on canEditAuditUniverse', () => {
    const pagePath = path.resolve(__dirname, '../../pages/risk-profile/audit-universe/index.vue')
    const content = fs.readFileSync(pagePath, 'utf8')

    expect(content).toContain(':disabled="!canEditAuditUniverse"')
    expect(content).toContain('v-if="canEditAuditUniverse"')
  })
})
