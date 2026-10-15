<template>
  <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
    <template #header>
      <div>
        <h3 class="text-lg font-bold text-gray-900 dark:text-white">
          {{ t('settings.profile.title') }}
        </h3>
        <p class="text-md sm:text-sm text-gray-500 dark:text-gray-400">
          {{ t('settings.profile.subtitle') }}
        </p>
      </div>
    </template>

    <div class="space-y-6">
      <!-- Profile Picture Section -->
      <div class="flex items-center gap-6 pb-6 border-b border-gray-100 dark:border-gray-800">
        <div class="w-20 h-20 rounded-2xl bg-primary-100 dark:bg-primary-950 flex items-center justify-center border border-primary-200 dark:border-primary-800 overflow-hidden shrink-0 shadow-md">
          <img
            v-if="preview"
            :src="preview"
            alt="Avatar"
            class="w-full h-full object-cover"
          >
          <span
            v-else
            class="text-2xl font-extrabold text-primary-700 dark:text-primary-400"
          >
            {{ userInitial }}
          </span>
        </div>
        <div class="space-y-2">
          <div class="flex items-center gap-3">
            <UButton
              color="primary"
              variant="soft"
              class="font-semibold rounded-xl"
              :loading="processingAvatar"
              @click="triggerUpload"
            >
              {{ t('settings.profile.changePhoto') }}
            </UButton>
            <UButton
              v-if="preview"
              color="neutral"
              variant="ghost"
              class="rounded-xl text-md"
              @click="removeAvatar"
            >
              {{ t('settings.profile.remove') }}
            </UButton>
          </div>
          <p class="text-md text-gray-500 dark:text-gray-400">
            {{ t('settings.profile.fileLimits', { size: AVATAR_MAX_INPUT_MB }) }}
          </p>
          <p
            v-if="statusMessage"
            class="text-md font-semibold"
            :class="isError ? 'text-red-500' : 'text-emerald-600'"
          >
            {{ statusMessage }}
          </p>
        </div>
      </div>

      <!-- User Information Form -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <UFormField
          :label="t('settings.profile.fullName')"
          :help="t('settings.profile.fullNameHelp')"
        >
          <UInput
            v-model="form.fullName"
            :placeholder="t('settings.profile.fullNamePlaceholder')"
            size="lg"
            class="w-full"
          />
        </UFormField>

        <UFormField :label="t('settings.profile.username')">
          <UInput
            :model-value="authStore.user?.username || '—'"
            disabled
            size="lg"
            class="w-full bg-gray-50 dark:bg-gray-900"
          />
        </UFormField>

        <UFormField
          :label="t('settings.profile.email')"
          :help="t('settings.profile.emailHelp')"
        >
          <UInput
            v-model="form.email"
            type="email"
            placeholder="Enter your email"
            disabled
            size="lg"
            class="w-full bg-gray-50 dark:bg-gray-900"
          />
        </UFormField>

        <UFormField :label="t('settings.profile.phone')">
          <UInput
            v-model="form.phone"
            :placeholder="t('settings.profile.phonePlaceholder')"
            size="lg"
            class="w-full"
          />
        </UFormField>

        <UFormField
          :label="t('settings.profile.department')"
          :error="departmentsError || undefined"
        >
          <USelectMenu
            v-model="form.department"
            :items="departmentOptions"
            value-key="value"
            label-key="label"
            :placeholder="t('settings.profile.selectDepartment')"
            :search-input="{ placeholder: t('settings.profile.departmentSearchPlaceholder') }"
            size="lg"
            class="w-full"
            :loading="loadingDepartments"
            :disabled="loadingDepartments || !!departmentsError"
          >
            <template #empty>
              {{ t('settings.profile.departmentsEmpty') }}
            </template>
          </USelectMenu>
          <UButton
            v-if="departmentsError"
            color="neutral"
            variant="link"
            size="xs"
            class="px-0 mt-1"
            @click="fetchDepartments"
          >
            {{ t('settings.profile.departmentsRetry') }}
          </UButton>
        </UFormField>

        <UFormField :label="t('settings.profile.position')">
          <UInput
            v-model="form.position"
            :placeholder="t('settings.profile.positionPlaceholder')"
            size="lg"
            class="w-full"
          />
        </UFormField>
      </div>

      <!-- User Roles -->
      <div class="pt-4 border-t border-gray-100 dark:border-gray-800">
        <UFormField :label="t('settings.profile.assignedRoles')">
          <div class="flex flex-wrap gap-2 pt-1">
            <UBadge
              v-for="role in (authStore.user?.roles || ['User'])"
              :key="role"
              color="primary"
              variant="subtle"
              class="font-semibold px-3 py-1 rounded-full text-md"
            >
              {{ role }}
            </UBadge>
          </div>
        </UFormField>
      </div>

      <!-- Actions -->
      <div class="flex justify-end gap-3 pt-4 border-t border-gray-100 dark:border-gray-800">
        <UButton
          color="primary"
          class="rounded-xl font-bold px-6"
          :loading="saving"
          @click="saveProfile"
        >
          {{ t('settings.profile.saveChanges') }}
        </UButton>
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useDepartmentApi } from '~/composables/useDepartmentApi'
import type { Department } from '~/types/master'
import { getErrorStatus, getUserErrorMessage } from '~/utils/error'
import { AVATAR_ALLOWED_TYPES, AVATAR_MAX_INPUT_MB, AvatarError, prepareAvatar } from '~/utils/avatar'

const authStore = useAuthStore()
const departmentApi = useDepartmentApi()
const toast = useToast()
const { t } = useI18n()

interface DepartmentOption {
  label: string
  value: string
}

const departments = ref<DepartmentOption[]>([])
const loadingDepartments = ref(false)
const departmentsError = ref('')

const form = ref({
  fullName: '',
  email: '',
  phone: '',
  department: '',
  position: ''
})

const saving = ref(false)
const processingAvatar = ref(false)
const isError = ref(false)
const statusMessage = ref('')

// Pending avatar change: undefined = untouched, '' = remove, data URL = replace
const pendingAvatar = ref<string | undefined>(undefined)

const preview = computed(() =>
  pendingAvatar.value !== undefined ? pendingAvatar.value : authStore.user?.avatarUrl || ''
)

const userInitial = computed(() => {
  const name = form.value.fullName || authStore.user?.username || 'U'
  return name.charAt(0).toUpperCase()
})

// Master departments, plus the user's current value when it is not in the list so it is never lost
const departmentOptions = computed<DepartmentOption[]>(() => {
  const options = [...departments.value]
  const known = new Set(options.map(o => o.value.toLowerCase()))
  for (const name of [authStore.user?.department, form.value.department]) {
    const trimmed = (name || '').trim()
    if (trimmed && !known.has(trimmed.toLowerCase())) {
      known.add(trimmed.toLowerCase())
      options.push({ label: trimmed, value: trimmed })
    }
  }
  return options
})

// Use the master data spelling when the stored value differs only by case/whitespace
const normalizeDepartment = (value?: string) => {
  const current = (value || '').trim()
  if (!current) return ''
  const match = departments.value.find(d => d.value.toLowerCase() === current.toLowerCase())
  return match ? match.value : current
}

const fetchDepartments = async () => {
  loadingDepartments.value = true
  departmentsError.value = ''
  try {
    const { departments: list } = await departmentApi.getDepartments({ page: 1, page_size: 1000 })
    departments.value = list
      .filter((d: Department) => d.is_active !== false && d.department_name)
      .map((d: Department) => ({ label: d.department_name, value: d.department_name }))
    form.value.department = normalizeDepartment(form.value.department)
  } catch (err) {
    console.error('Failed to load master departments:', err)
    departmentsError.value = getUserErrorMessage(err, t, { fallbackKey: 'settings.profile.departmentsLoadFailed' })
  } finally {
    loadingDepartments.value = false
  }
}

const applyUserToForm = (user: { fullName?: string, full_name?: string, username?: string, email?: string, phone?: string, department?: string, position?: string } | null | undefined) => {
  if (!user) return
  form.value.fullName = user.full_name || user.fullName || authStore.user?.username || ''
  form.value.email = user.email || authStore.user?.email || ''
  form.value.phone = user.phone || ''
  form.value.department = normalizeDepartment(user.department)
  form.value.position = user.position || ''
}

const fetchProfileData = async () => {
  try {
    const profile = await authStore.fetchUserProfile()
    applyUserToForm(profile ?? authStore.user)
  } catch {
    applyUserToForm(authStore.user)
  }
}

onMounted(async () => {
  await Promise.all([fetchDepartments(), fetchProfileData()])
  // Departments may have loaded after the form was filled
  form.value.department = normalizeDepartment(form.value.department)
})

watch(() => authStore.user, () => applyUserToForm(authStore.user), { immediate: true })

const setStatus = (message: string, error = false) => {
  statusMessage.value = message
  isError.value = error
}

const clearStatus = () => {
  statusMessage.value = ''
  isError.value = false
}

const avatarErrorMessage = (err: unknown) => {
  const code = err instanceof AvatarError ? err.code : 'readFailed'
  const keys = {
    invalidType: 'settings.profile.avatarInvalidType',
    tooLargeInput: 'settings.profile.avatarTooLargeInput',
    readFailed: 'settings.profile.avatarReadFailed',
    tooLargeOutput: 'settings.profile.avatarTooLargeOutput'
  } as const
  return t(keys[code], { size: AVATAR_MAX_INPUT_MB })
}

const removeAvatar = () => {
  pendingAvatar.value = ''
  setStatus(t('settings.profile.avatarRemovePending'))
}

const handleFileChosen = async (file: File) => {
  processingAvatar.value = true
  setStatus(t('settings.profile.avatarProcessing'))
  try {
    pendingAvatar.value = await prepareAvatar(file)
    setStatus(t('settings.profile.avatarReady', { name: file.name }))
  } catch (err) {
    if (!(err instanceof AvatarError)) console.error('Avatar processing failed:', err)
    setStatus(avatarErrorMessage(err), true)
  } finally {
    processingAvatar.value = false
  }
}

const triggerUpload = () => {
  clearStatus()

  const input = document.createElement('input')
  input.type = 'file'
  input.accept = AVATAR_ALLOWED_TYPES.join(',')
  input.style.display = 'none'
  document.body.appendChild(input)

  input.addEventListener('cancel', () => {
    document.body.removeChild(input)
  })

  input.addEventListener('change', () => {
    const file = input.files?.[0]
    document.body.removeChild(input)
    if (file) handleFileChosen(file)
  })

  input.click()
}

const saveProfile = async () => {
  saving.value = true
  const avatarChanged = pendingAvatar.value !== undefined
  try {
    await authStore.updateProfile({
      fullName: form.value.fullName,
      phone: form.value.phone,
      department: form.value.department,
      position: form.value.position,
      // Only sent when the user changed it
      ...(avatarChanged ? { avatarUrl: pendingAvatar.value } : {})
    })
    pendingAvatar.value = undefined
    clearStatus()

    toast.add({
      title: t('settings.profile.toastSuccessTitle'),
      description: t('settings.profile.toastSuccessDesc'),
      color: 'success',
      icon: 'i-lucide-check-circle'
    })
  } catch (err: unknown) {
    console.error('Failed to save profile:', err)
    const status = getErrorStatus(err)
    const message = getUserErrorMessage(err, t, { fallbackKey: 'settings.profile.toastErrorDesc' })
    // A generic validation error while a photo is attached is almost certainly the photo
    const photoRejected = avatarChanged && pendingAvatar.value && (status === 413 || (status === 400 && message === t('errors.validation')))
    toast.add({
      title: t('settings.profile.toastErrorTitle'),
      description: photoRejected ? t('settings.profile.avatarRejected') : message,
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    saving.value = false
  }
}
</script>
