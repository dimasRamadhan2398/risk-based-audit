<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import type { NavigationMenuItem } from '@nuxt/ui'
import { useAuthStore } from '~/stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const isMobileMenuOpen = ref(false)

watch(() => route.fullPath, () => {
  isMobileMenuOpen.value = false
})

const openMobileMenu = () => {
  isMobileMenuOpen.value = true
}

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false
}

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
      {
        label: 'Risk Factors',
        icon: 'i-lucide-activity',
        to: '/risk-profile/risk-factors',
      },
      {
        label: 'Audit Universe',
        icon: 'i-lucide-globe',
        to: '/risk-profile/audit-universe',
      },
      {
        label: 'Risk Control Matrix',
        icon: 'i-lucide-grid',
        to: '/risk-profile/risk-control-matrix',
      },
    ]
  }, 
  {
    label: '3. Strategic Audit Plan',
    icon: 'i-lucide-users',
    to: '/strategic-audit-plan',
    children: [
      {
        label: 'Create Strategic Audit Plan',
        icon: 'i-lucide-layout-dashboard',
        to: '/strategic-audit-plan',
      },
      {
        label: 'Import Strategic Audit Plan',
        icon: 'i-lucide-upload',
        to: '/strategic-audit-plan/upload',
      }
    ]
  },
  {
    label: '4. Annual Audit Plan',
    icon: 'i-lucide-users',
    to: '/annual-audit',
    children: [
      {
        label: 'Create Annual Audit Plan',
        icon: 'i-lucide-layout-dashboard',
        to: '/annual-audit',
      },
      {
        label: 'Import Annual Audit Plan Document',
        icon: 'i-lucide-upload',
        to: '/annual-audit/upload',
      },
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
        label: 'Import Activity Plan Document',
        icon: 'i-lucide-upload',
        to: '/audit-activity-plan/upload',
      }
    ]
  }, 
  {
    label: '6. Assignment Letter',
    icon: 'i-lucide-users',
    to: '/assignment-letter',
    children: [
      {
        label: 'Create Assignment Letter',
        icon: 'i-lucide-layout-dashboard',
        to: '/assignment-letter',
      },
      {
        label: 'Import Assignment Letter Document',
        icon: 'i-lucide-upload',
        to: '/assignment-letter/upload',
      }
    ]
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
        label: 'Import Working Paper Document',
        icon: 'i-lucide-upload',
        to: '/working-paper/upload',
      },
    ]
  },
  {
    label: '8. Audit Result Report',
    icon: 'i-lucide-file-text',
    to: '/audit-result-report',
    children: [
      {
        label: 'Result Reports (LHA)',
        icon: 'i-lucide-list',
        to: '/audit-result-report',
      },
      {
        label: 'Import LHA Document',
        icon: 'i-lucide-upload',
        to: '/audit-result-report/upload',
      },
      {
        label: 'Executive Summary',
        icon: 'i-lucide-presentation',
        to: '/audit-result-report/executive-summary',
      },
      {
        label: 'Import Executive Summary Document',
        icon: 'i-lucide-upload',
        to: '/audit-result-report/executive-summary-upload',
      },
      {
        label: 'Client Satisfaction Survey',
        icon: 'i-lucide-smile',
        to: '/audit-result-report/satisfaction-survey',
      }
    ]
  },
  {
    label: '9. Executive Summary Report',
    icon: 'i-lucide-file-text',
    to: '/executive-summary',
    children: [
      {
        label: 'Executive Summary Report Kompilasi',
        icon: 'i-lucide-presentation',
        to: '/executive-summary',
      },
      {
        label: 'Import Executive Summary Report',
        icon: 'i-lucide-upload',
        to: '/executive-summary/upload',
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
    children: [
      {
        label: 'Consulting Service Dashboard',
        icon: 'i-lucide-layout-dashboard',
        to: '/consulting-service',
      },
      {
        label: 'Import Consulting Document',
        icon: 'i-lucide-upload',
        to: '/consulting-service/upload',
      }
    ]
  },
  {
    label: '12. Internal Audit Performance',
    icon: 'i-lucide-trending-up',
    to: '/kpi-performance',
    children: [
      {
        label: 'Internal Audit Performance Dashboard',
        icon: 'i-lucide-layout-dashboard',
        to: '/kpi-performance',
      },
      {
        label: 'Import Laporan Kinerja',
        icon: 'i-lucide-upload',
        to: '/kpi-performance/upload',
      }
    ]
  },
  {
    label: '13. Quality Assurance Review',
    icon: 'i-lucide-shield-check',
    to: '/quality-assurance',
    children: [
      {
        label: 'Quality Assurance Dashboard',
        icon: 'i-lucide-layout-dashboard',
        to: '/quality-assurance',
      },
      {
        label: 'Import Periodic Self Assessment',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-periodic-self-assessment',
      },
      {
        label: 'Import SAIV',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-saiv',
      },
      {
        label: 'Import QAR Report',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import',
      },
      {
        label: 'Import IACM',
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-iacm',
      }
    ]
  },
  {
    label: 'Analytics',
    icon: 'i-lucide-pie-chart',
    to: '/analytics',
  },
  {
    label: 'Master Data',
    icon: 'i-lucide-database',
    to: '/master/employee',
    children: [
      {
        label: 'Employee Management',
        icon: 'i-lucide-users-round',
        to: '/master/employee',
      },
      {
        label: 'Department',
        icon: 'i-lucide-building-2',
        to: '/master/department',
      },
    ]
  },
  
]]

const searchQuery = ref('')
const { canImportPlanDocs, isAdmin, canManageAudits } = useRbac()

// 2. Gunakan Computed agar menu bereaksi setiap kali pengguna pindah halaman dan mempertimbangkan Hak Akses (RBAC)
const items = computed<NavigationMenuItem[][]>(() => {
  return rawItems.map(group => {
    return group
      .filter(parent => {
        // Master Data is Settings & User Management (NONE for Auditor/Auditee/Viewer)
        if (parent.to === '/master/employee' && !isAdmin.value && !canManageAudits.value) {
          return false
        }
        return true
      })
      .map(parent => {
        const q = searchQuery.value.toLowerCase()
        
        // Filter children according to RBAC (e.g. Import Plan submodules are NONE for Auditor)
        let filteredChildren = parent.children ? [...parent.children] : undefined
        
        if (filteredChildren) {
          filteredChildren = filteredChildren.filter(child => {
            const childTo = (child.to as string) || ''
            // Working paper upload is UPLOAD for Auditor, so allow it.
            // Other import pages (Annual Audit, Activity Plan, Assignment Letter, QA, Executive Summary, KPI) are NONE for Auditor.
            if (childTo.includes('/upload') || childTo.includes('/import')) {
              if (childTo.includes('/working-paper/upload')) return true
              if (!canImportPlanDocs.value) return false
            }
            return true
          })
        }

        let childrenMatches = false
        if (q && filteredChildren) {
          filteredChildren = filteredChildren.filter(child => (child.label?.toLowerCase() || '').includes(q) || (parent.label?.toLowerCase() || '').includes(q))
          childrenMatches = filteredChildren.length > 0
        }

        // Cek apakah ada submenu (child) yang URL-nya cocok dengan URL saat ini
        const hasActiveChild = filteredChildren?.some(child => 
          route.path.startsWith(child.to as string)
        )

        return {
          ...parent,
          children: filteredChildren,
          active: hasActiveChild,
          defaultOpen: hasActiveChild || parent.defaultOpen || (q.length > 0 && childrenMatches)
        }
      }).filter(parent => {
         if (parent.children && parent.children.length === 0 && rawItems.some(g => g.some(p => p.to === parent.to && p.children && p.children.length > 0))) {
           // If parent lost all its children due to RBAC and was only a parent container
         }
         if (!searchQuery.value) return true
         const q = searchQuery.value.toLowerCase()
         const parentMatches = (parent.label?.toLowerCase() || '').includes(q)
         const hasVisibleChildren = parent.children && parent.children.length > 0
         return parentMatches || hasVisibleChildren
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
    id="main-sidebar-final"
    collapsible 
    resizable
    :min-size="21"
    :default-size="21"
    :max-size="35"
    :collapsed-size="0"
  >
    <template #header="{ collapsed }">
      <Logo v-if="!collapsed" class="h-7 w-auto shrink-0" hide-subtitle text-class="text-xl" />
      <Logo v-else icon-only class="h-6 w-auto mx-auto" />
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
            <UButton v-if="!authStore.isLoggedIn" to="/auth/login" color="primary" variant="solid">Login</UButton>
            <UDropdownMenu v-else :items="userDropdownItems">
              <UAvatar :alt="authStore.getUser?.fullName || 'User'" size="md" />
              
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

  <!-- Mobile Navigation Slideover -->
  <USlideover v-model:open="isMobileMenuOpen" side="left">
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
          />
        </div>
      </div>
    </template>
  </USlideover>
  </UDashboardGroup>
</template>