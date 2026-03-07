<template>
  <div class="min-h-screen bg-gray-100">
    <nav class="bg-white shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <h1 class="text-xl font-bold text-gray-900">RBIA System</h1>
            </div>
            <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
              <NuxtLink
                to="/dashboard"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
              >
                Dashboard
              </NuxtLink>
            </div>
          </div>
          <UDropdownMenu :items="items" class="my-4" variant="soft" size="lg">
            <UButton icon="i-lucide-menu" color="secondary" variant="outline"></UButton>
          </UDropdownMenu>
          <!-- <div class="hidden sm:ml-6 sm:flex sm:items-center">
            <div class="ml-3 relative flex items-center space-x-4">
              <button
                @click="toggleLocale"
                class="text-gray-500 hover:text-gray-700 px-3 py-2 rounded-md text-sm font-medium"
              >
                {{ locale === "en" ? "EN" : "ID" }}
              </button>
              <span class="text-gray-700">{{ authStore.user?.fullName }}</span>
              <button
                @click="handleLogout"
                class="text-gray-500 hover:text-gray-700 px-3 py-2 rounded-md text-sm font-medium"
              >
                Logout
              </button>
            </div>
          </div> -->
        </div>
      </div>
    </nav>
    <UMain class="bg">
      <slot />
    </UMain>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "~/stores/auth";
import { useI18n } from "~/composables/useI18n";
import type { DropdownMenuItem } from "@nuxt/ui";

const authStore = useAuthStore();
const { locale, setLocale } = useI18n();

const handleLogout = async () => {
  await authStore.logout();
};

const navigateSetting = () => {
  navigateTo("/settings");
}

const items = [
    {
      label: 'Language',
      icon: 'language',
      slot: "language" as const,
      children: [
        [
          {
            label: 'English',
            onClick: () => setLocale('en')
          },
          {
            label: 'Indonesia',
            onClick: () => setLocale('id')
          },
        ],
      ]
    },
    {
      label: 'Settings',
      icon: 'setting',
      to: "/settings"
    },
    {
      label: 'Sign Out',
      color: "error",
      icon: 'logout',
      onClick: handleLogout
    },
] satisfies DropdownMenuItem[]
</script>
