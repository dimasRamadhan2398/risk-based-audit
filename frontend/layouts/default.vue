<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import type { NavigationMenuItem } from '@nuxt/ui'
import { useAuthStore } from '~/stores/auth'
import { useI18n } from '~/composables/useI18n'
import { triggerScrollReset } from '~/utils/scroll'

const route = useRoute()
const authStore = useAuthStore()
const { t, locale, setLocale } = useI18n()

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
const rawItems = computed<NavigationMenuItem[][]>(() => [[
  // 1. Dashboard
  {
    label: t('navigation.dashboard'),
    icon: 'i-lucide-layout-dashboard',
    to: '/dashboard'
  },

  // 2. Audit Charter
  {
    label: t('navigation.auditCharter'),
    icon: 'i-lucide-scroll-text',
    to: '/audit-charter'
  },

  // 3. Risk Profile & Audit Universe
  {
    label: t('navigation.riskProfileUniverse'),
    icon: 'i-lucide-shield-alert',
    to: '/risk-profile',
    children: [
      {
        label: t('navigation.corporateRiskProfile'),
        icon: 'i-lucide-users',
        to: '/risk-profile'
      },
      {
        label: t('navigation.riskAppetiteStatement'),
        icon: 'i-lucide-clipboard-check',
        to: '/risk-appetite'
      },
      {
        label: t('navigation.riskFactors'),
        icon: 'i-lucide-activity',
        to: '/risk-profile/risk-factors'
      },
      {
        label: t('navigation.auditUniverse'),
        icon: 'i-lucide-globe',
        to: '/risk-profile/audit-universe'
      },
      {
        label: t('navigation.auditPriority'),
        icon: 'i-lucide-list-ordered',
        to: '/risk-profile/audit-universe?tab=priority'
      },
      {
        label: t('navigation.riskControlMatrix'),
        icon: 'i-lucide-grid',
        to: '/risk-profile/risk-control-matrix'
      }
    ]
  },

  // 4. Audit Planning
  {
    label: t('navigation.auditPlanning'),
    icon: 'i-lucide-calendar-days',
    type: 'trigger',
    children: [
      {
        label: t('navigation.strategicAuditPlan'),
        icon: 'i-lucide-target',
        type: 'trigger',
        children: [
          {
            label: t('navigation.strategicAuditPlan'),
            icon: 'i-lucide-layout-dashboard',
            to: '/strategic-audit-plan'
          },
          {
            label: t('navigation.importStrategicAuditPlan'),
            icon: 'i-lucide-upload',
            to: '/strategic-audit-plan/upload'
          }
        ]
      },
      {
        label: t('navigation.internalAuditPerformance'),
        icon: 'i-lucide-trending-up',
        type: 'trigger',
        children: [
          {
            label: t('navigation.internalAuditPerformance'),
            icon: 'i-lucide-layout-dashboard',
            to: '/kpi-performance'
          },
          {
            label: t('navigation.importPerformanceReport'),
            icon: 'i-lucide-upload',
            to: '/kpi-performance/upload'
          }
        ]
      },
      {
        label: t('navigation.annualAuditPlan'),
        icon: 'i-lucide-calendar',
        type: 'trigger',
        children: [
          {
            label: t('navigation.createAnnualAuditPlan'),
            icon: 'i-lucide-layout-dashboard',
            to: '/annual-audit'
          },
          {
            label: t('navigation.importAnnualAuditPlan'),
            icon: 'i-lucide-upload',
            to: '/annual-audit/upload'
          },
          {
            label: t('navigation.auditExecutionStatus'),
            icon: 'i-lucide-check-circle-2',
            to: '/audit-execution-status'
          }
        ]
      },
      {
        label: t('navigation.auditActivityPlan'),
        icon: 'i-lucide-clipboard-list',
        type: 'trigger',
        children: [
          {
            label: t('navigation.createAuditActivityPlan'),
            icon: 'i-lucide-layout-dashboard',
            to: '/audit-activity-plan'
          },
          {
            label: t('navigation.importAuditActivityPlan'),
            icon: 'i-lucide-upload',
            to: '/audit-activity-plan/upload'
          }
        ]
      },
      {
        label: t('navigation.assignmentLetter'),
        icon: 'i-lucide-file-signature',
        type: 'trigger',
        children: [
          {
            label: t('navigation.createAssignmentLetter'),
            icon: 'i-lucide-layout-dashboard',
            to: '/assignment-letter'
          },
          {
            label: t('navigation.importAssignmentLetter'),
            icon: 'i-lucide-upload',
            to: '/assignment-letter/upload'
          }
        ]
      },
      {
        label: t('navigation.workingPaper'),
        icon: 'i-lucide-file-text',
        type: 'trigger',
        children: [
          {
            label: t('navigation.createWorkingPaper'),
            icon: 'i-lucide-file-plus',
            to: '/working-paper'
          },
          {
            label: t('navigation.importWorkingPaper'),
            icon: 'i-lucide-upload',
            to: '/working-paper/upload'
          }
        ]
      },
      {
        label: t('navigation.auditFieldwork'),
        icon: 'i-lucide-briefcase',
        to: '/audit-fieldwork'
      }
    ]
  },

  // 5. Audit Result Report
  {
    label: t('navigation.auditResultReport'),
    icon: 'i-lucide-file-check-2',
    type: 'trigger',
    children: [
      {
        label: t('navigation.resultReportsLha'),
        icon: 'i-lucide-list',
        to: '/audit-result-report'
      },
      {
        label: t('navigation.importLhaDocument'),
        icon: 'i-lucide-upload',
        to: '/audit-result-report/upload'
      },
      {
        label: t('navigation.executiveSummary'),
        icon: 'i-lucide-presentation',
        type: 'trigger',
        children: [
          {
            label: t('navigation.executiveSummary'),
            icon: 'i-lucide-layout-dashboard',
            to: '/audit-result-report/executive-summary'
          },
          {
            label: t('navigation.importExecutiveSummary'),
            icon: 'i-lucide-upload',
            to: '/audit-result-report/executive-summary-upload'
          }
        ]
      },
      {
        label: t('navigation.executiveSummaryCompilation'),
        icon: 'i-lucide-file-text',
        type: 'trigger',
        children: [
          {
            label: t('navigation.createExecutiveSummaryCompilation'),
            icon: 'i-lucide-presentation',
            to: '/executive-summary'
          },
          {
            label: t('navigation.importExecutiveSummaryCompilation'),
            icon: 'i-lucide-upload',
            to: '/executive-summary/upload'
          }
        ]
      },
      {
        label: t('navigation.actionTakenReport'),
        icon: 'i-lucide-clipboard-check',
        to: '/action-taken-report'
      },
      {
        label: t('navigation.clientSatisfactionSurvey'),
        icon: 'i-lucide-smile',
        to: '/audit-result-report/satisfaction-survey'
      }
    ]
  },

  // 6. Consulting Service
  {
    label: t('navigation.consultingService'),
    icon: 'i-lucide-messages-square',
    to: '/consulting-service',
    children: [
      {
        label: t('navigation.consultingServiceDashboard'),
        icon: 'i-lucide-layout-dashboard',
        to: '/consulting-service'
      },
      {
        label: t('navigation.importConsultingDocument'),
        icon: 'i-lucide-upload',
        to: '/consulting-service/upload'
      }
    ]
  },

  // 7. Quality Assurance Review
  {
    label: t('navigation.qualityAssuranceReview'),
    icon: 'i-lucide-shield-check',
    to: '/quality-assurance',
    children: [
      {
        label: t('navigation.qualityAssuranceDashboard'),
        icon: 'i-lucide-layout-dashboard',
        to: '/quality-assurance'
      },
      {
        label: t('navigation.importPeriodicSelfAssessment'),
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-periodic-self-assessment'
      },
      {
        label: t('navigation.importSaiv'),
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-saiv'
      },
      {
        label: t('navigation.importQarReport'),
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import'
      },
      {
        label: t('navigation.importIacm'),
        icon: 'i-lucide-upload',
        to: '/quality-assurance/import-iacm'
      }
    ]
  },

  // 8. Analytics
  {
    label: t('navigation.analytics'),
    icon: 'i-lucide-pie-chart',
    to: '/analytics',
    children: [
      {
        label: t('navigation.riskScoringPrediction'),
        icon: 'i-lucide-binary',
        to: '/analytics?tab=xgboost'
      },
      {
        label: t('navigation.anomalyDetection'),
        icon: 'i-lucide-shield-alert',
        to: '/analytics?tab=isolation'
      },
      {
        label: t('navigation.detectedAnomaliesDetail'),
        icon: 'i-lucide-table-properties',
        to: '/analytics?tab=isolation'
      },
      {
        label: t('navigation.nlpAnalysis'),
        icon: 'i-lucide-file-search',
        to: '/analytics?tab=nlp'
      },
      {
        label: t('navigation.kpiForecast'),
        icon: 'i-lucide-trending-up',
        to: '/analytics/kpi-forecast'
      }
    ]
  },

  // Master Data
  {
    label: t('navigation.masterData'),
    icon: 'i-lucide-database',
    to: '/master/employee',
    children: [
      {
        label: t('navigation.employeeManagement'),
        icon: 'i-lucide-users-round',
        to: '/master/employee'
      },
      {
        label: t('navigation.department'),
        icon: 'i-lucide-building-2',
        to: '/master/department'
      }
    ]
  },

  // Settings
  {
    label: t('navigation.settings'),
    icon: 'i-lucide-settings',
    to: '/settings',
    children: [
      {
        label: t('navigation.generalSettings'),
        icon: 'i-lucide-sliders',
        to: '/settings'
      },
      {
        label: t('navigation.security2fa'),
        icon: 'i-lucide-shield-check',
        to: '/settings/mfa'
      },
      {
        label: t('navigation.trustedDevices'),
        icon: 'i-lucide-laptop',
        to: '/settings/devices'
      }
    ]
  }
]])

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
  return rawItems.value.map((group) => {
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
      label: t('navigation.userMenu.settings'),
      icon: 'i-lucide-settings',
      to: '/settings'
    }
  ],
  [
    {
      label: t('navigation.userMenu.logout'),
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
            :placeholder="t('navigation.searchPlaceholder')"
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
          <div class="flex items-center gap-3">
            <!-- Language Switcher Toggle -->
            <div class="flex items-center gap-1 bg-[var(--bg-surface)] p-1 rounded-xl border border-[var(--border-main)] shadow-sm">
              <button
                type="button"
                class="px-2 py-0.5 text-xs font-bold rounded-lg transition-all duration-200 flex items-center gap-1"
                :class="locale === 'id' ? 'bg-primary-500 text-white shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
                @click="setLocale('id')"
              >
                <span>🇮🇩</span> ID
              </button>
              <button
                type="button"
                class="px-2 py-0.5 text-xs font-bold rounded-lg transition-all duration-200 flex items-center gap-1"
                :class="locale === 'en' ? 'bg-primary-500 text-white shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
                @click="setLocale('en')"
              >
                <span>🇬🇧</span> EN
              </button>
            </div>
            <UColorModeButton />
            <UButton
              v-if="!authStore.isLoggedIn"
              to="/auth/login"
              color="primary"
              variant="solid"
            >
              {{ t('auth.login.button') }}
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
              :title="t('navigation.securityWarningTitle')"
              :description="t('navigation.securityWarningDesc')"
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
