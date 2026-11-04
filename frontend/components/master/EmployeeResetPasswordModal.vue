<template>
  <UModal
    v-model:open="open"
    :dismissible="step !== 'loading' && step !== 'resetting' && step !== 'creating'"
    class="w-full sm:max-w-md"
  >
    <template #content>
      <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col border border-[var(--border-main)] transition-colors duration-300">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex items-center gap-3">
          <UIcon
            name="i-heroicons-key"
            class="text-primary-500 text-2xl"
          />
          <div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ t('masterData.employee.resetPassword') }}
            </h3>
            <p
              v-if="employee"
              class="text-sm text-[var(--text-muted)]"
            >
              {{ employee.full_name }} · {{ employee.employee_code }}
            </p>
          </div>
        </div>

        <div class="px-6 py-5 space-y-4">
          <!-- Looking up the linked account -->
          <div
            v-if="step === 'loading'"
            class="flex justify-center py-6"
          >
            <ULoadingIcon />
          </div>

          <!-- Employee has no login account: offer to create one -->
          <template v-else-if="step === 'no-account' || step === 'creating'">
            <UAlert
              color="neutral"
              variant="soft"
              icon="i-heroicons-information-circle"
              :title="t('masterData.employee.resetModal.noAccount')"
              :description="t('masterData.employee.account.createAccountDescription')"
            />
            <UFormField :label="t('masterData.employee.account.accessRole')">
              <USelect
                v-model="selectedRole"
                :items="roleItems"
                :disabled="step === 'creating'"
                class="w-full"
              />
            </UFormField>
          </template>

          <!-- Account created, linked or already existing -->
          <template v-else-if="step === 'created' && createResult">
            <template v-if="createResult.status === 'created' && createResult.temporary_password">
              <UAlert
                color="success"
                variant="soft"
                icon="i-heroicons-check-circle"
                :title="t('masterData.employee.account.createdSuccess', { username: createResult.username })"
              />
              <TemporaryPasswordPanel
                :username="createResult.username"
                :temporary-password="createResult.temporary_password"
              />
            </template>
            <UAlert
              v-else
              color="info"
              variant="soft"
              icon="i-heroicons-information-circle"
              :description="t(createResult.status === 'exists' ? 'masterData.employee.account.existsInfo' : 'masterData.employee.account.linkedInfo', { username: createResult.username })"
            />
          </template>

          <!-- Account cannot be reset from here -->
          <UAlert
            v-else-if="step === 'blocked'"
            color="warning"
            variant="soft"
            icon="i-heroicons-shield-exclamation"
            :title="t('masterData.employee.resetModal.cannotReset')"
            :description="blockedReason"
          />

          <!-- Confirm -->
          <template v-else-if="(step === 'confirm' || step === 'resetting') && account">
            <div class="rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 text-sm space-y-1">
              <div class="flex justify-between">
                <span class="text-[var(--text-muted)]">{{ t('settings.profile.username') }}</span><span class="font-medium text-[var(--text-main)]">{{ account.username }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-[var(--text-muted)]">{{ t('masterData.employee.resetModal.email') }}</span><span class="font-medium text-[var(--text-main)]">{{ account.email }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-[var(--text-muted)]">{{ t('masterData.employee.resetModal.roles') }}</span><span class="font-medium text-[var(--text-main)]">{{ account.roles.join(', ') || '-' }}</span>
              </div>
            </div>
            <p class="text-sm text-[var(--text-muted)]">
              {{ t('masterData.employee.resetModal.temporaryDescription') }}
            </p>
          </template>

          <!-- Result: temporary password shown once -->
          <template v-else-if="step === 'done' && result">
            <UAlert
              color="success"
              variant="soft"
              icon="i-heroicons-check-circle"
              :title="`Password for ${result.username} has been reset`"
              :description="t('masterData.employee.resetModal.shareSecurely')"
            />
            <div class="flex items-center gap-2">
              <code class="flex-1 rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 font-mono text-lg tracking-wider text-[var(--text-main)] select-all">{{ result.temporary_password }}</code>
              <UButton
                :icon="copied ? 'i-heroicons-check' : 'i-heroicons-clipboard-document'"
                color="neutral"
                variant="subtle"
                size="lg"
                @click="copyPassword"
              />
            </div>
          </template>

          <UAlert
            v-if="errorMsg"
            color="error"
            variant="soft"
            icon="i-heroicons-exclamation-circle"
            :title="errorMsg"
          />
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] rounded-b-xl flex justify-end gap-2">
          <UButton
            color="neutral"
            variant="subtle"
            class="rounded-xl font-semibold"
            :disabled="step === 'resetting'"
            @click="open = false"
          >
            {{ step === 'done' || step === 'created' ? t('masterData.employee.account.done') : 'Cancel' }}
          </UButton>
          <UButton
            v-if="step === 'no-account' || step === 'creating'"
            class="rounded-xl font-bold"
            :loading="step === 'creating'"
            @click="handleCreateAccount"
          >
            {{ t('masterData.employee.account.createAccount') }}
          </UButton>
          <UButton
            v-if="step === 'created'"
            color="neutral"
            class="rounded-xl font-bold"
            @click="continueToReset"
          >
            {{ t('masterData.employee.account.continueReset') }}
          </UButton>
          <UButton
            v-if="step === 'confirm' || step === 'resetting'"
            color="error"
            class="rounded-xl font-bold"
            :loading="step === 'resetting'"
            @click="handleReset"
          >
            {{ t('masterData.employee.resetPassword') }}
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import type { Employee, EmployeeLoginAccount } from '~/types/master'
import type { AdminResetPasswordResult, LinkedUserAccount } from '~/composables/useUserAccountApi'
import { useUserAccountApi } from '~/composables/useUserAccountApi'
import { useEmployeeApi } from '~/composables/useEmployeeApi'
import { useJobRoleApi } from '~/composables/useJobRoleApi'
import { useAuthStore } from '~/stores/auth'
import { getUserErrorMessage } from '~/utils/error'
import TemporaryPasswordPanel from '~/components/master/TemporaryPasswordPanel.vue'

const { t } = useI18n()

type Step = 'loading' | 'no-account' | 'creating' | 'created' | 'blocked' | 'confirm' | 'resetting' | 'done'

const open = defineModel<boolean>('open', { default: false })
const props = defineProps<{ employee: Employee | null }>()

const authStore = useAuthStore()
const { findAccountByEmployee, resetPassword } = useUserAccountApi()
const { createLoginAccount } = useEmployeeApi()
const { getAccessRoles } = useJobRoleApi()

// Sentinel meaning "let the backend pick the role" (select items cannot be empty)
const DEFAULT_ROLE = '__default__'

const step = ref<Step>('loading')
const account = ref<LinkedUserAccount | null>(null)
const result = ref<AdminResetPasswordResult | null>(null)
const blockedReason = ref('')
const errorMsg = ref('')
const copied = ref(false)
const createResult = ref<EmployeeLoginAccount | null>(null)
const selectedRole = ref(DEFAULT_ROLE)
const accessRoles = ref<string[]>([])
const roleItems = computed(() => [
  { label: t('masterData.employee.account.accessRoleDefault'), value: DEFAULT_ROLE },
  ...accessRoles.value.map(r => ({ label: r, value: r }))
])

const loadAccessRoles = async () => {
  if (accessRoles.value.length) return
  // ADMIN cannot be granted from here
  accessRoles.value = (await getAccessRoles()).filter(r => r.toUpperCase() !== 'ADMIN')
}

const lookupAccount = async (employee: Employee) => {
  step.value = 'loading'
  account.value = null
  result.value = null
  createResult.value = null
  selectedRole.value = DEFAULT_ROLE
  errorMsg.value = ''
  copied.value = false

  try {
    const found = await findAccountByEmployee(employee.employee_code, employee.email)
    if (!found) {
      step.value = 'no-account'
      loadAccessRoles()
      return
    }
    account.value = found

    // Mirrors the backend rules so the admin sees why before trying
    if (found.id === authStore.user?.id) {
      blockedReason.value = 'This is your own account. Change your password from Settings instead.'
      step.value = 'blocked'
    } else if (found.roles.some(r => r.toUpperCase() === 'ADMIN')) {
      blockedReason.value = 'Admin passwords can only be reset by the AuditSphere platform administrator.'
      step.value = 'blocked'
    } else {
      step.value = 'confirm'
    }
  } catch (err: any) {
    errorMsg.value = err.message
    step.value = 'blocked'
    blockedReason.value = 'The account could not be loaded.'
  }
}

const handleCreateAccount = async () => {
  if (!props.employee) return
  step.value = 'creating'
  errorMsg.value = ''
  try {
    createResult.value = await createLoginAccount(
      props.employee.id,
      selectedRole.value === DEFAULT_ROLE ? undefined : selectedRole.value
    )
    step.value = 'created'
  } catch (err) {
    errorMsg.value = getUserErrorMessage(err, t, { fallbackKey: 'masterData.errors.createLoginAccount' })
    step.value = 'no-account'
  }
}

const continueToReset = () => {
  if (props.employee) lookupAccount(props.employee)
}

const handleReset = async () => {
  if (!account.value) return
  step.value = 'resetting'
  errorMsg.value = ''
  try {
    result.value = await resetPassword(account.value.id)
    step.value = 'done'
  } catch (err: any) {
    errorMsg.value = err.message
    step.value = 'confirm'
  }
}

const copyPassword = async () => {
  if (!result.value) return
  await navigator.clipboard.writeText(result.value.temporary_password)
  copied.value = true
}

watch(open, (isOpen) => {
  if (isOpen && props.employee) {
    lookupAccount(props.employee)
  } else if (!isOpen) {
    // Drop the temporary password from memory once the modal closes
    result.value = null
    createResult.value = null
  }
})
</script>
