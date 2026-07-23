<template>
  <div class="min-h-screen flex">
    <UDashboardGroup>
      <UDashboardSidebar resizeable>
        <template #header>
          <div class="px-4 py-4">
            <h2 class="text-lg font-semibold text-gray-900">Account</h2>
          </div>
        </template>

        <template #default>
          <!-- <UDashboardSidebarLinks :links="links" /> -->
          <UNavigationMenu
            color="primary"
            :items="links"
            orientation="vertical"
            aria-orientation="vertical"
            :highlight="true"
          />
        </template>

        <template #footer>
          <div class="px-4 py-4 border-t border-gray-200">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-primary-100 flex items-center justify-center">
                <span class="text-sm font-semibold text-primary-700">U</span>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900 truncate">User Name</p>
                <p class="text-md text-gray-500 truncate">user@example.com</p>
              </div>
            </div>
          </div>
        </template>
      </UDashboardSidebar>

      <UDashboardPanel>
        <template #header>
          <div class="px-6 py-4 border-b border-gray-200">
            <h1 class="text-xl font-semibold text-gray-900">{{ currentPageTitle }}</h1>
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
import type { NavigationMenuItem } from '@nuxt/ui';

definePageMeta({
  middleware: "auth",
  layout: "dashboard",
  layoutTransition: {
    name: "fade",
    mode: "in-out",
    type: "animation",
    duration: 500,
    appear: true,
  },
});

const activeTab = ref("profile");

const links = computed<NavigationMenuItem[]>(() => [
  {
    label: "My Profile",
    icon: "i-lucide-user",
    slot: "profile" as const,
    onClick: () => activeTab.value = "profile",
    active: activeTab.value === "profile",
  },
  {
    label: "Settings",
    icon: "i-lucide-settings",
    slot: "settings" as const,
    onClick: () => activeTab.value = "settings",
    active: activeTab.value === "settings",
  },
  {
    label: "Security (MFA)",
    icon: "i-lucide-shield-check",
    slot: "mfa" as const,
    onClick: () => activeTab.value = "mfa",
    active: activeTab.value === "mfa",
  },
  {
    label: "Activity",
    icon: "i-lucide-clock",
    slot: "activity" as const,
    onClick: () => activeTab.value = "activity",
    active: activeTab.value === "activity",
  },
  {
    label: "Permissions",
    icon: "i-lucide-shield",
    slot: "permissions" as const,
    onClick: () => activeTab.value = "permissions",
    active: activeTab.value === "permissions",
  },
  {
    label: "FAQ",
    icon: "i-lucide-help-circle",
    slot: "faq" as const,
    onClick: () => activeTab.value = "faq",
    active: activeTab.value === "faq",
  },
]);

const currentPageTitle = computed(() => {
  const titles = {
    profile: "My Profile",
    settings: "Settings",
    mfa: "Two-Factor Authentication",
    activity: "Activity",
    permissions: "Permissions",
    faq: "FAQ",
  };
  return titles[activeTab.value as keyof typeof titles] || "Settings";
});

const settings = ref({
  emailNotifications: true,
  pushNotifications: false,
  darkMode: false,
  language: "en",
  timezone: "asia/jakarta",
});

const activities = [
  {
    id: 1,
    title: "Created new risk profile",
    description: "Added operational risk for IT department",
    icon: "i-lucide-plus-circle",
    time: "2 hours ago",
  },
  {
    id: 2,
    title: "Updated audit plan",
    description: "Modified Q4 2025 audit schedule",
    icon: "i-lucide-edit",
    time: "5 hours ago",
  },
  {
    id: 3,
    title: "Completed review",
    description: "Finished review of financial audit findings",
    icon: "i-lucide-check-circle",
    time: "1 day ago",
  },
  {
    id: 4,
    title: "Logged in",
    description: "Login from Jakarta, Indonesia",
    icon: "i-lucide-log-in",
    time: "2 days ago",
  },
];

const permissions = [
  {
    module: "Risk Management",
    description: "Create, edit, and delete risk profiles",
    icon: "i-lucide-shield-alert",
    access: "Full",
  },
  {
    module: "Audit Planning",
    description: "Create and manage audit plans",
    icon: "i-lucide-calendar",
    access: "Full",
  },
  {
    module: "Reports",
    description: "View and generate reports",
    icon: "i-lucide-file-text",
    access: "Full",
  },
  {
    module: "User Management",
    description: "Manage user accounts and roles",
    icon: "i-lucide-users",
    access: "Read Only",
  },
];
</script>
