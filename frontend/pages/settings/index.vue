<template>
  <div class="min-h-screen bg-gray-50">
    <UDashboardGroup>
      <UDashboardSidebar>
        <template #header>
          <div class="px-4 py-4">
            <h2 class="text-lg font-semibold text-gray-900">Settings</h2>
          </div>
        </template>

        <UDashboardSidebarLinks :links="links" />

        <template #footer>
          <div class="px-4 py-4 border-t border-gray-200">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-primary-100 flex items-center justify-center">
                <span class="text-sm font-semibold text-primary-700">U</span>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900 truncate">User Name</p>
                <p class="text-xs text-gray-500 truncate">user@example.com</p>
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

        <div class="p-6">
          <!-- My Profile Section -->
          <div v-if="activeTab === 'profile'">
            <UCard>
              <template #header>
                <h3 class="text-lg font-semibold text-gray-900">My Profile</h3>
                <p class="text-sm text-gray-600">Manage your personal information</p>
              </template>

              <div class="space-y-6">
                <div class="flex items-center gap-6">
                  <div class="w-24 h-24 rounded-full bg-primary-100 flex items-center justify-center">
                    <span class="text-3xl font-semibold text-primary-700">U</span>
                  </div>
                  <div>
                    <UButton color="primary" variant="soft">Change Photo</UButton>
                    <p class="text-xs text-gray-500 mt-2">JPG, PNG or GIF. Max 2MB.</p>
                  </div>
                </div>

                <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                  <UFormField label="Full Name">
                    <UInput placeholder="Enter your full name" model-value="User Name" />
                  </UFormField>

                  <UFormField label="Email">
                    <UInput type="email" placeholder="Enter your email" model-value="user@example.com" />
                  </UFormField>

                  <UFormField label="Phone Number">
                    <UInput placeholder="Enter your phone number" />
                  </UFormField>

                  <UFormField label="Department">
                    <USelect placeholder="Select department" :options="departments" />
                  </UFormField>
                </div>

                <UFormField label="Bio">
                  <UTextarea placeholder="Tell us about yourself" :rows="3" />
                </UFormField>

                <div class="flex justify-end gap-3">
                  <UButton variant="outline">Cancel</UButton>
                  <UButton color="primary">Save Changes</UButton>
                </div>
              </div>
            </UCard>
          </div>

          <!-- Settings Section -->
          <div v-if="activeTab === 'settings'">
            <div class="space-y-6">
              <UCard>
                <template #header>
                  <h3 class="text-lg font-semibold text-gray-900">General Settings</h3>
                </template>

                <div class="space-y-4">
                  <div class="flex items-center justify-between">
                    <div>
                      <p class="font-medium text-gray-900">Email Notifications</p>
                      <p class="text-sm text-gray-500">Receive email updates about your activity</p>
                    </div>
                    <UToggle v-model="settings.emailNotifications" />
                  </div>

                  <div class="flex items-center justify-between">
                    <div>
                      <p class="font-medium text-gray-900">Push Notifications</p>
                      <p class="text-sm text-gray-500">Receive push notifications on your device</p>
                    </div>
                    <UToggle v-model="settings.pushNotifications" />
                  </div>

                  <div class="flex items-center justify-between">
                    <div>
                      <p class="font-medium text-gray-900">Dark Mode</p>
                      <p class="text-sm text-gray-500">Toggle dark mode theme</p>
                    </div>
                    <UToggle v-model="settings.darkMode" />
                  </div>
                </div>
              </UCard>

              <UCard>
                <template #header>
                  <h3 class="text-lg font-semibold text-gray-900">Language & Region</h3>
                </template>

                <div class="space-y-4">
                  <UFormField label="Language">
                    <USelect v-model="settings.language" :options="languages" />
                  </UFormField>

                  <UFormField label="Timezone">
                    <USelect v-model="settings.timezone" :options="timezones" />
                  </UFormField>
                </div>
              </UCard>
            </div>
          </div>

          <!-- Activity Section -->
          <div v-if="activeTab === 'activity'">
            <UCard>
              <template #header>
                <h3 class="text-lg font-semibold text-gray-900">Recent Activity</h3>
                <p class="text-sm text-gray-600">Your recent actions and logs</p>
              </template>

              <div class="space-y-4">
                <div v-for="activity in activities" :key="activity.id" class="flex items-start gap-4 p-4 rounded-lg border border-gray-200 hover:bg-gray-50">
                  <div class="rounded-full bg-primary-100 w-10 h-10 flex items-center justify-center shrink-0">
                    <UIcon :name="activity.icon" class="text-primary-600 size-5" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="font-medium text-gray-900">{{ activity.title }}</p>
                    <p class="text-sm text-gray-600">{{ activity.description }}</p>
                    <p class="text-xs text-gray-500 mt-1">{{ activity.time }}</p>
                  </div>
                </div>
              </div>
            </UCard>
          </div>

          <!-- Permissions Section -->
          <div v-if="activeTab === 'permissions'">
            <UCard>
              <template #header>
                <h3 class="text-lg font-semibold text-gray-900">Permissions</h3>
                <p class="text-sm text-gray-600">Manage your access rights and roles</p>
              </template>

              <div class="space-y-6">
                <div>
                  <h4 class="font-medium text-gray-900 mb-3">Current Role</h4>
                  <div class="flex items-center gap-3 p-4 rounded-lg bg-primary-50 border border-primary-200">
                    <div class="rounded-full bg-primary-100 w-10 h-10 flex items-center justify-center">
                      <UIcon name="shield" class="text-primary-600 size-5" />
                    </div>
                    <div>
                      <p class="font-medium text-primary-900">Audit Manager</p>
                      <p class="text-sm text-primary-700">Full access to audit planning and execution</p>
                    </div>
                  </div>
                </div>

                <div>
                  <h4 class="font-medium text-gray-900 mb-3">Access Rights</h4>
                  <div class="space-y-3">
                    <div v-for="permission in permissions" :key="permission.module" class="flex items-center justify-between p-3 rounded-lg border border-gray-200">
                      <div class="flex items-center gap-3">
                        <UIcon :name="permission.icon" class="text-gray-600 size-5" />
                        <div>
                          <p class="font-medium text-gray-900">{{ permission.module }}</p>
                          <p class="text-sm text-gray-600">{{ permission.description }}</p>
                        </div>
                      </div>
                      <UBadge :color="permission.access === 'Full' ? 'success' : 'warning'" variant="soft">
                        {{ permission.access }}
                      </UBadge>
                    </div>
                  </div>
                </div>
              </div>
            </UCard>
          </div>

          <!-- FAQ Section -->
          <div v-if="activeTab === 'faq'">
            <UCard>
              <template #header>
                <h3 class="text-lg font-semibold text-gray-900">Frequently Asked Questions</h3>
                <p class="text-sm text-gray-600">Find answers to common questions</p>
              </template>

              <div class="space-y-4">
                <UAccordion :items="faqs" multiple>
                  <template #default="{ item, open }">
                    <UButton color="neutral" variant="ghost" class="w-full justify-between">
                      <span class="font-medium">{{ item.label }}</span>
                      <UIcon :name="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-5" />
                    </UButton>
                  </template>
                </UAccordion>
              </div>
            </UCard>
          </div>
        </div>
      </UDashboardPanel>
    </UDashboardGroup>
  </div>
</template>

<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui';

definePageMeta({
  middleware: "auth",
});

const activeTab = ref("profile");

const links = [
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
] as NavigationMenuItem[];

const currentPageTitle = computed(() => {
  const titles = {
    profile: "My Profile",
    settings: "Settings",
    activity: "Activity",
    permissions: "Permissions",
    faq: "FAQ",
  };
  return titles[activeTab.value as keyof typeof titles] || "Settings";
});

const departments = [
  { label: "Internal Audit", value: "internal-audit" },
  { label: "Risk Management", value: "risk-management" },
  { label: "Compliance", value: "compliance" },
  { label: "Finance", value: "finance" },
];

const languages = [
  { label: "English", value: "en" },
  { label: "Indonesia", value: "id" },
];

const timezones = [
  { label: "UTC+7 (Jakarta)", value: "asia/jakarta" },
  { label: "UTC+8 (Singapore)", value: "asia/singapore" },
  { label: "UTC+0 (London)", value: "europe/london" },
];

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

const faqs = [
  {
    label: "How do I change my password?",
    content: "Go to Settings > Security and click on 'Change Password'. You'll need to enter your current password and then create a new one.",
  },
  {
    label: "How do I create a new risk profile?",
    content: "Navigate to Risk Management > Risk Profiles and click on the 'Add New Profile' button. Fill in the required fields including risk name, category, impact level, and probability.",
  },
  {
    label: "Can I export audit reports to PDF?",
    content: "Yes, you can export any audit report to PDF. Open the report you want to export and click on the 'Export' button in the top right corner, then select PDF format.",
  },
  {
    label: "How do I change the language?",
    content: "You can change the language from the Settings page or by using the language selector in the top navigation menu. Currently, we support English and Indonesian.",
  },
  {
    label: "What are the different risk levels?",
    content: "Risk levels are calculated based on impact and probability scores: Low (1-3), Low to Moderate (4-6), Moderate (7-9), Moderate to High (10-12), and High (13-15).",
  },
];
</script>
