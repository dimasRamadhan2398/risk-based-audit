<template>
  <div class="min-h-screen flex">
    <UDashboardGroup>
      <UDashboardSidebar resizeable>
        <template #header>
          <div class="px-4 py-4">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('settings.sidebar.accountSettings') }}</h2>
          </div>
        </template>

        <template #default>
          <UNavigationMenu
            color="primary"
            :items="links"
            orientation="vertical"
            aria-orientation="vertical"
            :highlight="true"
          />
        </template>

        <template #footer>
          <div class="px-4 py-4 border-t border-gray-200 dark:border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center shrink-0">
                <span class="text-sm font-semibold text-primary-700 dark:text-primary-400">
                  {{ userInitial }}
                </span>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {{ authStore.user?.fullName || authStore.user?.username || 'User' }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400 truncate">
                  {{ authStore.user?.email || 'user@example.com' }}
                </p>
              </div>
            </div>
          </div>
        </template>
      </UDashboardSidebar>

      <UDashboardPanel>
        <template #header>
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ currentPageTitle }}</h1>
          </div>
        </template>

        <template #body>
          <div class="p-6">
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
