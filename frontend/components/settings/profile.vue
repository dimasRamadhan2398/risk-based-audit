<template>
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900">My Profile</h3>
        <p class="text-sm text-gray-600">Manage your personal information</p>
      </template>

      <div class="space-y-6">
        <div class="flex items-center gap-6">
          <div class="w-24 h-24 rounded-full flex items-center justify-center w-">
            <img src="../../assets/images/utils/avatar-placeholder.png" alt="avatar" />
          </div>
          <div>
            <UButton color="primary" variant="soft" @click="triggerUpload">Change Photo</UButton>
            <p class="text-xs text-gray-500 mt-2">JPG, PNG or GIF. Max 2MB.</p>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-6">
          <UFormField label="Full Name" orientation="horizontal" class="min-w-14">
            <UInput placeholder="Enter your full name" model-value="User Name" variant="outline" color="neutral" size="md"/>
          </UFormField>

          <UFormField label="Email" orientation="horizontal" help="Email yang terhubung dengan SSO tidak dapat diubah.">
            <UInput type="email" placeholder="Enter your email" model-value="user@example.com" />
          </UFormField>

          <UFormField label="Phone Number" orientation="horizontal">
            <UInput placeholder="Enter your phone number" />
          </UFormField>

          <UFormField label="Department" orientation="horizontal">
            <USelect placeholder="Select department" :options="departments" class="w-full"/>
          </UFormField>
        </div>

        <UFormField label="Jabatan" orientation="horizontal">
          <UInput placeholder="Masukkan Jabatan..." />
        </UFormField>

        <div class="flex justify-end gap-3">
          <UButton variant="outline">Cancel</UButton>
          <UButton color="primary">Save Changes</UButton>
        </div>
      </div>
    </UCard>
</template>
<script setup lang="ts">
const props = defineProps({
  accept: {
    type: Array<string>,
    default: () => ['image/jpeg', 'image/png', 'image/webp', 'image/gif'],
  },
  maxSizeMB: {
    type: Number,
    default: 5,
  },
})

const departments = [
  { label: "Internal Audit", value: "internal-audit" },
  { label: "Risk Management", value: "risk-management" },
  { label: "Compliance", value: "compliance" },
  { label: "Finance", value: "finance" },
];

const isLoading = ref(false)
const isError = ref(false)
const statusMessage = ref('')
const preview = ref('')

function setStatus(message : string, error = false) {
  statusMessage.value = message
  isError.value = error
}
function clearStatus() {
  statusMessage.value = ''
  isError.value = false
}

function validateFile(file : File) {
  if (!props.accept.includes(file.type)) {
    const formats = props.accept.map(t => t.split('/')[1]!.toUpperCase()).join(', ')
    return `Unsupported format. Accepted: ${formats}`
  }
  if (file.size > props.maxSizeMB * 1024 * 1024) {
    return `File too large. Maximum allowed size is ${props.maxSizeMB}MB.`
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

    // Validate
    const validationError = validateFile(file)
    if (validationError) {
      setStatus(validationError, true)
      return
    }

    // Read file
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
      const dataURL = result
      const objectURL = URL.createObjectURL(file)

      preview.value = dataURL
      isLoading.value = false
      setStatus(`✓ ${file.name} (${(file.size / 1024).toFixed(1)}KB)`)

      // TODO: send to server
      // const formData = new FormData()
      // formData.append('photo', file)
      // await fetch('/api/upload', { method: 'POST', body: formData })
    }

    reader.readAsDataURL(file)
  })

  input.click()
}
</script>