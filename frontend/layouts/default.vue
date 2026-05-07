<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { NavigationMenuItem } from '@nuxt/ui'

// Gunakan useRoute untuk mendapatkan URL saat ini
const route = useRoute()

// 1. Simpan data menu dalam variabel mentah (raw data)
const rawItems: NavigationMenuItem[][] = [[
  {
    label: '1. Audit Charter',
    icon: 'charter',
    to: '/audit-charter'
  }, 
  {
    label: '2. Risk Profile',
    icon: 'i-lucide-inbox',
    to: '/risk-profile'
  }, 
  {
    label: '3. Strategic Plan Internal Audit',
    icon: 'i-lucide-users',
    to: '/strategic-audit-plan',
  },
  {
    label: '4. Annual Audit Plan',
    icon: 'i-lucide-users',
    to: '/annual-audit',
  },
  {
    label: '5. Audit Activity Plan',
    icon: 'i-lucide-users',
    to: '/audit-activity-plan'
  }, 
  {
    label: '6. Assignment Letter',
    icon: 'i-lucide-users',
    to: '/assignment-letter',
  },
  {
    label: '7. Working Papers',
    icon: 'i-lucide-users',
    to: '/working-paper',
    children :[
      {
        label: 'Audit Field Work',
        icon: 'i-lucide-users',
        to: '/audit-fieldwork'
      }
    ]
  },
  {
    label: '8. Audit Result Report',
    icon: 'i-lucide-users',
    to: '/audit-result-report',
  }, 
  {
    label: '9. Dashboard',
    icon: 'i-lucide-users',
    to: '/dashboard' 
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
      <Logo v-if="!collapsed" class="h-5 w-auto shrink-0" />
      <UIcon v-else name="i-simple-icons-nuxtdotjs" class="size-5 text-primary mx-auto" />
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
          <h1 class="text-xl font-semibold text-[var(--text-main)]">RBIA System</h1>
          <UColorModeButton />
        </div>
    </template>
    <template #body>
      <div class="min-h-screen bg-[var(--bg-main)] min-w-max transition-colors duration-300">
        <UMain class="max-w-7xl mx-auto">
          <slot />
        </UMain>
      </div>
    </template>
  </UDashboardPanel>

  
  </UDashboardGroup>
</template>