import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import {
  useCharterStore,
  getRoleFullName,
  ROLE_DEFAULT_FULL_NAMES
} from '~/stores/charter'
import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))
vi.mock('~/composables/useRbac', () => ({
  useRbac: () => ({ canManageCharter: { value: true } })
}))
vi.mock('~/stores/global-modal', () => ({
  useGlobalModalStore: () => ({ confirmDelete: vi.fn(async () => true) })
}))

describe('Audit Charter Uploaded By Sync with Account Settings Full Name for each role', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('getRoleFullName resolution helper', () => {
    it('uses the custom Full Name when set in Account Settings', () => {
      expect(getRoleFullName({ fullName: 'Budi Santoso', roles: ['admin'] })).toBe('Budi Santoso')
      expect(getRoleFullName({ full_name: 'Dewi Lestari', roles: ['auditor'] })).toBe('Dewi Lestari')
    })

    it('resolves correct default Full Name for each role matching Account Settings seeds', () => {
      expect(getRoleFullName({ roles: [UserRole.ADMIN] })).toBe('System Administrator')
      expect(getRoleFullName({ roles: ['admin'] })).toBe('System Administrator')
      expect(getRoleFullName({ roles: [UserRole.AUDITOR] })).toBe('John Auditor')
      expect(getRoleFullName({ roles: ['auditor'] })).toBe('John Auditor')
      expect(getRoleFullName({ roles: [UserRole.AUDIT_MANAGER] })).toBe('Audit Manager')
      expect(getRoleFullName({ roles: ['audit_manager'] })).toBe('Audit Manager')
      expect(getRoleFullName({ roles: [UserRole.CHIEF_AUDIT_EXECUTIVE] })).toBe('Chief Audit Executive')
      expect(getRoleFullName({ roles: ['cae'] })).toBe('Chief Audit Executive')
      expect(getRoleFullName({ roles: [UserRole.DEPARTMENT_HEAD] })).toBe('Alice Department Head')
      expect(getRoleFullName({ roles: ['department_head'] })).toBe('Alice Department Head')
      expect(getRoleFullName({ roles: [UserRole.AUDITEE] })).toBe('Bob Auditee')
      expect(getRoleFullName({ roles: ['auditee'] })).toBe('Bob Auditee')
      expect(getRoleFullName({ roles: [UserRole.VIEWER] })).toBe('Carol Viewer')
      expect(getRoleFullName({ roles: ['viewer'] })).toBe('Carol Viewer')
      expect(getRoleFullName({ roles: [UserRole.AUDIT_STAFF] })).toBe('Audit Staff')
      expect(getRoleFullName({ roles: ['audit_staff'] })).toBe('Audit Staff')
    })

    it('falls back to username or System Administrator if roles are empty', () => {
      expect(getRoleFullName({ username: 'custom_user' })).toBe('custom_user')
      expect(getRoleFullName(null)).toBe('System Administrator')
      expect(getRoleFullName(undefined)).toBe('System Administrator')
    })
  })

  describe('charter store uploadedBy synchronization', () => {
    it('initializes form.uploadedBy with the logged-in role Full Name', () => {
      const auth = useAuthStore()
      auth.user = {
        id: '1',
        username: 'admin',
        email: 'admin@company.com',
        fullName: 'System Administrator',
        roles: ['admin']
      }

      const store = useCharterStore()
      expect(store.currentUserFullName).toBe('System Administrator')
      expect(store.form.uploadedBy).toBe('System Administrator')
    })

    it('initializes form.uploadedBy with the Auditor Full Name when logged in as auditor', () => {
      const auth = useAuthStore()
      auth.user = {
        id: '2',
        username: 'auditor',
        email: 'auditor@company.com',
        fullName: 'John Auditor',
        roles: ['auditor']
      }

      const store = useCharterStore()
      expect(store.currentUserFullName).toBe('John Auditor')
      expect(store.form.uploadedBy).toBe('John Auditor')
    })

    it('dynamically updates form.uploadedBy when user changes Full Name in Account Settings', async () => {
      const auth = useAuthStore()
      auth.user = {
        id: '1',
        username: 'admin',
        email: 'admin@company.com',
        fullName: 'System Administrator',
        roles: ['admin']
      }

      const store = useCharterStore()
      expect(store.form.uploadedBy).toBe('System Administrator')

      // User changes profile in Account Settings
      auth.user.fullName = 'Dimas Administrator'
      expect(store.currentUserFullName).toBe('Dimas Administrator')
    })

    it('maps legacy "Dimas (HIA)" or empty uploaded_by in fetched charters to current role Full Name', async () => {
      const auth = useAuthStore()
      auth.user = {
        id: '1',
        username: 'admin',
        email: 'admin@company.com',
        fullName: 'System Administrator',
        roles: ['admin']
      }

      const store = useCharterStore()

      // Fake backend response
      const g = globalThis as any
      g.$fetch = vi.fn().mockResolvedValue({
        data: {
          charters: [
            { id: 'c1', title: 'Charter Active', version: '2.0', is_active: true, uploaded_by: 'Dimas (HIA)' },
            { id: 'c2', title: 'Charter Hist 1', version: '1.9', is_active: false, uploaded_by: '' },
            { id: 'c3', title: 'Charter Custom', version: '1.8', is_active: false, uploaded_by: 'Pak Direktur' }
          ],
          pagination: { page: 1, page_size: 10, total_count: 3, total_pages: 1 }
        }
      })

      await store.fetchCharters()

      expect(store.charters[0]?.uploadedBy).toBe('System Administrator')
      expect(store.charters[1]?.uploadedBy).toBe('System Administrator')
      expect(store.charters[2]?.uploadedBy).toBe('Pak Direktur')
    })

    it('resets form.uploadedBy to currentUserFullName on closeModal()', () => {
      const auth = useAuthStore()
      auth.user = {
        id: '3',
        username: 'manager',
        email: 'manager@company.com',
        fullName: 'Audit Manager',
        roles: ['audit_manager']
      }

      const store = useCharterStore()
      store.form.uploadedBy = 'Some Random Value'
      store.closeModal()

      expect(store.form.uploadedBy).toBe('Audit Manager')
    })

    it('replaces legacy "Dimas (HIA)" with currentUserFullName on handleEdit()', () => {
      const auth = useAuthStore()
      auth.user = {
        id: '1',
        username: 'admin',
        email: 'admin@company.com',
        fullName: 'System Administrator',
        roles: ['admin']
      }

      const store = useCharterStore()
      store.handleEdit({
        id: 'c1',
        title: 'Old Charter',
        version: '1.0',
        content: 'Content',
        date: '2026-01-01',
        uploadedBy: 'Dimas (HIA)',
        approvedBy: 'Komite Audit',
        isActive: true
      })

      expect(store.form.uploadedBy).toBe('System Administrator')
    })
  })
})
