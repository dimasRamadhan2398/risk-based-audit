<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import type { NavigationMenuItem } from '@nuxt/ui'
import { useAuthStore } from '~/stores/auth'
import { triggerScrollReset } from '~/utils/scroll'

const route = useRoute()
const authStore = useAuthStore()

const isMobileMenuOpen = ref(false)

watch(() => route.fullPath, () => {
  isMobileMenuOpen.value = false
  triggerScrollReset()
})

const openMobileMenu = () => {
  isMobileMenuOpen.value = true
}

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false
}

// 1. Simpan data menu dalam variabel mentah (raw data)
const rawItems: NavigationMenuItem[][] = [[
  // 1. Dashboard
  {
    label: 'Dashboard',
    icon: 'i-lucide-layout-dashboard',
    to: '/dashboard'
  },

  // 2. Audit Charter
  {
    label: 'Audit Charter',
    icon: 'i-lucide-scroll-text',
    to: '/audit-charter'
  },

  // 3. Risk Profile & Audit Universe
  {
    label: 'Risk Profile & Audit Universe',
    icon: 'i-lucide-shield-alert',
    to: '/risk-profile',
    children: [
      {
        label: 'Corporate Risk Profile',
        icon: 'i-lucide-users',
        to: '/risk-profile'
      },
      {
        label: 'Risk Appetite Statement',
        icon: 'i-lucide-clipboard-check',
        to: '/risk-appetite'
      },
      {
        label: 'Risk Factors',
        icon: 'i-lucide-activity',
        to: '/risk-profile/risk-factors'
      },
      {
        label: 'Audit Universe',
        icon: 'i-lucide-globe',
        to: '/risk-profile/audit-universe'
      },
      {
        label: 'Audit Priority',
        icon: 'i-lucide-list-ordered',
        to: '/risk-profile/audit-universe?tab=priority'
      },
      {
        label: 'Risk Control Matrix',
        icon: 'i-lucide-grid',
        to: '/risk-profile/risk-control-matrix'
      }
    ]
  },

  // 4. Audit Planning
  {
    label: 'Audit Planning',
    icon: 'i-lucide-calendar-days',
    type: 'trigger',
    children: [
      {
        label: 'Strategic Audit Plan',
        icon: 'i-lucide-target',
        type: 'trigger',
        children: [
          {
            label: 'Strategic Audit Plan',
            icon: 'i-lucide-layout-dashboard',
            to: '/strategic-audit-plan'
          },
          {
            label: 'Import Strategic Audit Plan',
            icon: 'i-lucide-upload',
            to: '/strategic-audit-plan/upload'
          }
        ]
      },
      {
        label: 'Internal Audit Performance',
        icon: 'i-lucide-trending-up',
        type: 'trigger',
        children: [
          {
            label: 'Internal Audit Performance',
            icon: 'i-lucide-layout-dashboard',
            to: '/kpi-performance'
          },
          {
            label: 'Import Laporan Kinerja',
            icon: 'i-lucide-upload',
            to: '/kpi-performance/upload'
          }
        ]
      },
      {
        label: 'Annual Audit Plan',
        icon: 'i-lucide-calendar',
        type: 'trigger',
        children: [
          {
            label: 'Create Annual Audit Plan',
            icon: 'i-lucide-layout-dashboard',
            to: '/annual-audit'
          },
          {
            label: 'Import Annual Audit Plan',
            icon: 'i-lucide-upload',
            to: '/annual-audit/upload'
          },
          {
            label: 'Audit Execution Status',
            icon: 'i-lucide-check-circle-2',
            to: '/audit-execution-status'
          }
        ]
      },
      {
        label: 'Audit Activity Plan',
        icon: 'i-lucide-clipboard-list',
        type: 'trigger',
        children: [
          {
            label: 'Create Audit Activity Plan',
            icon: 'i-lucide-layout-dashboard',
            to: '/audit-activity-plan'
          },
          {
            label: 'Import Audit Activity Plan',
            icon: 'i-lucide-upload',
            to: '/audit-activity-plan/upload'
          }
        ]
      },
      {
        label: 'Assignment Letter',
        icon: 'i-lucide-file-signature',
        type: 'trigger',
        children: [
          {
            label: 'Create Assignment Letter',
            icon: 'i-lucide-layout-dashboard',
            to: '/assignment-letter'
          },
          {
            label: 'Import Assignment Letter Document',
            icon: 'i-lucide-upload',
            to: '/assignment-letter/upload'
          }
        ]
      },
      {
        label: 'Working Paper',
        icon: 'i-lucide-file-text',
        type: 'trigger',
        children: [
          {
            label: 'Create Working Paper',
            icon: 'i-lucide-file-plus',
            to: '/working-paper'
          },
          {
            label: 'Import Working Paper Document',
            icon: 'i-lucide-upload',
            to: '/working-paper/upload'
          }
        ]
      },
      {
        label: 'Audit Fieldwork',
        icon: 'i-lucide-briefcase',
        to: '/audit-fieldwork'
      },
      {
        label: 'AOI & RCA',
        icon: 'i-lucide-search-check',
        to: '/working-paper?step=f04'
      },
      {
        label: 'Action Plan',
        icon: 'i-lucide-list-todo',
        to: '/mitigation'
      }
    ]
  },

  // 5. Audit Result Report
  {
    label: 'Audit Result Report',
    icon: 'i-lucide-file-check-2',
    type: 'trigger',
    children: [
      {
        label: 'Result Reports (LHA)',
        icon: 'i-lucide-list',
        to: '/audit-result-report'
      },
      {
        label: 'Import LHA Document',
        icon: 'i-lucide-upload',
        to: '/audit-result-report/upload'
      },
      {
        label: 'Auto Generate Report',
        icon: 'i-lucide-sparkles',
        to: '/audit-result-report'
      },
      {
        label: 'Executive Summary',
        icon: 'i-lucide-presentation',
        type: 'trigger',
        children: [
          {
            label: 'Executive Summary',
            icon: 'i-lucide-layout-dashboard',
            to: '/audit-result-report/executive-summary'
          },
          {
            label: 'Import Executive Summary Document',
            icon: 'i-lucide-upload',
            to: '/audit-result-report/executive-summary-upload'
          }
        ]
      },
      {
        label: 'Executive Summary Report Kompilasi',
        icon: 'i-lucide-file-text',
        type: 'trigger',
        children: [
          {
            label: 'Executive Summary Report Kompilasi',
            icon: 'i-lucide-presentation',
            to: '/executive-summary'
          },
          {
            label: 'Import Executive Summary Report',
            icon: 'i-lucide-upload',
            to: '/executive-summary/upload'
          }
        ]
      },
      {
        label: 'Action Taken Report',
        icon: 'i-lucide-clipboard-check',
        to: '/action-taken-report'
      },
      {
        label: 'Client Satisfaction Survey',
        icon: 'i-lucide-smile',
        to: '/audit-result-report/satisfaction-survey'
      }
    ]
  },

  // 6. Consulting Service
  {
    label: 'Consulting Service',
    icon: 'i-lucide-messages-square',
    to: '/consulting-service',
    children: [
      {
        label: 'Consulting Service Dashboard',
        icon: 'i-lucide-layout-dashboard',
        to: '/consulting-service'
      },
      {
        label: 'Import Consulting Document',
        icon: 'i-lucide-upload',
        to: '/consulting-service/upload'
      }
    ]
  },

  // 7. Quality Assurance Review
  {
    label: 'Quality Assurance Review',
    icon: 'i-lucide-shield-check',
    to: '/quality-assurance',
    children: [
      {
        label: 'Quality Assurance Dashboard',
        icon: 'i-lucide-layout-dashboard',
        to: '/quality-assurance'
      },
      {
        label: 'Import Periodic Self Assessment',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-periodic-self-assessment'
      },
      {
        label: 'Import SAIV',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-saiv'
      },
      {
        label: 'Import QAR Report',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import'
      },
      {
        label: 'Import IACM',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-iacm'
      }
    ]
  },

  // 8. Analytics
  {
    label: 'Analytics',
    icon: 'i-lucide-pie-chart',
    to: '/analytics',
    children: [
      {
        label: 'Risk Scoring Prediction',
        icon: 'i-lucide-binary',
        to: '/analytics?tab=xgboost'
      },
      {
        label: 'Anomaly Detection',
        icon: 'i-lucide-shield-alert',
        to: '/analytics?tab=isolation'
      },
      {
        label: 'Detected Anomalies Detail',
        icon: 'i-lucide-table-properties',
        to: '/analytics?tab=isolation'
      },
      {
        label: 'NLP Analysis',
        icon: 'i-lucide-file-search',
        to: '/analytics?tab=nlp'
      },
      {
        label: 'KPI Forecast',
        icon: 'i-lucide-trending-up',
        to: '/analytics/kpi-forecast'
      }
    ]
  },

  // Master Data
  {
    label: 'Master Data',
    icon: 'i-lucide-database',
    to: '/master/employee',
    children: [
      {
        label: 'Employee Management',
        icon: 'i-lucide-users-round',
        to: '/master/employee'
      },
      {
        label: 'Department',
        icon: 'i-lucide-building-2',
        to: '/master/department'
      }
    ]
  },

  // Settings
  {
    label: 'Settings',
    icon: 'i-lucide-settings',
    to: '/settings',
    children: [
      {
        label: 'General Settings',
        icon: 'i-lucide-sliders',
        to: '/settings'
      },
      {
        label: 'Security & 2FA',
        icon: 'i-lucide-shield-check',
        to: '/settings/mfa'
      },
      {
        label: 'Trusted Devices',
        icon: 'i-lucide-laptop',
        to: '/settings/devices'
      }
    ]
  }
]]

const searchQuery = ref('')
const { isAdmin, canManageAudits } = useRbac()

const isPathActive = (targetPath?: string): boolean => {
  if (!targetPath) return false
  if (targetPath.includes('?')) {
    return route.fullPath === targetPath
  }
  const [basePath] = targetPath.split('?')
  if (basePath === '/dashboard') return route.path === '/dashboard'
  return route.path === basePath || route.path.startsWith(basePath + '/')
}

const checkItemActive = (item: NavigationMenuItem): boolean => {
  if (isPathActive(item.to as string)) return true
  if (item.children && item.children.length > 0) {
    return item.children.some(checkItemActive)
  }
  return false
}

const processMenuItem = (item: NavigationMenuItem, q: string): NavigationMenuItem | null => {
  // Master data permission check
  if (item.to === '/master/employee' && !isAdmin.value && !canManageAudits.value) {
    return null
  }

  let processedChildren: NavigationMenuItem[] | undefined = undefined
  if (item.children && item.children.length > 0) {
    const list = item.children
      .map(child => processMenuItem(child, q))
      .filter((c): c is NavigationMenuItem => c !== null)
    if (list.length > 0) {
      processedChildren = list
    }
  }

  const labelMatches = !q || (item.label?.toLowerCase() || '').includes(q)
  const hasMatchingChildren = Boolean(processedChildren && processedChildren.length > 0)

  if (q && !labelMatches && !hasMatchingChildren) {
    return null
  }

  const isCurrentActive = isPathActive(item.to as string)
  const hasActiveChild = Boolean(
    processedChildren?.some(child => child.active || (child.children && child.children.some(checkItemActive)))
  )

  return {
    ...item,
    children: processedChildren,
    active: isCurrentActive || hasActiveChild,
    defaultOpen: hasActiveChild || Boolean(item.defaultOpen) || (Boolean(q) && hasMatchingChildren)
  }
}

// 2. Gunakan Computed agar menu bereaksi secara rekursif saat pindah halaman dan pencarian
const items = computed<NavigationMenuItem[][]>(() => {
  const q = searchQuery.value.toLowerCase().trim()
  return rawItems.map((group) => {
    return group
      .map(item => processMenuItem(item, q))
      .filter((item): item is NavigationMenuItem => item !== null)
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
      id="main-sidebar-final"
      collapsible
      resizable
      :min-size="21"
      :default-size="21"
      :max-size="35"
      :collapsed-size="0"
    >
      <template #header="{ collapsed }">
        <Logo
          v-if="!collapsed"
          class="h-7 w-auto shrink-0"
          hide-subtitle
          text-class="text-xl"
        />
        <Logo
          v-else
          icon-only
          class="h-6 w-auto mx-auto"
        />
        <UDashboardSidebarCollapse variant="subtle" />
      </template>

      <template #default="{ collapsed }">
        <div class="flex flex-col gap-3">
          <UInput
            v-if="!collapsed"
            v-model="searchQuery"
            placeholder="Search..."
            icon="i-lucide-search"
            color="neutral"
            variant="outline"
            class="w-full"
          />

          <UNavigationMenu
            :collapsed="collapsed"
            :items="items[0]"
            orientation="vertical"
            class="w-full"
            :ui="{
              childList: 'ps-3 border-l border-[var(--border-main)] ml-2.5 my-0.5 space-y-0.5'
            }"
          />
        </div>

        <UNavigationMenu
          v-if="items[1]"
          :collapsed="collapsed"
          :items="items[1]"
          orientation="vertical"
          class="mt-auto pt-3 border-t border-[var(--border-main)]"
        />
      </template>
    </UDashboardSidebar>

    <UDashboardPanel>
      <template #header>
        <div class="flex items-center justify-between px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-main)]">
          <div class="flex items-center gap-3">
            <UButton
              icon="i-lucide-menu"
              color="neutral"
              variant="ghost"
              class="lg:hidden"
              @click="openMobileMenu"
            />
            <Logo class="h-6 w-auto text-2xl" />
          </div>
          <div class="flex items-center gap-4">
            <UColorModeButton />
            <UButton
              v-if="!authStore.isLoggedIn"
              to="/auth/login"
              color="primary"
              variant="solid"
            >
              Login
            </UButton>
            <UDropdownMenu
              v-else
              :items="userDropdownItems"
            >
              <UAvatar
                :alt="authStore.getUser?.fullName || 'User'"
                size="md"
              />

              <template #account-item="{ item }">
                <div class="text-left">
                  <p class="text-sm font-medium text-gray-900 truncate">
                    {{ item.label }}
                  </p>
                  <p class="text-md text-gray-500 truncate">
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
          <div
            v-if="authStore.isNewDevice"
            class="max-w-7xl mx-auto p-4 pb-0"
          >
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

    <!-- Mobile Navigation Slideover -->
    <USlideover
      v-model:open="isMobileMenuOpen"
      side="left"
    >
      <template #content>
        <div class="flex flex-col h-full bg-[var(--bg-main)] border-r border-[var(--border-main)] overflow-y-auto w-72">
          <div class="flex items-center justify-between p-4 border-b border-[var(--border-main)]">
            <Logo class="h-6 w-auto text-xl" />
            <UButton
              icon="i-lucide-x"
              color="neutral"
              variant="ghost"
              size="sm"
              @click="closeMobileMenu"
            />
          </div>
          <div class="flex-1 p-4 space-y-4">
            <UNavigationMenu
              :items="items[0]"
              orientation="vertical"
              class="w-full"
              :ui="{
                childList: 'ps-3 border-l border-[var(--border-main)] ml-2.5 my-0.5 space-y-0.5'
              }"
            />
          </div>
        </div>
      </template>
    </USlideover>
  </UDashboardGroup>
</template>
