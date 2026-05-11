<template>
  <UModal
    v-model:open="isOpenLocal"
    title=""
    :ui="{
      content: 'sm:max-w-lg',
    }"
  >
    <template #header>
      <div class="flex items-center justify-between w-full">
        <h3 class="text-lg font-semibold flex items-center gap-2">
          <UIcon :name="modalConfig.icon" class="w-5 h-5 text-primary-500" />
          {{ modalConfig.title }}
        </h3>
      </div>
    </template>

    <template #body>
      <!-- Preview Mode -->
      <div v-if="currentMode === 'preview'" class="space-y-5">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">ID Risiko</p>
            <UBadge color="neutral" variant="subtle">#{{ localData.id }}</UBadge>
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Kategori</p>
            <p class="font-medium">
              {{ categoryIcons[localData.category] || '📌' }} {{ localData.category }}
            </p>
          </div>
        </div>

        <div>
          <p class="text-sm text-gray-500 dark:text-gray-400">Nama Risiko</p>
          <p class="text-lg font-semibold">{{ localData.name }}</p>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Impact (Dampak)</p>
            <p class="font-medium">{{ impactLabels[localData.impact] }} ({{ localData.impact }})</p>
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Likelihood (Kemungkinan)</p>
            <p class="font-medium">{{ likelihoodLabels[localData.likelihood] }} ({{ localData.likelihood }})</p>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">Tingkat Risiko</p>
            <UBadge :style="{ backgroundColor: riskConfig.bg, color: riskConfig.color }" class="font-bold">
              {{ riskConfig.label }}
            </UBadge>
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">Severity Weight</p>
            <span class="text-lg font-bold">{{ localData.severity }}</span>
          </div>
        </div>

        <div>
          <p class="text-sm text-gray-500 dark:text-gray-400">Deskripsi</p>
          <div class="mt-1 p-3 bg-gray-50 dark:bg-gray-800 rounded-md text-sm">
            {{ localData.description || 'Tidak ada deskripsi.' }}
          </div>
        </div>
      </div>

      <!-- Add/Edit Form -->
      <Uform v-else id="riskForm" @submit.prevent="submitForm" class="space-y-4">
        <div>
          <label class="text-sm font-medium mb-1 block">Nama Risiko *</label>
          <UInput v-model="localData.name" placeholder="Contoh: Perubahan Regulasi Pajak" class="w-full" />
        </div>

        <div>
          <label class="text-sm font-medium mb-1 block">Kategori</label>
          <USelectMenu
            v-model="localData.category"
            :items="categoryOptions"
            value-key="value"
            placeholder="Pilih Kategori"
            class="w-full"
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="text-sm font-medium mb-1 block">Impact (1-5)</label>
            <USelectMenu
              v-model="localData.impact"
              :items="impactOptions"
              value-key="value"
              placeholder="Impact"
              class="w-full"
            />
          </div>
          <div>
            <label class="text-sm font-medium mb-1 block">Likelihood (1-5)</label>
            <USelectMenu
              v-model="localData.likelihood"
              :items="likelihoodOptions"
              value-key="value"
              placeholder="Likelihood"
              class="w-full"
            />
          </div>
        </div>

        <div>
          <label class="text-sm font-medium mb-1 block">Severity / Danger Weight: {{ localData.severity }}</label>
          <input
            type="range"
            v-model.number="localData.severity"
            min="1"
            max="100"
            class="w-full accent-orange-500"
          />
        </div>

        <div>
          <label class="text-sm font-medium mb-1 block">Deskripsi</label>
          <UTextarea
            v-model="localData.description"
            :rows="3"
            placeholder="Jelaskan risiko ini secara singkat..."
            class="w-full"
          />
        </div>
      </Uform>
      <div class="flex flex-col items-center gap-4 py-4 border-t border-gray-100 ">
          <p class="text-sm text-gray-500 italic">Klik tombol di bawah untuk mengelola langkah mitigasi</p>
          
          <UButton
            label="Open Mitigation Risk & Control"
            icon="i-heroicons-shield-check"
            color="primary"
            variant="solid"
            size="lg"
            class="font-bold shadow-lg px-8"
            @click="goToMitigationPage"
          />
        </div>
    </template>

    <template #footer>
      <div class="flex justify-between w-full">
        <!-- Delete Button and Confirmation -->
        <div class="flex items-center">
          <template v-if="mode !== 'add'">
            <div v-if="!showDeleteConfirm">
              <UButton 
                variant="ghost" 
                color="red" 
                icon="i-lucide-trash-2"
                @click="showDeleteConfirm = true"
              >
                Hapus
              </UButton>
            </div>
            <div v-else class="flex items-center gap-2 bg-red-50 dark:bg-red-900/20 p-1 rounded-md border border-red-200 dark:border-red-800">
              <span class="text-xs font-medium text-red-600 dark:text-red-400 px-2">Yakin hapus?</span>
              <UButton size="xs" color="red" @click="onDelete">Ya, Hapus</UButton>
              <UButton size="xs" variant="ghost" color="neutral" @click="showDeleteConfirm = false">Batal</UButton>
            </div>
          </template>
        </div>

        <div class="flex gap-3">
          <UButton v-if="currentMode === 'preview'" variant="ghost" color="neutral" @click="close">
            Tutup
          </UButton>
          <template v-if="currentMode === 'preview'">
            <UButton color="primary" icon="i-lucide-edit" @click="currentMode = 'edit'">
              Edit Risiko
            </UButton>
          </template>
          <template v-else>
            <UButton variant="ghost" color="neutral" @click="close">
              Batal
            </UButton>
            <UButton type="submit" form="riskForm" color="primary">
              {{ currentMode === 'add' ? 'Tambahkan Risiko' : 'Simpan Perubahan' }}
            </UButton>
          </template>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { getRiskLevel, riskLevelConfig, impactLabels, likelihoodLabels, categoryIcons } from '~/stores/profile-risk'

const goToMitigationPage = () => {
  // Menutup modal terlebih dahulu (opsional, tergantung kebutuhan UX)
  emit("update:isOpen", false);

  // Navigasi ke halaman baru
  router.push({
    path: '/risk-profile/mitigation',
    query: { 
      id: riskData.value.risk_id,
      name: riskData.value.risk_name 
    }
  });
};

const props = defineProps({
  isOpen: Boolean,
  mode: {
    type: String,
    default: 'preview'
  },
  riskData: {
    type: Object,
    default: () => null
  }
})

const emit = defineEmits(['close', 'submit', 'update:isOpen', 'delete'])

const isOpenLocal = computed({
  get: () => props.isOpen,
  set: (value) => {
    emit('update:isOpen', value)
    if (!value) emit('close')
  }
})

const localData = ref({
  id: null,
  name: '',
  category: 'Strategic',
  impact: 3,
  likelihood: 3,
  severity: 50,
  description: ''
})

const currentMode = ref(props.mode)
const showDeleteConfirm = ref(false)

// Options for select menus
const categoryOptions = computed(() =>
  Object.keys(categoryIcons).map(cat => ({ value: cat, label: `${categoryIcons[cat]} ${cat}` }))
)
const impactOptions = computed(() =>
  Object.keys(impactLabels).map(i => ({ value: Number(i), label: `${i} - ${impactLabels[Number(i)]}` }))
)
const likelihoodOptions = computed(() =>
  Object.keys(likelihoodLabels).map(i => ({ value: Number(i), label: `${i} - ${likelihoodLabels[Number(i)]}` }))
)

// Helper to sync data
function syncData() {
  currentMode.value = props.mode
  showDeleteConfirm.value = false
  if (props.mode === 'add') {
    localData.value = {
      id: `NEW-${Date.now().toString().slice(-4)}`,
      name: '',
      category: 'Strategic',
      impact: 3,
      likelihood: 3,
      severity: 50,
      description: ''
    }
  } else if (props.riskData) {
    localData.value = { ...props.riskData }
  }
}

// Sync data when modal opens or riskData changes
watch(() => props.isOpen, (newVal) => {
  if (newVal) syncData()
})

watch(() => props.riskData, () => {
  if (props.isOpen) syncData()
}, { deep: true })

const riskLevel = computed(() => getRiskLevel(localData.value.likelihood, localData.value.impact))
const riskConfig = computed(() => {
  return riskLevelConfig[riskLevel.value] || { label: 'Unknown', color: '#000', bg: '#eee', cellBg: '#333', priority: false }
})

const modalConfig = computed(() => {
  if (currentMode.value === 'preview') return { title: 'Detail Risiko', icon: 'i-lucide-eye' }
  if (currentMode.value === 'edit') return { title: 'Edit Data Risiko', icon: 'i-lucide-edit' }
  return { title: 'Tambah Risiko Baru', icon: 'i-lucide-plus-circle' }
})

function close() {
  emit('close')
  emit('update:isOpen', false)
}

function onDelete() {
  emit('delete', localData.value.id)
  close()
}

function submitForm() {
  if (!localData.value.name.trim()) {
    alert('Nama risiko wajib diisi!')
    return
  }
  emit('submit', { ...localData.value }, currentMode.value)
  close()
}
</script>