<template>
  <UModal
    v-model:open="store.showModal"
    dismissible
    :ui="{ content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden' }"
    @update:open="(val: boolean) => { if (!val) store.closeModal() }"
  >
    <template #content>
      <UForm
        :state="store.form"
        @submit.prevent="store.handleSubmit"
      >
        <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">
          <!-- Header -->
          <div class="px-4 sm:px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <div class="flex items-center gap-3">
              <UIcon
                name="i-heroicons-briefcase"
                class="text-primary-500 text-2xl"
              />
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ store.isEditing ? t('masterData.jobRole.editTitle') : t('masterData.jobRole.createTitle') }}
              </h3>
            </div>
            <button
              type="button"
              class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors p-1 rounded-lg focus:outline-none"
              :aria-label="t('common.close')"
              @click="store.closeModal"
            >
              <UIcon
                name="i-heroicons-x-mark"
                class="text-2xl"
              />
            </button>
          </div>

          <!-- Body -->
          <div class="p-4 sm:p-6 overflow-y-auto space-y-4">
            <UAlert
              v-if="store.errorMsg"
              color="error"
              variant="soft"
              :title="store.errorMsg"
              icon="i-heroicons-exclamation-circle"
              class="mb-4"
            />

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('masterData.jobRole.sectionInfo') }}
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField
                  :label="t('masterData.jobRole.fieldCode')"
                  size="lg"
                >
                  <UInput
                    v-model="store.form.job_role_code"
                    type="text"
                    :placeholder="t('masterData.jobRole.fieldCodePlaceholder')"
                    :disabled="store.isEditing && !!store.originalCode"
                    required
                  />
                </UFormField>

                <UFormField
                  :label="t('masterData.jobRole.fieldName')"
                  size="lg"
                >
                  <UInput
                    v-model="store.form.job_role_name"
                    type="text"
                    :placeholder="t('masterData.jobRole.fieldNamePlaceholder')"
                    required
                  />
                </UFormField>

                <UFormField
                  :label="t('masterData.jobRole.fieldPositionType')"
                  size="lg"
                  class="md:col-span-2"
                >
                  <USelectMenu
                    :model-value="store.form.job_position_type || undefined"
                    :items="positionTypeItems"
                    value-key="value"
                    :placeholder="t('masterData.jobRole.fieldPositionTypePlaceholder')"
                    class="w-full"
                    @update:model-value="(v: string) => store.onPositionTypeChange(v)"
                  />
                </UFormField>

                <UFormField
                  :label="t('masterData.jobRole.fieldAccessRole')"
                  size="lg"
                  class="md:col-span-2"
                >
                  <USelectMenu
                    :model-value="store.form.default_access_role || undefined"
                    :items="accessRoleItems"
                    value-key="value"
                    :clear="true"
                    :placeholder="t('masterData.jobRole.fieldAccessRolePlaceholder')"
                    class="w-full"
                    @update:model-value="(v: string | undefined) => store.onAccessRoleChange(v)"
                  />
                  <p class="text-xs text-gray-500 mt-1">
                    {{ t('masterData.jobRole.fieldAccessRoleHelp') }}
                  </p>
                  <p
                    v-if="suggestedHint"
                    class="text-xs text-primary-500 mt-1"
                  >
                    {{ suggestedHint }}
                  </p>
                </UFormField>

                <UFormField
                  :label="t('masterData.jobRole.fieldDescription')"
                  size="lg"
                  class="md:col-span-2"
                >
                  <UTextarea
                    v-model="store.form.job_role_description"
                    :placeholder="t('masterData.jobRole.fieldDescriptionPlaceholder')"
                    :rows="3"
                    autoresize
                  />
                </UFormField>
              </div>

              <UFormField
                :label="t('masterData.jobRole.status')"
                size="lg"
              >
                <div class="flex items-center gap-3">
                  <USwitch v-model="store.form.is_active" />
                  <span>{{ store.form.is_active ? t('masterData.jobRole.active') : t('masterData.jobRole.inactive') }}</span>
                </div>
                <p class="text-xs text-gray-500 mt-1">
                  {{ t('masterData.jobRole.statusHelp') }}
                </p>
              </UFormField>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-4 sm:px-6 py-4 bg-[var(--bg-surface)] border-t border-[var(--border-main)] rounded-b-xl flex flex-col-reverse sm:flex-row justify-end gap-3">
            <UButton
              :label="t('common.cancel')"
              color="neutral"
              variant="soft"
              class="w-full sm:w-auto"
              type="button"
              @click="store.closeModal"
            />
            <UButton
              :label="store.isEditing ? t('common.update') : t('common.create')"
              color="primary"
              :loading="store.loading"
              class="w-full sm:w-auto font-bold"
              type="submit"
            />
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useJobRoleStore } from '~/stores/job-role'
import { JOB_POSITION_TYPES, labelOrRaw, suggestAccessRole } from '~/utils/jobRole'

const store = useJobRoleStore()
const { t } = useI18n()

const positionTypeItems = computed(() =>
  JOB_POSITION_TYPES.map(v => ({
    label: labelOrRaw(t, `masterData.jobRole.positionTypes.${v}`, v),
    value: v
  }))
)

// Keep a stored role selectable even if the allowed list no longer contains it.
const accessRoleItems = computed(() => {
  const names = new Set(store.accessRoles)
  if (store.form.default_access_role) names.add(store.form.default_access_role)
  return Array.from(names).map(v => ({
    label: labelOrRaw(t, `masterData.jobRole.accessRoles.${v}`, v),
    value: v
  }))
})

const suggestedHint = computed(() => {
  const suggested = suggestAccessRole(store.form.job_position_type)
  if (!suggested || store.form.default_access_role === suggested) return ''
  return t('masterData.jobRole.suggestedHint', {
    role: labelOrRaw(t, `masterData.jobRole.accessRoles.${suggested}`, suggested)
  })
})

watch(
  () => store.showModal,
  (open) => {
    if (open) store.fetchAccessRoles()
  },
  { immediate: true }
)
</script>
