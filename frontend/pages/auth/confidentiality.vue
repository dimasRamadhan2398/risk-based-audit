<template>
  <div class="min-h-screen flex bg-[var(--bg-main)] transition-colors duration-300 relative overflow-hidden">
    <!-- Left panel: Vertical Stepper (hidden on mobile) -->
    <div class="hidden lg:flex lg:w-4/12 relative overflow-hidden border-r border-[var(--border-main)] p-12 flex-col justify-between select-none" style="background: linear-gradient(160deg, color-mix(in srgb, var(--color-secondary-500) 4%, var(--bg-surface)) 0%, var(--bg-surface) 40%, color-mix(in srgb, var(--color-primary-500) 3%, var(--bg-main)) 100%)">
      <!-- Background subtle grid -->
      <div class="absolute inset-0 bg-[linear-gradient(to_bottom,rgba(var(--ui-primary),0.01)_1px,transparent_1px),linear-gradient(to_right,rgba(var(--ui-primary),0.01)_1px,transparent_1px)] bg-[size:48px_48px] opacity-40" />

      <!-- Top brand logo -->
      <div class="relative z-10 flex items-center gap-3">
        <Logo class="h-10" />
      </div>

      <!-- Mid content: Vertical stepper progress -->
      <div class="relative z-10 my-auto space-y-8 pl-2">
        <div class="flex items-start gap-4">
          <span class="w-6 h-6 rounded-full bg-success-500/10 border border-success-500/20 flex items-center justify-center text-success-600 font-bold text-md">✓</span>
          <div>
            <h4 class="text-sm font-bold text-[var(--text-main)]">Autentikasi</h4>
            <p class="text-md text-[var(--text-muted)]">Kredensial login berhasil diverifikasi</p>
          </div>
        </div>
        
        <div class="flex items-start gap-4">
          <span 
            class="w-6 h-6 rounded-full flex items-center justify-center font-bold text-md transition-all duration-300"
            :class="currentStep === 2 ? 'bg-secondary-500 text-white shadow-lg shadow-secondary-500/25' : currentStep > 2 ? 'bg-success-500/10 border border-success-500/20 text-success-600' : 'bg-[var(--border-main)] text-[var(--text-muted)]'"
          >
            <span v-if="currentStep > 2">✓</span>
            <span v-else>2</span>
          </span>
          <div>
            <h4 
              class="text-sm font-bold transition-all duration-300"
              :class="currentStep === 2 ? 'text-secondary-500' : 'text-[var(--text-main)]'"
            >Pakta Integritas</h4>
            <p class="text-md text-[var(--text-muted)]">Persetujuan ketentuan kerahasiaan data</p>
          </div>
        </div>

        <div class="flex items-start gap-4">
          <span 
            class="w-6 h-6 rounded-full flex items-center justify-center font-bold text-md transition-all duration-300"
            :class="currentStep === 3 ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/25' : 'bg-[var(--border-main)] text-[var(--text-muted)]'"
          >3</span>
          <div>
            <h4 
              class="text-sm font-bold transition-all duration-300"
              :class="currentStep === 3 ? 'text-primary-500' : 'text-[var(--text-main)]'"
            >Lengkapi Profil & Peran</h4>
            <p class="text-md text-[var(--text-muted)]">Informasi profil & tinjauan hak akses (RBAC)</p>
          </div>
        </div>

        <div class="flex items-start gap-4">
          <span class="w-6 h-6 rounded-full bg-[var(--border-main)] text-[var(--text-muted)] flex items-center justify-center font-bold text-md">4</span>
          <div>
            <h4 class="text-sm font-bold text-[var(--text-main)]">Selesai</h4>
            <p class="text-md text-[var(--text-muted)]">Masuk ke dasbor sistem utama</p>
          </div>
        </div>
      </div>

      <!-- Bottom metadata -->
      <div class="relative z-10 text-md text-[var(--text-muted)] opacity-60">
        <span>Sistem Audit Internal Berbasis Risiko v1.0</span>
      </div>
    </div>

    <!-- Right panel: Full screen content area -->
    <div class="w-full lg:w-8/12 flex flex-col justify-between bg-[var(--bg-main)] overflow-y-auto min-h-screen">
      <!-- Mobile top stepper banner (hidden on large screens) -->
      <div class="lg:hidden px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] flex items-center justify-between text-md font-semibold select-none">
        <span class="text-primary-500">Langkah {{ currentStep }} dari 3</span>
        <span class="text-[var(--text-muted)] font-medium">Pakta Integritas & Profil</span>
      </div>

      <!-- Center content section -->
      <div class="my-auto max-w-2xl w-full mx-auto p-6 sm:p-12 md:p-16 lg:p-20 space-y-8 animate-fade-in">
        
        <!-- Onboarding Step 2: Pakta Integritas -->
        <div v-if="currentStep === 2" class="space-y-6">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-xl bg-secondary-500/10 border border-secondary-500/20 flex items-center justify-center text-secondary-500">
              <UIcon name="i-lucide-newspaper" class="w-7 h-7" />
            </div>
            <div>
              <h1 class="text-2xl font-bold text-[var(--text-main)]">Pakta Integritas</h1>
              <p class="text-sm text-[var(--text-muted)]">Wajib disetujui sebelum menggunakan sistem</p>
            </div>
          </div>

          <!-- User initials metadata row -->
          <div class="flex items-center gap-3 border-b border-[var(--border-main)] pb-4 text-md text-[var(--text-muted)]">
            <span class="font-bold text-[var(--text-main)] capitalize">{{ authStore.user?.fullName }}</span>
            <span>·</span>
            <span>{{ authStore.user?.username }}</span>
            <span class="ml-auto">{{ currentDateTime }}</span>
          </div>

          <!-- Document content container, setting text color to white in dark mode -->
          <div
            ref="scrollContainer"
            class="border border-[var(--border-main)] rounded-xl bg-[var(--bg-surface)] p-6 max-h-96 overflow-y-auto text-slate-700 dark:text-white text-sm leading-relaxed space-y-4 scrollbar-thin scrollbar-thumb-[var(--border-main)] scrollbar-track-transparent"
            @scroll="handleScroll"
          >
            <p class="text-[var(--text-main)] dark:text-white font-bold text-base border-b border-[var(--border-main)] pb-3">
              PAKTA INTEGRITAS<br>
              SISTEM AUDIT INTERNAL BERBASIS RISIKO (RBIA)
            </p>

            <p>
              Saya yang bertanda tangan di bawah ini, dengan ini menyatakan dan berjanji dengan sungguh-sungguh bahwa:
            </p>

            <ol class="list-decimal list-inside space-y-3 pl-2">
              <li>
                <strong class="text-[var(--text-main)] dark:text-white">Kerahasiaan Informasi:</strong>
                Saya akan menjaga kerahasiaan seluruh informasi, data, dan dokumen yang saya akses melalui sistem ini. Saya tidak akan mengungkapkan, menyebarkan, atau menggunakan informasi tersebut untuk kepentingan pribadi atau pihak lain yang tidak berwenang.
              </li>
              <li>
                <strong class="text-[var(--text-main)] dark:text-white">Penggunaan yang Sah:</strong>
                Saya hanya akan menggunakan sistem ini untuk keperluan pekerjaan yang sah dan sesuai dengan tugas dan wewenang yang diberikan kepada saya oleh organisasi.
              </li>
              <li>
                <strong class="text-[var(--text-main)] dark:text-white">Integritas Data:</strong>
                Saya tidak akan memanipulasi, memalsukan, atau menghapus data dan informasi dalam sistem ini tanpa izin yang sah dari pihak yang berwenang.
              </li>
              <li>
                <strong class="text-[var(--text-main)] dark:text-white">Keamanan Akun:</strong>
                Saya bertanggung jawab penuh atas keamanan kredensial akses saya (username dan password). Saya tidak akan berbagi akses saya kepada pihak lain dan akan segera melaporkan jika terjadi penyalahgunaan.
              </li>
              <li>
                <strong class="text-[var(--text-main)] dark:text-white">Kepatuhan Regulasi:</strong>
                Saya akan mematuhi seluruh peraturan perundang-undangan yang berlaku, termasuk namun tidak terbatas pada ketentuan mengenai perlindungan data pribadi dan kerahasiaan informasi perusahaan.
              </li>
              <li>
                <strong class="text-[var(--text-main)] dark:text-white">Pelaporan Pelanggaran:</strong>
                Saya akan melaporkan setiap pelanggaran atau potensi pelanggaran atas ketentuan ini kepada atasan atau pihak yang berwenang sesegera mungkin.
              </li>
              <li>
                <strong class="text-[var(--text-main)] dark:text-white">Sanksi:</strong>
                Saya memahami bahwa pelanggaran atas pakta integritas ini dapat mengakibatkan sanksi administratif, disiplin, dan/atau sanksi hukum sesuai dengan ketentuan yang berlaku.
              </li>
            </ol>

            <p class="border-t border-[var(--border-main)] pt-4 text-md text-[var(--text-muted)]">
              Dengan mengklik tombol <strong class="text-[var(--text-main)] dark:text-white">"Saya Setuju"</strong>, saya menyatakan bahwa saya telah membaca, memahami, dan menyetujui seluruh ketentuan dalam Pakta Integritas ini. Persetujuan ini bersifat mengikat secara hukum dan akan dicatat dalam sistem beserta timestamp, alamat IP, dan informasi perangkat yang digunakan.
            </p>
          </div>

          <!-- Scroll progress indicator -->
          <div class="space-y-2">
            <div class="h-1.5 bg-[var(--border-main)] rounded-full overflow-hidden">
              <div
                class="h-full rounded-full transition-all duration-500"
                :style="{ width: `${scrollProgress}%`, background: 'linear-gradient(90deg, var(--color-secondary-500), var(--color-primary-500))' }"
              />
            </div>
            <p v-if="!hasScrolledToBottom" class="text-md text-[var(--text-muted)] text-center flex items-center justify-center gap-1">
              <span>↓</span> Gulir ke bawah untuk membaca seluruh ketentuan
            </p>
            <p v-else class="text-md text-success-600 dark:text-success-400 text-center flex items-center justify-center gap-1 font-medium">
              <span>✓</span> Anda telah membaca seluruh ketentuan
            </p>
          </div>

          <!-- Action buttons for Step 2 -->
          <div class="flex flex-col sm:flex-row gap-3 pt-4">
            <button
              id="confidentiality-reject-btn"
              class="flex-1 px-4 py-3 rounded-xl border border-error-500/30 bg-error-500/10 text-error-600 dark:text-error-400 hover:bg-error-500/20 text-sm font-semibold transition-all duration-200"
              @click="handleReject"
            >
              ✕ Tolak & Keluar
            </button>
            <button
              id="confidentiality-accept-btn"
              :disabled="!hasScrolledToBottom || accepting"
              class="flex-1 px-4 py-3 rounded-xl font-bold text-sm transition-all duration-200"
              :style="hasScrolledToBottom && !accepting ? { background: 'linear-gradient(135deg, var(--color-secondary-500), var(--color-primary-500))', bomdhadow: '0 8px 24px -4px color-mix(in srgb, var(--color-secondary-500) 30%, transparent)' } : {}"
              :class="hasScrolledToBottom && !accepting
                ? 'text-white hover:opacity-90'
                : 'bg-[var(--border-main)] text-[var(--text-muted)] cursor-not-allowed'"
              @click="handleAccept"
            >
              <span v-if="accepting">Memproses...</span>
              <span v-else>✓ Saya Setuju & Lanjutkan</span>
            </button>
          </div>
        </div>

        <!-- Onboarding Step 3: Profile Completion -->
        <div v-else-if="currentStep === 3" class="space-y-6">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500">
              <UIcon name="i-lucide-user-round-cog" class="w-7 h-7" />
            </div>
            <div>
              <h1 class="text-2xl font-bold text-[var(--text-main)]">Lengkapi Profil & Peran</h1>
              <p class="text-sm text-[var(--text-muted)]">Sesuaikan profil dan tinjau hak akses Anda di sistem</p>
            </div>
          </div>

          <!-- Form inputs -->
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField name="fullName">
                <template #label>
                  <span class="text-md font-semibold text-[var(--text-main)]">Nama Lengkap</span>
                </template>
                <UInput
                  v-model="profileState.fullName"
                  placeholder="Masukkan nama lengkap"
                  class="w-full"
                  :ui="{
                    base: 'bg-[var(--bg-surface)] border-[var(--border-main)] text-[var(--text-main)] rounded-xl py-3 px-4',
                  }"
                />
              </UFormField>

              <UFormField name="phone">
                <template #label>
                  <span class="text-md font-semibold text-[var(--text-main)]">Nomor Telepon</span>
                </template>
                <UInput
                  v-model="profileState.phone"
                  placeholder="Masukkan nomor telepon"
                  class="w-full"
                  :ui="{
                    base: 'bg-[var(--bg-surface)] border-[var(--border-main)] text-[var(--text-main)] rounded-xl py-3 px-4',
                  }"
                />
              </UFormField>

              <UFormField name="department" class="md:col-span-2">
                <template #label>
                  <span class="text-md font-semibold text-[var(--text-main)]">Unit Kerja / Departemen</span>
                </template>
                <UInput
                  v-model="profileState.department"
                  placeholder="Masukkan unit kerja"
                  class="w-full"
                  :ui="{
                    base: 'bg-[var(--bg-surface)] border-[var(--border-main)] text-[var(--text-main)] rounded-xl py-3 px-4',
                  }"
                />
              </UFormField>
            </div>
          </div>

          <!-- RBAC details -->
          <div class="border-t border-[var(--border-main)] pt-6 space-y-4">
            <div>
              <span class="text-md font-semibold text-[var(--text-main)] block mb-2">Peran Anda (Default Database)</span>
              <div class="flex flex-wrap gap-2">
                <span v-for="role in userRoles" :key="role" class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-md font-bold bg-secondary-500/10 border border-secondary-500/20 text-secondary-600 dark:text-secondary-400 capitalize">
                  <UIcon name="i-lucide-key-round" class="w-3 h-3" /> {{ role }}
                </span>
              </div>
            </div>

            <div>
              <span class="text-md font-semibold text-[var(--text-main)] block mb-2">Daftar Hak Akses Sistem Anda:</span>
              <div class="bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-xl p-4 max-h-48 overflow-y-auto space-y-2.5">
                <div v-for="(perm, idx) in permissionsList" :key="perm" class="flex items-start gap-2.5 text-md text-[var(--text-main)]">
                  <UIcon 
                    name="i-lucide-shield-check" 
                    class="w-3.5 h-3.5 mt-0.5 flex-shrink-0"
                    :class="idx % 2 === 0 ? 'text-secondary-500' : 'text-primary-500'"
                  />
                  <div class="flex flex-col">
                    <span class="font-mono font-bold text-[10px] uppercase tracking-wide" :class="idx % 2 === 0 ? 'text-secondary-500/70' : 'text-primary-500/70'">{{ perm }}</span>
                    <span class="text-[var(--text-muted)] mt-0.5">{{ getPermissionLabel(perm) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Error message -->
          <Transition name="fade">
            <div v-if="error" class="flex items-start gap-3 rounded-xl bg-error-500/10 border border-error-500/20 px-4 py-3">
              <span class="text-error-500 text-lg leading-none">✕</span>
              <p class="text-sm text-error-600 dark:text-error-400">{{ error }}</p>
            </div>
          </Transition>

          <!-- Action buttons for Step 3 -->
          <div class="flex flex-col sm:flex-row gap-3 pt-4">
            <button
              class="flex-1 px-4 py-3 rounded-xl border border-[var(--border-main)] hover:bg-[var(--bg-surface)] text-sm font-semibold transition-all duration-200 text-[var(--text-main)]"
              @click="currentStep = 2"
            >
              ← Kembali
            </button>
            <button
              id="profile-complete-btn"
              :disabled="savingProfile || !profileState.fullName"
              class="flex-1 px-4 py-3 rounded-xl font-bold text-sm transition-all duration-200 text-white disabled:bg-[var(--border-main)] disabled:text-[var(--text-muted)] disabled:cursor-not-allowed"
              :style="!(savingProfile || !profileState.fullName) ? { background: 'linear-gradient(135deg, var(--color-primary-500), var(--color-secondary-500))', bomdhadow: '0 8px 24px -4px color-mix(in srgb, var(--color-primary-500) 30%, transparent)' } : {}"
              @click="handleSaveProfile"
            >
              <span v-if="savingProfile">Menyimpan...</span>
              <span v-else>Simpan & Masuk ke Dashboard →</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Legal note at footer -->
      <div class="p-8 text-center text-md text-[var(--text-muted)] border-t border-[var(--border-main)] bg-[var(--bg-surface)]">
        <span>Keamanan terjamin · Aktivitas Onboarding login dicatat untuk audit kepatuhan.</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

definePageMeta({
  layout: 'auth',
  middleware: 'auth',
  pageTransition: { name: 'fade', mode: 'out-in' },
})

const authStore = useAuthStore()
const router = useRouter()

const currentStep = ref(2)
const scrollContainer = ref<HTMLElement | null>(null)
const scrollProgress = ref(0)
const hasScrolledToBottom = ref(false)
const accepting = ref(false)

// Step 3 Profile State
const profileState = reactive({
  fullName: '',
  phone: '',
  department: '',
})
const userRoles = ref<string[]>([])
const permissionsList = ref<string[]>([])
const savingProfile = ref(false)
const error = ref('')

// Initialize profile state from store
onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/auth/login')
    return
  }

  // Pre-populate values from JWT token / initial session data
  profileState.fullName = authStore.user?.fullName ?? ''
  profileState.phone = authStore.user?.phone ?? ''
  profileState.department = authStore.user?.department ?? ''
  userRoles.value = authStore.user?.roles ?? []

  // Load detailed profile including db roles & permissions
  const detailedUser = await authStore.fetchUserProfile()
  if (detailedUser) {
    profileState.fullName = detailedUser.full_name || profileState.fullName
    profileState.phone = detailedUser.phone || profileState.phone
    profileState.department = detailedUser.department || profileState.department
    userRoles.value = detailedUser.roles || userRoles.value
  }

  // Map permissions list dynamically based on roles
  permissionsList.value = getPermissionsForRoles(userRoles.value)
})

const userInitials = computed(() => {
  const name = profileState.fullName || authStore.user?.fullName || 'U'
  return name.split(' ').map(n => n[0]).join('').toUpperCase().substring(0, 2)
})

const currentDateTime = computed(() => {
  return new Date().toLocaleString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
})

const handleScroll = () => {
  const el = scrollContainer.value
  if (!el) return
  const progress = (el.scrollTop / (el.scrollHeight - el.clientHeight)) * 100
  scrollProgress.value = Math.min(100, Math.round(progress))
  if (scrollProgress.value >= 95) {
    hasScrolledToBottom.value = true
  }
}

const handleAccept = async () => {
  if (!hasScrolledToBottom.value || accepting.value) return
  accepting.value = true
  try {
    await authStore.acceptConfidentialityAgreement()
    // Go to next step (Profile Completion)
    currentStep.value = 3
  }
  catch {
    currentStep.value = 3
  }
  finally {
    accepting.value = false
  }
}

const handleSaveProfile = async () => {
  if (!profileState.fullName || savingProfile.value) return
  savingProfile.value = true
  error.value = ''

  try {
    await authStore.updateProfile({
      fullName: profileState.fullName,
      phone: profileState.phone,
      department: profileState.department,
    })
    // Go to dashboard
    router.push('/dashboard')
  }
  catch (err: any) {
    error.value = err.message || 'Gagal menyimpan profil. Silakan coba lagi.'
  }
  finally {
    savingProfile.value = false
  }
}

const handleReject = async () => {
  await authStore.logout()
}

// Helpers for RBAC Roles & Permissions mapping matching the database seed
const getPermissionsForRoles = (roles: string[]): string[] => {
  const adminPerms = [
    'user:create', 'user:read', 'user:update', 'user:delete',
    'audit:create', 'audit:read', 'audit:update', 'audit:delete',
    'risk:create', 'risk:read', 'risk:update',
  ]
  const auditorPerms = [
    'audit:create', 'audit:read', 'audit:update', 'risk:read',
  ]
  const viewerPerms = [
    'audit:read', 'risk:read',
  ]

  const perms = new Set<string>()
  for (const role of roles) {
    const roleLower = role.toLowerCase()
    if (roleLower === 'admin') adminPerms.forEach(p => perms.add(p))
    else if (roleLower === 'auditor') auditorPerms.forEach(p => perms.add(p))
    else if (roleLower === 'viewer') viewerPerms.forEach(p => perms.add(p))
  }
  return Array.from(perms)
}

const getPermissionLabel = (permission: string): string => {
  const labels: Record<string, string> = {
    'user:create': 'Membuat akun pengguna baru',
    'user:read': 'Melihat daftar dan detail pengguna',
    'user:update': 'Mengubah data pengguna dan peran',
    'user:delete': 'Menghapus/menonaktifkan akun pengguna',
    'audit:create': 'Membuat piagam audit dan surat tugas baru',
    'audit:read': 'Melihat dokumen dan laporan hasil audit',
    'audit:update': 'Melakukan penilaian kertas kerja dan rekomendasi',
    'audit:delete': 'Menghapus data program kerja audit',
    'risk:create': 'Mengisi profile risiko unit kerja',
    'risk:read': 'Melihat matriks penilaian risiko organisasi',
    'risk:update': 'Mengubah bobot mitigasi risiko',
  }
  return labels[permission] || permission
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.4s ease-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
