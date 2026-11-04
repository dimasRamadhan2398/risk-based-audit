<template>
  <UModal
    :open="!!account"
    :dismissible="false"
    class="w-full sm:max-w-md"
  >
    <template #content>
      <div
        v-if="account && account.temporary_password"
        class="bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col border border-[var(--border-main)]"
      >
        <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex items-center gap-3">
          <UIcon
            name="i-heroicons-key"
            class="text-primary-500 text-2xl"
          />
          <div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ t('masterData.employee.account.credentialsTitle') }}
            </h3>
            <p class="text-sm text-[var(--text-muted)]">
              {{ t('masterData.employee.account.credentialsDescription', { name: store.createdAccountEmployeeName }) }}
            </p>
          </div>
        </div>
        <div class="px-6 py-5">
          <TemporaryPasswordPanel
            :username="account.username"
            :temporary-password="account.temporary_password"
          />
        </div>
        <div class="px-6 py-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] rounded-b-xl flex justify-end">
          <UButton
            class="rounded-xl font-bold"
            @click="store.clearCreatedAccount"
          >
            {{ t('masterData.employee.account.done') }}
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useEmployeeStore } from '~/stores/employee'
import { useI18n } from '~/composables/useI18n'
import TemporaryPasswordPanel from '~/components/master/TemporaryPasswordPanel.vue'

// Shows the one-time temporary password after an employee is created
const store = useEmployeeStore()
const { t } = useI18n()
const account = computed(() => store.createdAccount)

// Never keep the password around after leaving the page
onBeforeUnmount(() => store.clearCreatedAccount())
</script>
