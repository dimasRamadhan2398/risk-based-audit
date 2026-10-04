<template>
  <div class="min-h-screen flex">
    <UDashboardGroup>
      <UDashboardSidebar 
        id="settings-sidebar"
        resizable
        :min-size="12"
        :default-size="17"
        :max-size="22"
      >
        <template #header>
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-white">{{ t('settings.sidebar.accountSettings') }}</h2>
          </div>
        </template>

        <template #default>
          <UNavigationMenu
            color="primary"
            :items="links"
            orientation="vertical"
            aria-orientation="vertical"
          />
        </template>

        <template #footer>
          <div class="w-full border-t border-gray-200 dark:border-gray-800 px-3 py-3">
            <div class="flex items-center gap-2.5 min-w-0">
              <div class="w-8 h-8 rounded-full bg-primary/20 text-primary flex items-center justify-center font-extrabold text-xs shrink-0 shadow-xs overflow-hidden">
                <img v-if="authStore.user?.avatarUrl" :src="authStore.user.avatarUrl" alt="" class="w-full h-full object-cover">
                <template v-else>
                  {{ userInitial }}
                </template>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-xs font-bold text-gray-900 dark:text-white truncate leading-tight" :title="authStore.user?.fullName || authStore.user?.username || 'User'">
                  {{ authStore.user?.fullName || authStore.user?.username || 'User' }}
                </p>
                <p class="text-[11px] text-gray-400 dark:text-gray-400 truncate leading-tight mt-0.5" :title="authStore.user?.email || 'user@example.com'">
                  {{ authStore.user?.email || 'user@example.com' }}
                </p>
              </div>
            </div>
          </div>
        </template>
      </UDashboardSidebar>

      <UDashboardPanel>
        <template #header>
          <div class="px-4 sm:px-6 py-3 sm:py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
            <div class="flex items-center gap-2 sm:gap-3 min-w-0">
              <UButton
                icon="i-lucide-arrow-left"
                color="neutral"
                variant="ghost"
                size="sm"
                class="rounded-xl font-semibold hover:bg-gray-100 dark:hover:bg-gray-800 shrink-0"
                @click="goBack"
              >
                {{ t('common.back') }}
              </UButton>
              <div class="h-4 w-px bg-gray-200 dark:bg-gray-700 shrink-0"></div>
              <h1 class="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white truncate">{{ currentPageTitle }}</h1>
            </div>
          </div>
          <!-- Mobile & Tablet Scrollable Tab Bar (visible on < lg when sidebar is hidden) -->
          <div class="lg:hidden border-b border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] px-3 py-2 overflow-x-auto flex items-center gap-1.5 scrollbar-none">
            <button
              v-for="link in links"
              :key="link.slot"
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold whitespace-nowrap transition-all duration-200 shrink-0"
              :class="activeTab === link.slot
                ? 'bg-primary-500 text-white shadow-xs'
                : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              @click="selectTab(link.slot as string)"
            >
              <UIcon :name="link.icon" class="size-4" />
              <span>{{ link.label }}</span>
            </button>
          </div>
        </template>

        <template #body>
          <div class="p-4 sm:p-6">
            <!-- My Profile Section -->
            <div v-if="activeTab === 'profile'">
              <SettingsProfile />
            </div>

            <!-- Settings Section -->
            <div v-if="activeTab === 'settings'">
              <SettingsGeneral />
            </div>

            <!-- Two-Factor Authentication (MFA) Section -->
            <div v-if="activeTab === 'mfa'">
              <SettingsMfa />
            </div>

            <!-- Activity Section -->
            <div v-if="activeTab === 'activity'">
              <SettingsActivity />
            </div>

            <!-- Permissions Section -->
            <div v-if="activeTab === 'permissions'">
              <SettingsPermission />
            </div>

            <!-- Data Sources Section -->
            <div v-if="activeTab === 'datasource'">
              <SettingsDataSource />
            </div>

            <!-- FAQ Section -->
            <div v-if="activeTab === 'faq'">
              <SettingsFaq />
            </div>
          </div>
        </template>
      </UDashboardPanel>
    </UDashboardGroup>
  </div>
</template>

<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { useAuthStore } from '~/stores/auth'

definePageMeta({
  middleware: 'auth',
  layout: 'dashboard',
  layoutTransition: {
    name: 'fade',
    mode: 'in-out',
    type: 'animation',
    duration: 500,
    appear: true,
  },
})

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const validTabs = ['profile', 'settings', 'mfa', 'activity', 'permissions', 'datasource', 'faq']
const initialTab = computed(() => {
  const tabQuery = route.query.tab as string
  return validTabs.includes(tabQuery) ? tabQuery : 'profile'
})

const activeTab = ref(initialTab.value)

watch(
  () => route.query.tab,
  (newTab) => {
    if (newTab && typeof newTab === 'string' && validTabs.includes(newTab)) {
      activeTab.value = newTab
    }
  }
)

const selectTab = (tab: string) => {
  activeTab.value = tab
  router.replace({ query: { ...route.query, tab } })
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

const userInitial = computed(() => {
  const name = authStore.user?.fullName || authStore.user?.username || 'U'
  return name.charAt(0).toUpperCase()
})

const links = computed<NavigationMenuItem[]>(() => [
  {
    label: t('settings.sidebar.myProfile'),
    icon: 'i-lucide-user',
    slot: 'profile' as const,
    onSelect: () => selectTab('profile'),
    onClick: () => selectTab('profile'),
    active: activeTab.value === 'profile',
  },
  {
    label: t('settings.sidebar.settings'),
    icon: 'i-lucide-settings',
    slot: 'settings' as const,
    onSelect: () => selectTab('settings'),
    onClick: () => selectTab('settings'),
    active: activeTab.value === 'settings',
  },
  {
    label: t('settings.sidebar.securityMfa'),
    icon: 'i-lucide-shield-check',
    slot: 'mfa' as const,
    onSelect: () => selectTab('mfa'),
    onClick: () => selectTab('mfa'),
    active: activeTab.value === 'mfa',
  },
  {
    label: t('settings.sidebar.activity'),
    icon: 'i-lucide-clock',
    slot: 'activity' as const,
    onSelect: () => selectTab('activity'),
    onClick: () => selectTab('activity'),
    active: activeTab.value === 'activity',
  },
  {
    label: t('settings.sidebar.permissions'),
    icon: 'i-lucide-shield',
    slot: 'permissions' as const,
    onSelect: () => selectTab('permissions'),
    onClick: () => selectTab('permissions'),
    active: activeTab.value === 'permissions',
  },
  {
    label: t('settings.sidebar.dataSources'),
    icon: 'i-lucide-database',
    slot: 'datasource' as const,
    onSelect: () => selectTab('datasource'),
    onClick: () => selectTab('datasource'),
    active: activeTab.value === 'datasource',
  },
  {
    label: t('settings.sidebar.faq'),
    icon: 'i-lucide-help-circle',
    slot: 'faq' as const,
    onSelect: () => selectTab('faq'),
    onClick: () => selectTab('faq'),
    active: activeTab.value === 'faq',
  },
])

const currentPageTitle = computed(() => {
  const titles: Record<string, string> = {
    profile: t('settings.titles.profile'),
    settings: t('settings.titles.settings'),
    mfa: t('settings.titles.mfa'),
    activity: t('settings.titles.activity'),
    permissions: t('settings.titles.permissions'),
    datasource: t('settings.titles.datasource'),
    faq: t('settings.titles.faq'),
  }
  return titles[activeTab.value] || t('settings.titles.settings')
})
</script>
