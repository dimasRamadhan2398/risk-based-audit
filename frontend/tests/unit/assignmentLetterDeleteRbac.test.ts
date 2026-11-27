import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

const showErrorMock = vi.fn()
const showSuccessMock = vi.fn()
const confirmDeleteMock = vi.fn()

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
    showWarning: vi.fn(),
    showInfo: vi.fn(),
    success: showSuccessMock,
    error: showErrorMock
  })
}))

vi.mock('~/stores/global-modal', () => ({
  useGlobalModalStore: () => ({
    confirmDelete: confirmDeleteMock
  })
}))

const TABLE_SFC = resolve(__dirname, '../../components/assignment-letter/AssignmentLetterTable.vue')

describe('Assignment Letter - Delete Button & Action Restricted to ADMIN', () => {
  let store: ReturnType<typeof useAssignmentLetterStore>
  let authStore: ReturnType<typeof useAuthStore>
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useAssignmentLetterStore()
    authStore = useAuthStore()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    confirmDeleteMock.mockReset().mockResolvedValue(true)
    fetchMock = vi.fn().mockResolvedValue({ success: true })
    vi.stubGlobal('$fetch', fetchMock)
    vi.stubGlobal('useGlobalModalStore', () => ({
      confirmDelete: confirmDeleteMock
    }))
  })

  it('AssignmentLetterTable.vue guards the delete button strictly with isAdmin', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('isAdmin')
    expect(sfc).toMatch(/<UTooltip[^>]*v-if="isAdmin"[^>]*:text="t\('assignmentLetter\.formExtra\.delete'\)"/)
    expect(sfc).not.toMatch(/<UTooltip[^>]*v-if="canWrite"[^>]*:text="t\('assignmentLetter\.formExtra\.delete'\)"/)
  })

  it.each([
    UserRole.EXECUTIVE,
    UserRole.AUDITOR,
    UserRole.AUDIT_MANAGER,
    UserRole.CHIEF_AUDIT_EXECUTIVE,
    UserRole.DEPARTMENT_HEAD,
    UserRole.AUDITEE,
    UserRole.VIEWER
  ])('blocks deleteSuratTugas for non-ADMIN role %s', async (role) => {
    authStore.user = { id: 'u1', username: 'user', email: 'u@example.com', role, roles: [role] } as any

    await store.deleteSuratTugas('letter-123')

    expect(confirmDeleteMock).not.toHaveBeenCalled()
    expect(fetchMock).not.toHaveBeenCalled()
    expect(showErrorMock).toHaveBeenCalled()
  })

  it('allows deleteSuratTugas for ADMIN role', async () => {
    authStore.user = { id: 'u-admin', username: 'admin', email: 'admin@example.com', role: UserRole.ADMIN, roles: [UserRole.ADMIN] } as any

    await store.deleteSuratTugas('letter-123')

    expect(confirmDeleteMock).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/assignment-letters/letter-123'),
      expect.objectContaining({ method: 'DELETE' })
    )
    expect(showSuccessMock).toHaveBeenCalled()
  })
})
