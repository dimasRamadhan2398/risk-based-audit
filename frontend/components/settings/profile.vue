<template>
  <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
    <template #header>
      <div>
        <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.profile.title') }}</h3>
        <p class="text-md sm:text-sm text-gray-500 dark:text-gray-400">{{ t('settings.profile.subtitle') }}</p>
      </div>
    </template>

    <div class="space-y-6">
      <!-- Profile Picture Section -->
      <div class="flex items-center gap-6 pb-6 border-b border-gray-100 dark:border-gray-800">
        <div class="w-20 h-20 rounded-2xl bg-primary-100 dark:bg-primary-950 flex items-center justify-center border border-primary-200 dark:border-primary-800 overflow-hidden shrink-0 shadow-md">
          <img v-if="preview" :src="preview" alt="Avatar" class="w-full h-full object-cover" />
          <span v-else class="text-2xl font-extrabold text-primary-700 dark:text-primary-400">
            {{ userInitial }}
          </span>
        </div>
        <div class="space-y-2">
          <div class="flex items-center gap-3">
            <UButton color="primary" variant="soft" class="font-semibold rounded-xl" @click="triggerUpload">
              {{ t('settings.profile.changePhoto') }}
            </UButton>
            <UButton v-if="preview" color="neutral" variant="ghost" class="rounded-xl text-md" @click="preview = ''">
              {{ t('settings.profile.remove') }}
            </UButton>
          </div>
          <p class="text-md text-gray-500 dark:text-gray-400">{{ t('settings.profile.fileLimits', { size: mamdizeMB }) }}</p>
          <p v-if="statusMessage" class="text-md font-semibold" :class="isError ? 'text-red-500' : 'text-emerald-600'">
            {{ statusMessage }}
          </p>
        </div>
      </div>

      <!-- User Information Form -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <UFormField :label="t('settings.profile.fullName')" :help="t('settings.profile.fullNameHelp')">
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

        <UFormField :label="t('settings.profile.email')" :help="t('settings.profile.emailHelp')">
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

        <UFormField :label="t('settings.profile.department')">
          <USelect
            v-model="form.department"
            :placeholder="t('settings.profile.selectDepartment')"
            :options="departments"
            size="lg"
            class="w-full"
          />
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
        <UButton variant="subtle" color="neutral" class="rounded-xl font-semibold px-5" @click="resetForm">
          {{ t('settings.profile.cancel') }}
        </UButton>
        <UButton color="primary" class="rounded-xl font-bold px-6" :loading="saving" @click="saveProfile">
          {{ t('settings.profile.saveChanges') }}
        </UButton>
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const props = defineProps({
  accept: {
    type: Array as PropType<string[]>,
    default: () => ['image/jpeg', 'image/png', 'image/webp', 'image/gif'],
  },
  mamdizeMB: {
    type: Number,
    default: 5,
  },
})

const authStore = useAuthStore()
const toast = useToast()
const { t } = useI18n()

const departments = [
  { label: 'Internal Audit', value: 'internal-audit' },
  { label: 'Risk Management', value: 'risk-management' },
  { label: 'Compliance', value: 'compliance' },
  { label: 'Finance', value: 'finance' },
  { label: 'Information Technology', value: 'it' },
]

const form = ref({
  fullName: '',
  email: '',
  phone: '',
  department: '',
  position: '',
})

const saving = ref(false)
const isLoading = ref(false)
const isError = ref(false)
const statusMessage = ref('')
const preview = ref('')

const userInitial = computed(() => {
  const name = form.value.fullName || authStore.user?.username || 'U'
  return name.charAt(0).toUpperCase()
})

const syncFormFromStore = () => {
  if (authStore.user) {
    form.value.fullName = authStore.user.fullName || authStore.user.username || ''
    form.value.email = authStore.user.email || ''
    form.value.phone = authStore.user.phone || ''
    form.value.department = authStore.user.department || 'internal-audit'
    form.value.position = authStore.user.position || ''
  }
}

const fetchProfileData = async () => {
  try {
    const profile = await authStore.fetchUserProfile()
    if (profile) {
      form.value.fullName = profile.full_name || profile.fullName || authStore.user?.username || ''
      form.value.email = profile.email || authStore.user?.email || ''
      form.value.phone = profile.phone || ''
      form.value.department = profile.department || 'internal-audit'
      form.value.position = profile.position || ''
    } else {
      syncFormFromStore()
    }
  } catch {
    syncFormFromStore()
  }
}

onMounted(() => {
  fetchProfileData()
})

watch(() => authStore.user, syncFormFromStore, { immediate: true })

const resetForm = () => {
  fetchProfileData()
  clearStatus()
}

const saveProfile = async () => {
  saving.value = true
  try {
    await authStore.updateProfile({
      fullName: form.value.fullName,
      phone: form.value.phone,
      department: form.value.department,
      position: form.value.position,
    })

    toast.add({
      title: t('settings.profile.toastSuccessTitle'),
      description: t('settings.profile.toastSuccessDesc'),
      color: 'success',
      icon: 'i-lucide-check-circle',
    })
  } catch (err: any) {
    toast.add({
      title: t('settings.profile.toastErrorTitle'),
      description: err.message || t('settings.profile.toastErrorDesc'),
      color: 'error',
      icon: 'i-lucide-alert-triangle',
    })
  } finally {
    saving.value = false
  }
}

function setStatus(message: string, error = false) {
  statusMessage.value = message
  isError.value = error
}

function clearStatus() {
  statusMessage.value = ''
  isError.value = false
}

function validateFile(file: File) {
  if (!props.accept.includes(file.type)) {
    const formats = props.accept.map((t) => t.split('/')[1]!.toUpperCase()).join(', ')
    return `Unsupported format. Accepted: ${formats}`
  }
  if (file.size > props.mamdizeMB * 1024 * 1024) {
    return `File too large. Maximum allowed size is ${props.mamdizeMB}MB.`
  }
  return null
}

const triggerUpload = () => {
  clearStatus()

  const input = document.createElement('input')
  input.type = 'file'
  input.accept = props.accept.join(',')
  input.style.display = 'none'
  document.body.appendChild(input)

  input.addEventListener('cancel', () => {
    document.body.removeChild(input)
  })

  input.addEventListener('change', () => {
    const file = input.files?.[0]
    document.body.removeChild(input)

    if (!file) return

    const validationError = validateFile(file)
    if (validationError) {
      setStatus(validationError, true)
      return
    }

    isLoading.value = true
    const reader = new FileReader()

    reader.onerror = () => {
      const msg = 'Failed to read the file. Please try again.'
      setStatus(msg, true)
      isLoading.value = false
    }

    reader.onload = (e: ProgressEvent<FileReader>) => {
      const result = e.target?.result
      if (!result || typeof result !== 'string') {
        const msg = 'Failed to read the file. Please try again.'
        setStatus(msg, true)
        isLoading.value = false
        return
      }
      preview.value = result
      isLoading.value = false
      setStatus(`✓ ${file.name} (${(file.size / 1024).toFixed(1)}KB)`)
    }

    reader.readAsDataURL(file)
  })

  input.click()
}
</script>