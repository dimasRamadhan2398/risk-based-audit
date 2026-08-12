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
            <h4 class="text-sm font-bold text-[var(--text-main)]">{{ t.stepper.step1Title }}</h4>
            <p class="text-md text-[var(--text-muted)]">{{ t.stepper.step1Desc }}</p>
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
            >{{ t.stepper.step2Title }}</h4>
            <p class="text-md text-[var(--text-muted)]">{{ t.stepper.step2Desc }}</p>
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
            >{{ t.stepper.step3Title }}</h4>
            <p class="text-md text-[var(--text-muted)]">{{ t.stepper.step3Desc }}</p>
          </div>
        </div>

        <div class="flex items-start gap-4">
          <span class="w-6 h-6 rounded-full bg-[var(--border-main)] text-[var(--text-muted)] flex items-center justify-center font-bold text-md">4</span>
          <div>
            <h4 class="text-sm font-bold text-[var(--text-main)]">{{ t.stepper.step4Title }}</h4>
            <p class="text-md text-[var(--text-muted)]">{{ t.stepper.step4Desc }}</p>
          </div>
        </div>
      </div>

      <!-- Bottom metadata -->
      <div class="relative z-10 text-md text-[var(--text-muted)] opacity-60">
        <span>{{ t.stepper.systemVersion }}</span>
      </div>
    </div>

    <!-- Right panel: Full screen content area -->
    <div class="w-full lg:w-8/12 flex flex-col justify-between bg-[var(--bg-main)] overflow-y-auto min-h-screen">
      <!-- Top Bar with Language Switcher Toggle -->
      <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] flex items-center justify-between select-none">
        <div class="flex items-center gap-2">
          <span class="text-xs font-semibold text-primary-500 lg:hidden">{{ t.stepper.mobileStep }} {{ currentStep }} {{ t.stepper.mobileOf }}</span>
          <span class="hidden sm:inline text-xs text-[var(--text-muted)] font-medium lg:hidden">· {{ t.stepper.mobileBannerDesc }}</span>
        </div>

        <!-- Language Switcher Toggle -->
        <div class="flex items-center gap-1 bg-[var(--bg-main)] border border-[var(--border-main)] rounded-xl p-1 shadow-xs ml-auto">
          <button
            type="button"
            class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all duration-200 flex items-center gap-1.5"
            :class="locale === 'id' ? 'bg-primary-500 text-white shadow-xs' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
            @click="locale = 'id'"
          >
            <span>🇮🇩</span> ID
          </button>
          <button
            type="button"
            class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all duration-200 flex items-center gap-1.5"
            :class="locale === 'en' ? 'bg-primary-500 text-white shadow-xs' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
            @click="locale = 'en'"
          >
            <span>🇬🇧</span> EN
          </button>
        </div>
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
              <h1 class="text-2xl font-bold text-[var(--text-main)]">{{ t.step2.title }}</h1>
              <p class="text-sm text-[var(--text-muted)]">{{ t.step2.subtitle }}</p>
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
              {{ t.step2.docHeaderTitle }}<br>
              {{ t.step2.docHeaderSub }}
            </p>

            <p>
              {{ t.step2.docOpening }}
            </p>

            <ol class="list-decimal list-inside space-y-3 pl-2">
              <li v-for="(point, idx) in t.step2.points" :key="idx">
                <strong class="text-[var(--text-main)] dark:text-white">{{ point.title }} </strong>
                {{ point.text }}
              </li>
            </ol>

            <p class="border-t border-[var(--border-main)] pt-4 text-md text-[var(--text-muted)]">
              {{ t.step2.docClosing }}
            </p>
          </div>

          <!-- Scroll progress indicator -->
          <div class="space-y-2">
            <div class="h-1.5 bg-[var(--border-main)] rounded-full overflow-hidden">
              <div
                class="h-full rounded-full transition-all duration-500 bg-primary-400"
                :style="{ width: `${scrollProgress}%` }"
              />
            </div>
            <p v-if="!hasScrolledToBottom" class="text-md text-[var(--text-muted)] text-center flex items-center justify-center gap-1">
              <span>↓</span> {{ t.step2.scrollToRead }}
            </p>
            <p v-else class="text-md text-success-600 dark:text-success-400 text-center flex items-center justify-center gap-1 font-medium">
              <span>✓</span> {{ t.step2.hasRead }}
            </p>
          </div>

          <!-- Action buttons for Step 2 -->
          <div class="flex flex-col sm:flex-row gap-3 pt-4">
            <DangerButton
              id="confidentiality-reject-btn"
              variant="solid"
              size="lg"
              class="whitespace-nowrap shrink-0"
              @click="handleReject"
            >
              {{ t.step2.rejectBtn }}
            </DangerButton>
            <button
              id="confidentiality-accept-btn"
              :disabled="!hasScrolledToBottom || accepting"
              class="flex-1 px-6 py-3 rounded-xl font-bold text-sm transition-all duration-200 whitespace-nowrap flex items-center justify-center gap-1.5"
              :class="hasScrolledToBottom && !accepting
                ? 'bg-primary-400 hover:bg-primary-500 text-white shadow-lg shadow-primary-400/25'
                : 'bg-[var(--border-main)] text-[var(--text-muted)] cursor-not-allowed'"
              @click="handleAccept"
            >
              <span v-if="accepting">{{ t.step2.acceptingBtn }}</span>
              <span v-else>{{ t.step2.acceptBtn }}</span>
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
              <h1 class="text-2xl font-bold text-[var(--text-main)]">{{ t.step3.title }}</h1>
              <p class="text-sm text-[var(--text-muted)]">{{ t.step3.subtitle }}</p>
            </div>
          </div>

          <!-- Form inputs -->
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField name="fullName">
                <template #label>
                  <span class="text-md font-semibold text-[var(--text-main)]">{{ t.step3.fullNameLabel }}</span>
                </template>
                <UInput
                  v-model="profileState.fullName"
                  :placeholder="t.step3.fullNamePlaceholder"
                  class="w-full"
                  :ui="{
                    base: 'bg-[var(--bg-surface)] border-[var(--border-main)] text-[var(--text-main)] rounded-xl py-3 px-4',
                  }"
                />
              </UFormField>

              <UFormField name="phone">
                <template #label>
                  <span class="text-md font-semibold text-[var(--text-main)]">{{ t.step3.phoneLabel }}</span>
                </template>
                <UInput
                  v-model="profileState.phone"
                  :placeholder="t.step3.phonePlaceholder"
                  class="w-full"
                  :ui="{
                    base: 'bg-[var(--bg-surface)] border-[var(--border-main)] text-[var(--text-main)] rounded-xl py-3 px-4',
                  }"
                />
              </UFormField>

              <UFormField name="department" class="md:col-span-2">
                <template #label>
                  <span class="text-md font-semibold text-[var(--text-main)]">{{ t.step3.deptLabel }}</span>
                </template>
                <UInput
                  v-model="profileState.department"
                  :placeholder="t.step3.deptPlaceholder"
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
              <span class="text-md font-semibold text-[var(--text-main)] block mb-2">{{ t.step3.rolesLabel }}</span>
              <div class="flex flex-wrap gap-2">
                <span v-for="role in userRoles" :key="role" class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-md font-bold bg-secondary-500/10 border border-secondary-500/20 text-secondary-600 dark:text-secondary-400 capitalize">
                  <UIcon name="i-lucide-key-round" class="w-3 h-3" /> {{ role }}
                </span>
              </div>
            </div>

            <div>
              <span class="text-md font-semibold text-[var(--text-main)] block mb-2">{{ t.step3.permissionsLabel }}</span>
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
              class="px-5 py-3 rounded-xl border border-[var(--border-main)] hover:bg-[var(--bg-surface)] text-sm font-semibold transition-all duration-200 text-[var(--text-main)] whitespace-nowrap shrink-0"
              @click="currentStep = 2"
            >
              {{ t.step3.backBtn }}
            </button>
            <button
              id="profile-complete-btn"
              :disabled="savingProfile || !profileState.fullName"
              class="flex-1 px-6 py-3 rounded-xl font-bold text-sm transition-all duration-200 whitespace-nowrap flex items-center justify-center gap-1.5"
              :class="!(savingProfile || !profileState.fullName)
                ? 'bg-primary-400 hover:bg-primary-500 text-white shadow-lg shadow-primary-400/25'
                : 'bg-[var(--border-main)] text-[var(--text-muted)] cursor-not-allowed'"
              @click="handleSaveProfile"
            >
              <span v-if="savingProfile">{{ t.step3.savingBtn }}</span>
              <span v-else>{{ t.step3.saveBtn }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Legal note at footer -->
      <div class="p-8 text-center text-md text-[var(--text-muted)] border-t border-[var(--border-main)] bg-[var(--bg-surface)]">
        <span>{{ t.footerNote }}</span>
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

const locale = ref<'id' | 'en'>('id')
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

const t = computed(() => {
  if (locale.value === 'en') {
    return {
      stepper: {
        step1Title: 'Authentication',
        step1Desc: 'Login credentials verified successfully',
        step2Title: 'Integrity Pact',
        step2Desc: 'Data confidentiality agreement',
        step3Title: 'Complete Profile & Roles',
        step3Desc: 'Profile information & access rights review (RBAC)',
        step4Title: 'Complete',
        step4Desc: 'Enter main system dashboard',
        systemVersion: 'Risk-Based Internal Audit System v1.0',
        mobileStep: 'Step',
        mobileOf: 'of 3',
        mobileBannerDesc: 'Integrity Pact & Profile',
      },
      step2: {
        title: 'Integrity Pact',
        subtitle: 'Mandatory agreement before using the system',
        docHeaderTitle: 'INTEGRITY PACT',
        docHeaderSub: 'RISK-BASED INTERNAL AUDIT SYSTEM (RBIA)',
        docOpening: 'I, the undersigned, hereby solemnly declare and promise that:',
        points: [
          {
            title: 'Information Confidentiality:',
            text: 'I will maintain the strict confidentiality of all information, data, and documents accessed through this system. I will not disclose, distribute, or use such information for personal gain or unauthorized third parties.'
          },
          {
            title: 'Authorized Usage:',
            text: 'I will only use this system for legitimate business purposes in accordance with my assigned duties and authorization granted by the organization.'
          },
          {
            title: 'Data Integrity:',
            text: 'I will not manipulate, falsify, or delete data and information within this system without valid authorization from designated authorities.'
          },
          {
            title: 'Account Security:',
            text: 'I assume full responsibility for the security of my access credentials (username and password). I will not share access credentials and will immediately report any unauthorized activity.'
          },
          {
            title: 'Regulatory Compliance:',
            text: 'I will comply with all applicable laws and regulations, including but not limited to personal data protection policies and corporate data confidentiality rules.'
          },
          {
            title: 'Violation Reporting:',
            text: 'I will promptly report any actual or potential breach of these terms to my supervisor or designated authority.'
          },
          {
            title: 'Sanctions & Penalties:',
            text: 'I understand that any violation of this integrity pact may result in administrative, disciplinary, and/or legal sanctions in accordance with applicable laws.'
          }
        ],
        docClosing: 'By clicking "I Agree & Continue", I declare that I have read, understood, and agreed to all terms in this Integrity Pact. This agreement is legally binding and will be recorded in the audit system with timestamp, IP address, and device metadata.',
        scrollToRead: 'Scroll down to read all terms',
        hasRead: 'You have read all terms',
        rejectBtn: '✕ Decline & Logout',
        acceptingBtn: 'Processing...',
        acceptBtn: '✓ I Agree & Continue',
      },
      step3: {
        title: 'Complete Profile & Roles',
        subtitle: 'Customize your profile and review your assigned system permissions',
        fullNameLabel: 'Full Name',
        fullNamePlaceholder: 'Enter full name',
        phoneLabel: 'Phone Number',
        phonePlaceholder: 'Enter phone number',
        deptLabel: 'Department / Work Unit',
        deptPlaceholder: 'Enter work unit',
        rolesLabel: 'Your Roles (Database Defaults)',
        permissionsLabel: 'Your System Permissions List:',
        backBtn: '← Back',
        savingBtn: 'Saving...',
        saveBtn: 'Save & Proceed to Dashboard →',
        errorSave: 'Failed to save profile. Please try again.',
      },
      footerNote: 'Secured · Onboarding login activity is logged for compliance audit.',
    }
  }

  return {
    stepper: {
      step1Title: 'Autentikasi',
      step1Desc: 'Kredensial login berhasil diverifikasi',
      step2Title: 'Pakta Integritas',
      step2Desc: 'Persetujuan ketentuan kerahasiaan data',
      step3Title: 'Lengkapi Profil & Peran',
      step3Desc: 'Informasi profil & tinjauan hak akses (RBAC)',
      step4Title: 'Selesai',
      step4Desc: 'Masuk ke dasbor sistem utama',
      systemVersion: 'Sistem Audit Internal Berbasis Risiko v1.0',
      mobileStep: 'Langkah',
      mobileOf: 'dari 3',
      mobileBannerDesc: 'Pakta Integritas & Profil',
    },
    step2: {
      title: 'Pakta Integritas',
      subtitle: 'Wajib disetujui sebelum menggunakan sistem',
      docHeaderTitle: 'PAKTA INTEGRITAS',
      docHeaderSub: 'SISTEM AUDIT INTERNAL BERBASIS RISIKO (RBIA)',
      docOpening: 'Saya yang bertanda tangan di bawah ini, dengan ini menyatakan dan berjanji dengan sungguh-sungguh bahwa:',
      points: [
        {
          title: 'Kerahasiaan Informasi:',
          text: 'Saya akan menjaga kerahasiaan seluruh informasi, data, dan dokumen yang saya akses melalui sistem ini. Saya tidak akan mengungkapkan, menyebarkan, atau menggunakan informasi tersebut untuk kepentingan pribadi atau pihak lain yang tidak berwenang.'
        },
        {
          title: 'Penggunaan yang Sah:',
          text: 'Saya hanya akan menggunakan sistem ini untuk keperluan pekerjaan yang sah dan sesuai dengan tugas dan wewenang yang diberikan kepada saya oleh organisasi.'
        },
        {
          title: 'Integritas Data:',
          text: 'Saya tidak akan memanipulasi, memalsukan, atau menghapus data dan informasi dalam sistem ini tanpa izin yang sah dari pihak yang berwenang.'
        },
        {
          title: 'Keamanan Akun:',
          text: 'Saya bertanggung jawab penuh atas keamanan kredensial akses saya (username dan password). Saya tidak akan berbagi akses saya kepada pihak lain dan akan segera melaporkan jika terjadi penyalahgunaan.'
        },
        {
          title: 'Kepatuhan Regulasi:',
          text: 'Saya akan mematuhi seluruh peraturan perundang-undangan yang berlaku, termasuk namun tidak terbatas pada ketentuan mengenai perlindungan data pribadi dan kerahasiaan informasi perusahaan.'
        },
        {
          title: 'Pelaporan Pelanggaran:',
          text: 'Saya akan melaporkan setiap pelanggaran atau potensi pelanggaran atas ketentuan ini kepada atasan atau pihak yang berwenang sesegera mungkin.'
        },
        {
          title: 'Sanksi:',
          text: 'Saya memahami bahwa pelanggaran atas pakta integritas ini dapat mengakibatkan sanksi administratif, disiplin, dan/atau sanksi hukum sesuai dengan ketentuan yang berlaku.'
        }
      ],
      docClosing: 'Dengan mengklik tombol "Saya Setuju & Lanjutkan", saya menyatakan bahwa saya telah membaca, memahami, dan menyetujui seluruh ketentuan dalam Pakta Integritas ini. Persetujuan ini bersifat mengikat secara hukum dan akan dicatat dalam sistem beserta timestamp, alamat IP, dan informasi perangkat yang digunakan.',
      scrollToRead: 'Gulir ke bawah untuk membaca seluruh ketentuan',
      hasRead: 'Anda telah membaca seluruh ketentuan',
      rejectBtn: '✕ Tolak & Keluar',
      acceptingBtn: 'Memproses...',
      acceptBtn: '✓ Saya Setuju & Lanjutkan',
    },
    step3: {
      title: 'Lengkapi Profil & Peran',
      subtitle: 'Sesuaikan profil dan tinjauan hak akses Anda di sistem',
      fullNameLabel: 'Nama Lengkap',
      fullNamePlaceholder: 'Masukkan nama lengkap',
      phoneLabel: 'Nomor Telepon',
      phonePlaceholder: 'Masukkan nomor telepon',
      deptLabel: 'Unit Kerja / Departemen',
      deptPlaceholder: 'Masukkan unit kerja',
      rolesLabel: 'Peran Anda (Default Database)',
      permissionsLabel: 'Daftar Hak Akses Sistem Anda:',
      backBtn: '← Kembali',
      savingBtn: 'Menyimpan...',
      saveBtn: 'Simpan & Masuk ke Dashboard →',
      errorSave: 'Gagal menyimpan profil. Silakan coba lagi.',
    },
    footerNote: 'Keamanan terjamin · Aktivitas Onboarding login dicatat untuk audit kepatuhan.',
  }
})

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
  const localeFormat = locale.value === 'en' ? 'en-US' : 'id-ID'
  return new Date().toLocaleString(localeFormat, {
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
    error.value = err.message || t.value.step3.errorSave
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
  const labelsID: Record<string, string> = {
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
  const labelsEN: Record<string, string> = {
    'user:create': 'Create new user accounts',
    'user:read': 'View user list and details',
    'user:update': 'Modify user data and roles',
    'user:delete': 'Delete/deactivate user accounts',
    'audit:create': 'Create audit charters and assignment letters',
    'audit:read': 'View audit documents and result reports',
    'audit:update': 'Perform working paper reviews & recommendations',
    'audit:delete': 'Delete audit work program data',
    'risk:create': 'Fill in work unit risk profiles',
    'risk:read': 'View organizational risk assessment matrix',
    'risk:update': 'Modify risk mitigation weights',
  }

  const map = locale.value === 'en' ? labelsEN : labelsID
  return map[permission] || permission
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
