<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { NavigationMenuItem } from '@nuxt/ui'
import { useAuthStore } from '~/stores/auth'

// Gunakan useRoute untuk mendapatkan URL saat ini
const route = useRoute()
const authStore = useAuthStore()

// 1. Simpan data menu dalam variabel mentah (raw data)
const rawItems: NavigationMenuItem[][] = [[
  {
    label: 'Dashboard',
    icon: 'i-lucide-users',
    to: '/dashboard' 
  },
  {
    label: '1. Audit Charter',
    icon: 'charter',
    to: '/audit-charter'
  }, 
  {
    label: '2. Risk Profile',
    icon: 'i-lucide-inbox',
    to: '/risk-profile',
    children: [
      {
        label: 'Corporate Risk Profile',
        icon: 'i-lucide-users',
        to: '/risk-profile',
      },
      {
        label: 'Risk Appetite Statement',
        icon: 'i-lucide-clipboard-check',
        to: '/risk-appetite',
      },
    ]
  }, 
  {
    label: '3. Strategic Audit Plan',
    icon: 'i-lucide-users',
    to: '/strategic-audit-plan',
    children: [
      {
        label: 'KPI Performance',
        icon: 'i-lucide-users',
        to: '/kpi-performance',
      },
    ]
  },
  {
    label: '4. Annual Audit Plan',
    icon: 'i-lucide-users',
    to: '/annual-audit',
    children: [
      {
        label: 'Audit Execution Status',
        icon: 'i-lucide-users',
        to: '/audit-execution-status',
      }
    ]
  },
  {
    label: '5. Audit Activity Plan',
    icon: 'i-lucide-users',
    to: '/audit-activity-plan',
    children: [
      {
        label: 'Create Activity Plan Document',
        icon: 'i-lucide-layout-dashboard',
        to: '/audit-activity-plan',
      },
      {
        label: 'Upload Activity Plan Document',
        icon: 'i-lucide-upload',
        to: '/audit-activity-plan/upload',
      }
    ]
  }, 
  {
    label: '6. Assignment Letter',
    icon: 'i-lucide-users',
    to: '/assignment-letter',
  },
  {
    label: '7. Audit Field Work',
    icon: 'i-lucide-users',
    to: '/audit-fieldwork',
    children: [
      {
        label: 'Create Working Paper',
        icon: 'i-lucide-file-plus',
        to: '/working-paper',
      },
      {
        label: 'Import Working Paper',
        icon: 'i-lucide-upload',
        to: '/import-working-paper',
      },
    ]
  },
  {
    label: '8. Audit Result Report',
    icon: 'i-lucide-users',
    to: '/audit-result-report',
  },
  {
    label: '9. Executive Summary Report',
    icon: 'i-lucide-file-text',
    to: '/executive-summary',
    children: [
      {
        label: 'Compilation & Narrative',
        icon: 'i-lucide-clipboard-list',
        to: '/executive-summary',
      },
      {
        label: 'Matriks Induk Temuan',
        icon: 'i-lucide-table',
        to: '/executive-summary/matriks',
      }
    ]
  },
  {
    label: '10. Action Taken Report',
    icon: 'i-lucide-users',
    to: '/action-taken-report',
  },
  {
    label: '11. Consulting Service',
    icon: 'i-lucide-users',
    to: '/consulting-service',
  },
  {
    label: '12. Quality Assurance Review',
    icon: 'i-lucide-shield-check',
    to: '/quality-assurance',
    children: [
      {
        label: 'Create QAR Report',
        icon: 'i-lucide-layout-dashboard',
        to: '/quality-assurance',
      },
      {
        label: 'Import QAR Report',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import',
      }
    ]
  },
  {
    label: 'Analytics',
    icon: 'i-lucide-pie-chart',
    to: '/analytics',
  },
  
]]

// 2. Gunakan Computed agar menu bereaksi setiap kali pengguna pindah halaman
const items = computed<NavigationMenuItem[][]>(() => {
  return rawItems.map(group => {
    return group.map(parent => {
      
      // Cek apakah ada submenu (child) yang URL-nya cocok dengan URL saat ini
      const hasActiveChild = parent.children?.some(child => 
        route.path.startsWith(child.to as string)
      )

      return {
        ...parent,
        // Jika ada child yang aktif, buat parent menyala (active: true) 
        // dan otomatis terbuka dropdown-nya (defaultOpen: true)
        active: hasActiveChild,
        defaultOpen: hasActiveChild || parent.defaultOpen
      }
    })
  })
})

const userDropdownItems = computed(() => [
  [
    {
      label: authStore.getUser?.fullName || 'User',
      slot: 'account',
      disabled: true
    }
  ],
  [
    {
      label: 'Settings',
      icon: 'i-lucide-settings',
      to: '/settings'
    }
  ],
  [
    {
      label: 'Logout',
      icon: 'i-lucide-log-out',
      onSelect: async () => {
        await authStore.logout()
      }
    }
  ]
])
</script>

<template>
  <UDashboardGroup>
  <UDashboardSidebar 
    collapsible 
    resizable
    :min-size="20"
    :default-size="25"
    :max-size="30"
    :collapsed-size="0"
    :ui="{ footer: 'border-t border-default' }"
  >
    <template #header="{ collapsed }">
      <Logo v-if="!collapsed" class="h-8 w-auto shrink-0" hide-subtitle text-class="text-xl" />
      <Logo v-else icon-only class="h-6 w-auto mx-auto" />
      <UDashboardSidebarCollapse variant="subtle" />
    </template>

    <template #default="{ collapsed }">
      <UButton
        :label="collapsed ? undefined : 'Search...'"
        icon="i-lucide-search"
        color="neutral"
        variant="outline"
        block
        :square="collapsed"
      >
        <template v-if="!collapsed" #trailing>
          <div class="flex items-center gap-0.5 ms-auto">
            
          </div>
        </template>
      </UButton>

      <UNavigationMenu
        :collapsed="collapsed"
        :items="items[0]"
        orientation="vertical"
      />

      <UNavigationMenu
        v-if="items[1]" 
        :collapsed="collapsed"
        :items="items[1]"
        orientation="vertical"
        class="mt-auto"
      />
    </template>
    
  </UDashboardSidebar>

  <UDashboardPanel>
    <template #header>
        <div class="flex items-center justify-between px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-main)]">
          <Logo class="h-6 w-auto mx-auto text-2xl" />
          <div class="flex items-center gap-4">
            <UColorModeButton />
            <UButton v-if="!authStore.isLoggedIn" to="/auth/login" color="primary" variant="solid">Login</UButton>
            <UDropdownMenu v-else :items="userDropdownItems">
              <UAvatar :alt="authStore.getUser?.fullName || 'User'" size="md" />
              
              <template #account-item="{ item }">
                <div class="text-left">
                  <p class="text-sm font-medium text-gray-900 truncate">
                    {{ item.label }}
                  </p>
                  <p class="text-xs text-gray-500 truncate">
                    {{ authStore.getUser?.email }}
                  </p>
                </div>
              </template>
            </UDropdownMenu>
          </div>
        </div>
    </template>
    <template #body>
      <div class="min-h-screen bg-[var(--bg-main)] min-w-max transition-colors duration-300">
        <div v-if="authStore.isNewDevice" class="max-w-7xl mx-auto p-4 pb-0">
          <UAlert
            icon="i-lucide-alert-triangle"
            color="warning"
            variant="solid"
            title="Security Warning"
            description="Your account was recently accessed from a new device. If this wasn't you, please change your password immediately."
            :closable="true"
            @close="authStore.isNewDevice = false"
          />
        </div>
        <UMain class="max-w-7xl mx-auto">
          <slot />
        </UMain>
      </div>
    </template>
  </UDashboardPanel>

  
  </UDashboardGroup>
</template>