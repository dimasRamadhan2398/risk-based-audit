<template>
  <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xs">
    <template #header>
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.permission.title') }}</h3>
          <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400">
            {{ t('settings.permission.subtitle') }}
          </p>
        </div>
      </div>
    </template>

    <!-- Role Selector -->
    <div class="mb-6">
      <h4 class="font-medium text-gray-900 dark:text-white mb-3">
        {{ t('settings.permission.currentRole') }} <span class="text-primary-600 dark:text-primary-400 font-bold">{{ currentRole }}</span>
      </h4>
    </div>

    <!-- Permissions Accordion -->
    <UAccordion
      :items="accordionItems"
      variant="subtle"
      color="primary"
      :ui="{
        trigger: 'text-gray-900 dark:text-gray-200 hover:text-gray-700 dark:hover:text-gray-400 transition-colors',
        label: 'font-bold text-gray-900 dark:text-gray-100 hover:text-gray-700 dark:hover:text-gray-400'
      }"
    >
      <template #content="{ item }">
        <div class="pb-4">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-800">
                <th class="text-left py-3 px-4 font-medium text-sm text-gray-700 dark:text-gray-300">{{ t('settings.permission.tableModule') }}</th>
                <th class="text-left py-3 px-4 font-medium text-sm text-gray-700 dark:text-gray-300">{{ t('settings.permission.tableDescription') }}</th>
                <th class="text-center py-3 px-4 font-medium text-sm text-gray-700 dark:text-gray-300">{{ t('settings.permission.tableRead') }}</th>
                <th class="text-center py-3 px-4 font-medium text-sm text-gray-700 dark:text-gray-300">{{ t('settings.permission.tableWrite') }}</th>
                <th class="text-center py-3 px-4 font-medium text-sm text-gray-700 dark:text-gray-300">{{ t('settings.permission.tableDelete') }}</th>
                <th class="text-center py-3 px-4 font-medium text-sm text-gray-700 dark:text-gray-300">{{ t('settings.permission.tableAction') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="permission in item.permissions"
                :key="permission.id"
                class="border-b border-gray-100 dark:border-gray-800/60 hover:bg-gray-50 dark:hover:bg-gray-800/40 group transition-colors"
              >
                <td class="py-3 px-4">
                  <div class="flex items-center gap-3">
                    <span class="font-medium text-gray-900 dark:text-gray-200 group-hover:text-gray-700 dark:group-hover:text-gray-400 transition-colors">
                      {{ permission.module }}
                    </span>
                  </div>
                </td>
                <td class="py-3 px-4 text-gray-600 dark:text-gray-400 group-hover:text-gray-700 dark:group-hover:text-gray-400 transition-colors">
                  {{ permission.description }}
                </td>
                <td class="py-3 px-4 text-center">
                  <UCheckbox :modelValue="getAccessCheckbomdtate(permission.access, 'read')" disabled />
                </td>
                <td class="py-3 px-4 text-center">
                  <UCheckbox :modelValue="getAccessCheckbomdtate(permission.access, 'write')" disabled />
                </td>
                <td class="py-3 px-4 text-center">
                  <UCheckbox :modelValue="getAccessCheckbomdtate(permission.access, 'delete')" disabled />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </UAccordion>

    <!-- Add/Edit Modal -->
    <UModal
      :title="isEditMode ? t('settings.permission.editModalTitle') : t('settings.permission.addModalTitle')"
      :open="isModalOpen"
      :close="{
        color: 'neutral',
        variant: 'outline',
        onClick: closeModal,
      }"
      variant="default"
      dismissible
    >
      <template #body>
        <UForm @submit.prevent="handleSubmit">
          <div class="space-y-4">
            <UFormField
              :label="t('settings.permission.labelModule')"
              required
            >
              <UInput
                v-model="formData.module"
                :placeholder="t('settings.permission.placeholderModule')"
              />
            </UFormField>

            <UFormField
              :label="t('settings.permission.labelIcon')"
            >
              <UInput
                v-model="formData.icon"
                placeholder="i-lucide-shield"
              />
            </UFormField>

            <UFormField
              :label="t('settings.permission.labelDescription')"
            >
              <UTextarea
                v-model="formData.description"
                :placeholder="t('settings.permission.placeholderDescription')"
                :rows="2"
              />
            </UFormField>

            <UFormField
              :label="t('settings.permission.labelAccess')"
            >
              <USelect
                v-model="formData.access"
                :items="accessLevels"
                :placeholder="t('settings.permission.placeholderAccess')"
              />
            </UFormField>
          </div>
        </UForm>
      </template>

      <template #footer>
        <div class="flex justify-end gap-3">
          <UButton
            :label="t('settings.permission.cancel')"
            color="neutral"
            variant="outline"
            @click="closeModal"
          />
          <UButton
            :label="isEditMode ? t('settings.permission.save') : t('settings.permission.add')"
            color="primary"
            @click="handleSubmit"
          />
        </div>
      </template>
    </UModal>

    <!-- Delete Confirmation -->
    <ConfirmationPopup
      v-model:isOpen="isDeleteModalOpen"
      :title="t('settings.permission.deleteTitle')"
      :question="t('settings.permission.deleteQuestion', { module: permissionToDelete?.module || '' })"
      :confirmText="t('settings.permission.confirmDelete')"
      :cancelText="t('settings.permission.cancel')"
      variant="danger"
      @confirm="confirmDelete"
    />
  </UCard>
</template>

<script setup lang="ts">
import ConfirmationPopup from '../shared/ConfirmationPopup.vue'

const { t } = useI18n()

interface AuditPermission {
  id: number
  module: string
  icon: string
  description: string
  access: 'None' | 'Read' | 'Write' | 'Full' | ''
  category: string
}

interface AccordionItem {
  label: string
  permissions: AuditPermission[]
}

const currentRole = ref('Auditor')
const isModalOpen = ref(false)
const isDeleteModalOpen = ref(false)
const isEditMode = ref(false)
const permissionToDelete = ref<AuditPermission | null>(null)
const editingPermissionId = ref<number | null>(null)

const formData = ref({
  module: '',
  icon: 'i-lucide-shield',
  description: '',
  access: '' as 'None' | 'Read' | 'Write' | 'Full' | ''
})

const accessLevels = computed(() => [
  { value: 'None', label: t('settings.permission.accessNone') },
  { value: 'Read', label: t('settings.permission.accessRead') },
  { value: 'Write', label: t('settings.permission.accessWrite') },
  { value: 'Full', label: t('settings.permission.accessFull') },
])

const auditPermissions = ref<AuditPermission[]>([
  {
    id: 1,
    module: 'Audit Charter',
    icon: 'i-lucide-shield',
    description: 'Manage audit charter',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 2,
    module: 'Strategic Plan Internal Audit',
    icon: 'i-lucide-shield',
    description: 'Manage Strategic Plan Internal Audit',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 3,
    module: 'Annual Audit Plan',
    icon: 'i-lucide-shield',
    description: 'Manage audit planning and execution',
    access: 'Read',
    category: 'Internal Audit'
  },
  {
    id: 4,
    module: 'Risk Assessment',
    icon: 'i-lucide-bar-chart-3',
    description: 'Manage risk assessment',
    access: 'Write',
    category: 'Risk Management'
  },
  {
    id: 5,
    module: 'Risk Profile',
    icon: 'i-lucide-file-text',
    description: 'Manage risk profile',
    access: 'Full',
    category: 'Risk Management'
  },
  {
    id: 6,
    module: 'Kertas Kerja Audit',
    icon: 'i-lucide-file-text',
    description: 'Manage Audit Working Paper',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 7,
    module: 'Surat Tugas Audit',
    icon: 'i-lucide-file-text',
    description: 'Manage Audit Assignment Letter',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 8,
    module: 'Audit Fieldwork',
    icon: 'i-lucide-file-text',
    description: 'Manage Audit Fieldwork',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 9,
    module: 'Data Analytics',
    icon: 'i-lucide-file-text',
    description: 'Manage Audit Assignment Letter',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 10,
    module: 'Monitoring ATR',
    icon: 'i-lucide-file-text',
    description: 'Monitoring Action Taken Reports',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 11,
    module: 'Quality Assurance Review',
    icon: 'i-lucide-file-text',
    description: 'Monitoring Action Taken Reports',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 12,
    module: 'Internal Audit Performance',
    icon: 'i-lucide-file-text',
    description: 'KPI Achievement & Work Plan Realization',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 13,
    module: 'Risk Report Update',
    icon: 'i-lucide-file-text',
    description: 'Manage Risk Report Update',
    access: 'Full',
    category: 'Internal Audit'
  },
  {
    id: 14,
    module: 'Reporting',
    icon: 'i-lucide-file-text',
    description: 'Audit Findings, Draft & Final Report',
    access: 'Full',
    category: 'Internal Audit'
  },
])

const accordionItems = computed<AccordionItem[]>(() => {
  const categories = ['Internal Audit', 'Risk Management']
  return categories.map(category => ({
    label: category,
    permissions: auditPermissions.value.filter(p => p.category === category)
  }))
})

function getAccessCheckbomdtate(access: string, checkType: 'read' | 'write' | 'delete'): boolean {
  switch (checkType) {
    case 'read':
      return access === 'Read' || access === 'Write' || access === 'Full'
    case 'write':
      return access === 'Write' || access === 'Full'
    case 'delete':
      return access === 'Full'
    default:
      return false
  }
}

function openAddModal() {
  isEditMode.value = false
  editingPermissionId.value = null
  formData.value = {
    module: '',
    icon: 'i-lucide-shield',
    description: '',
    access: ''
  }
  isModalOpen.value = true
}

function openEditModal(permission: AuditPermission) {
  isEditMode.value = true
  editingPermissionId.value = permission.id
  formData.value = {
    module: permission.module,
    icon: permission.icon,
    description: permission.description,
    access: permission.access
  }
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  formData.value = {
    module: '',
    icon: 'i-lucide-shield',
    description: '',
    access: ''
  }
}

function openDeleteModal(permission: AuditPermission) {
  permissionToDelete.value = permission
  isDeleteModalOpen.value = true
}

function handleSubmit() {
  if (!formData.value.module) {
    return
  }

  if (isEditMode.value && editingPermissionId.value) {
    const index = auditPermissions.value.findIndex(p => p.id === editingPermissionId.value)
    if (index !== -1) {
      const existingPermission = auditPermissions.value[index]
      if (existingPermission) {
        const updatedPermission: AuditPermission = {
          id: editingPermissionId.value,
          module: formData.value.module,
          icon: formData.value.icon,
          description: formData.value.description,
          access: formData.value.access,
          category: existingPermission.category
        }
        auditPermissions.value[index] = updatedPermission
      }
    }
  } else {
    const newId = Math.max(...auditPermissions.value.map(p => p.id), 0) + 1
    const newPermission: AuditPermission = {
      id: newId,
      module: formData.value.module,
      icon: formData.value.icon,
      description: formData.value.description,
      access: formData.value.access,
      category: 'Internal Audit'
    }
    auditPermissions.value.push(newPermission)
  }

  closeModal()
}

function confirmDelete() {
  if (permissionToDelete.value) {
    auditPermissions.value = auditPermissions.value.filter(p => p.id !== permissionToDelete.value!.id)
    permissionToDelete.value = null
  }
}
</script>
