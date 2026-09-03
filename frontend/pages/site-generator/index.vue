<template>
  <div class="space-y-6 max-w-7xl mx-auto pb-12">
    <!-- Page Header & Banner -->
    <div class="rounded-2xl p-6 sm:p-8 bg-gradient-to-r from-primary-900 via-primary-800 to-gray-900 text-white shadow-xl relative overflow-hidden">
      <!-- Background decorative pattern -->
      <div class="absolute -right-10 -bottom-10 opacity-10 pointer-events-none">
        <UIcon
          name="i-lucide-globe"
          class="w-80 h-80"
        />
      </div>

      <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div class="space-y-2 max-w-2xl">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/10 backdrop-blur-md border border-white/20 text-xs font-semibold text-primary-200">
            <UIcon
              name="i-lucide-layers"
              class="w-3.5 h-3.5"
            />
            AuditSphere Multi-Tenant Cloud Architecture
          </div>
          <h1 class="text-2xl sm:text-3xl font-bold tracking-tight text-white">
            Client Site Generator & Provisioning
          </h1>
          <p class="text-sm text-primary-100/90 leading-relaxed">
            Instantly deploy isolated, enterprise-grade audit portals with dedicated PostgreSQL databases, Kong API Gateways, and Google Drive storage for corporate clients like <span class="font-mono font-semibold text-white underline decoration-primary-400">accenture.auditsphere.id</span> and <span class="font-mono font-semibold text-white underline decoration-primary-400">telkom.auditsphere.id</span>.
          </p>
        </div>

        <div class="flex items-center gap-3 shrink-0">
          <ReusableButton
            variant="fill"
            color="primary"
            size="lg"
            icon="i-lucide-plus"
            class="bg-white text-primary-900 hover:bg-primary-50 shadow-lg font-bold border-none"
            @click="isModalOpen = true"
          >
            Generate New Site
          </ReusableButton>
        </div>
      </div>
    </div>

    <!-- Quick Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-50 dark:bg-emerald-950/60 border border-emerald-200 dark:border-emerald-800 flex items-center justify-center text-emerald-600 dark:text-emerald-400 shrink-0">
          <UIcon
            name="i-lucide-check-circle"
            class="w-6 h-6"
          />
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-medium">
            Active Client Portals
          </div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white mt-0.5">
            {{ activeSitesCount }}
          </div>
          <div class="text-[11px] text-emerald-600 dark:text-emerald-400 font-medium mt-0.5 flex items-center gap-1">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-ping" />
            All systems operational
          </div>
        </div>
      </div>

      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-blue-50 dark:bg-blue-950/60 border border-blue-200 dark:border-blue-800 flex items-center justify-center text-blue-600 dark:text-blue-400 shrink-0">
          <UIcon
            name="i-lucide-globe"
            class="w-6 h-6"
          />
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-medium">
            Domain Root
          </div>
          <div class="text-lg font-mono font-bold text-gray-900 dark:text-white mt-0.5">
            *.auditsphere.id
          </div>
          <div class="text-[11px] text-gray-400 font-mono mt-0.5">
            Wildcard SSL Active
          </div>
        </div>
      </div>

      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-purple-50 dark:bg-purple-950/60 border border-purple-200 dark:border-purple-800 flex items-center justify-center text-purple-600 dark:text-purple-400 shrink-0">
          <UIcon
            name="i-lucide-database"
            class="w-6 h-6"
          />
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-medium">
            Isolated Databases
          </div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white mt-0.5">
            {{ activeSitesCount * 5 }}
          </div>
          <div class="text-[11px] text-purple-600 dark:text-purple-400 font-medium mt-0.5">
            5 DBs per client tenant
          </div>
        </div>
      </div>

      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-50 dark:bg-amber-950/60 border border-amber-200 dark:border-amber-800 flex items-center justify-center text-amber-600 dark:text-amber-400 shrink-0">
          <UIcon
            name="i-lucide-hard-drive"
            class="w-6 h-6"
          />
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-medium">
            Cloud Evidence Vaults
          </div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white mt-0.5">
            {{ activeSitesCount }}
          </div>
          <div class="text-[11px] text-amber-600 dark:text-amber-400 font-medium mt-0.5">
            Google Drive Silo Isolation
          </div>
        </div>
      </div>
    </div>

    <!-- Toolbar: Search, Filters & View Toggle -->
    <div class="p-4 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col sm:flex-row items-center justify-between gap-4">
      <div class="flex items-center gap-3 w-full sm:w-auto">
        <div class="relative flex-1 sm:w-80">
          <UIcon
            name="i-lucide-search"
            class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
          />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by client, domain, or slug..."
            class="w-full pl-9 pr-4 py-2 text-xs rounded-xl border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all"
          >
        </div>

        <select
          v-model="statusFilter"
          class="text-xs rounded-xl border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
        >
          <option value="ALL">
            All Status
          </option>
          <option value="Active">
            Active
          </option>
          <option value="Provisioning">
            Provisioning
          </option>
          <option value="Maintenance">
            Maintenance
          </option>
        </select>
      </div>

      <div class="flex items-center gap-2 self-end sm:self-auto">
        <span class="text-xs text-gray-500 font-medium">
          Showing <span class="text-gray-900 dark:text-white font-bold">{{ filteredSites.length }}</span> client sites
        </span>
      </div>
    </div>

    <!-- Client Sites Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
      <div
        v-for="site in filteredSites"
        :key="site.id"
        class="rounded-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 shadow-sm hover:shadow-md transition-all duration-200 overflow-hidden flex flex-col justify-between"
      >
        <!-- Card Header -->
        <div class="p-6 border-b border-gray-100 dark:border-gray-800">
          <div class="flex items-start justify-between gap-4">
            <div class="flex items-center gap-3">
              <div
                class="w-12 h-12 rounded-xl flex items-center justify-center text-white font-bold text-lg shadow-sm shrink-0"
                :style="{ backgroundColor: site.brandColor || '#0284c7' }"
              >
                {{ site.clientName.slice(0, 2).toUpperCase() }}
              </div>
              <div>
                <h3 class="font-bold text-base text-gray-900 dark:text-white flex items-center gap-2">
                  {{ site.clientName }}
                </h3>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ site.industry }}
                </p>
              </div>
            </div>

            <!-- Status Badge -->
            <span
              :class="[
                'px-2.5 py-1 rounded-full text-xs font-semibold flex items-center gap-1.5 shrink-0 select-none',
                site.status === 'Active'
                  ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800/40'
                  : 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 border border-amber-200 dark:border-amber-800/40'
              ]"
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :class="site.status === 'Active' ? 'bg-emerald-500' : 'bg-amber-500'"
              />
              {{ site.status }}
            </span>
          </div>

          <!-- Domains & URLs Display -->
          <div class="mt-4 space-y-2 bg-gray-50 dark:bg-gray-850/60 p-3.5 rounded-xl border border-gray-100 dark:border-gray-800">
            <div class="flex items-center justify-between gap-2 text-xs">
              <span class="text-gray-400 flex items-center gap-1.5 shrink-0">
                <UIcon
                  name="i-lucide-globe"
                  class="w-3.5 h-3.5 text-primary"
                />
                Frontend:
              </span>
              <a
                :href="`https://${site.domain}`"
                target="_blank"
                class="font-mono font-semibold text-primary hover:underline truncate"
                :title="`Open https://${site.domain}`"
              >
                https://{{ site.domain }}
              </a>
            </div>

            <div class="flex items-center justify-between gap-2 text-xs">
              <span class="text-gray-400 flex items-center gap-1.5 shrink-0">
                <UIcon
                  name="i-lucide-network"
                  class="w-3.5 h-3.5 text-secondary"
                />
                API Gateway:
              </span>
              <span class="font-mono text-gray-600 dark:text-gray-300 truncate">
                https://{{ site.apiDomain }}
              </span>
            </div>
          </div>
        </div>

        <!-- Card Meta Info -->
        <div class="px-6 py-4 grid grid-cols-2 gap-3 text-xs border-b border-gray-100 dark:border-gray-800 bg-gray-50/40 dark:bg-gray-850/20">
          <div>
            <span class="text-gray-400 block">Lead / Admin</span>
            <span class="font-medium text-gray-800 dark:text-gray-200">{{ site.adminName }}</span>
          </div>
          <div>
            <span class="text-gray-400 block">Container Ports</span>
            <span class="font-mono text-gray-700 dark:text-gray-300">FE: {{ site.frontendPort }} | Kong: {{ site.kongPort }}</span>
          </div>
          <div>
            <span class="text-gray-400 block">Database Cluster</span>
            <span class="font-mono text-gray-700 dark:text-gray-300">rb_audit_*_{{ site.slug }}</span>
          </div>
          <div>
            <span class="text-gray-400 block">Compliance Standard</span>
            <span class="font-medium text-gray-800 dark:text-gray-200 truncate block">{{ site.complianceFramework }}</span>
          </div>
        </div>

        <!-- Card Actions Footer -->
        <div class="px-6 py-3.5 bg-gray-50 dark:bg-gray-850/80 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="text-xs text-gray-500 hover:text-gray-900 dark:hover:text-white flex items-center gap-1 font-medium transition-colors"
              @click="openInspectModal(site)"
            >
              <UIcon
                name="i-lucide-code-2"
                class="w-3.5 h-3.5"
              />
              CLI Script
            </button>
            <span class="text-gray-300 dark:text-gray-700">•</span>
            <button
              type="button"
              class="text-xs text-gray-500 hover:text-gray-900 dark:hover:text-white flex items-center gap-1 font-medium transition-colors"
              @click="openNginxModal(site)"
            >
              <UIcon
                name="i-lucide-server"
                class="w-3.5 h-3.5"
              />
              Nginx Conf
            </button>
          </div>

          <a
            :href="`https://${site.domain}`"
            target="_blank"
            class="inline-flex items-center gap-1.5 text-xs font-bold text-primary hover:text-primary-700 hover:underline transition-all"
          >
            Visit Portal
            <UIcon
              name="i-lucide-external-link"
              class="w-3.5 h-3.5"
            />
          </a>
        </div>
      </div>
    </div>

    <!-- Empty State when filter yields 0 -->
    <div
      v-if="filteredSites.length === 0"
      class="p-12 text-center rounded-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800"
    >
      <UIcon
        name="i-lucide-globe-2"
        class="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-3"
      />
      <h3 class="font-bold text-gray-800 dark:text-gray-200">
        No client sites matched your search
      </h3>
      <p class="text-xs text-gray-500 mt-1">
        Try modifying your keyword or status filter, or generate a new client site.
      </p>
    </div>

    <!-- Modal: Generate New Client Site -->
    <TenantProvisionModal
      v-model="isModalOpen"
      :next-suggested-port="nextPort"
      :next-suggested-kong-port="nextKongPort"
      @site-created="onSiteCreated"
    />

    <!-- Inspect Modal: Automated Onboarding Script -->
    <UModal
      v-model="isInspectModalOpen"
      :ui="{ width: 'sm:max-w-2xl' }"
    >
      <div
        v-if="selectedSite"
        class="p-6 space-y-4"
      >
        <div class="flex items-center justify-between border-b pb-3 border-gray-200 dark:border-gray-800">
          <div class="flex items-center gap-2">
            <UIcon
              name="i-lucide-terminal"
              class="w-5 h-5 text-primary"
            />
            <h3 class="font-bold text-base text-gray-900 dark:text-white">
              Provisioning Command: {{ selectedSite.clientName }}
            </h3>
          </div>
          <button
            type="button"
            class="text-gray-400 hover:text-gray-600 p-1"
            @click="isInspectModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <p class="text-xs text-gray-500 dark:text-gray-400">
          Execute this automated script on the host server to create the isolated databases, seed standard RBAC roles, configure Nginx, and activate SSL:
        </p>

        <div class="bg-gray-950 text-emerald-400 p-4 rounded-xl font-mono text-xs overflow-x-auto shadow-inner border border-gray-800">
          <code>./scripts/onboard-tenant.sh {{ selectedSite.slug }} "{{ selectedSite.clientName }}" "1bX7yZ9kL0mN8pQ2rS4tU6vW8xYz1234A" {{ selectedSite.frontendPort }} {{ selectedSite.kongPort }}</code>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <ReusableButton
            variant="fill"
            color="primary"
            size="sm"
            @click="isInspectModalOpen = false"
          >
            Done
          </ReusableButton>
        </div>
      </div>
    </UModal>

    <!-- Inspect Modal: Nginx Virtual Host Conf -->
    <UModal
      v-model="isNginxModalOpen"
      :ui="{ width: 'sm:max-w-2xl' }"
    >
      <div
        v-if="selectedSite"
        class="p-6 space-y-4"
      >
        <div class="flex items-center justify-between border-b pb-3 border-gray-200 dark:border-gray-800">
          <div class="flex items-center gap-2">
            <UIcon
              name="i-lucide-server"
              class="w-5 h-5 text-secondary"
            />
            <h3 class="font-bold text-base text-gray-900 dark:text-white">
              Nginx Virtual Host Config: /etc/nginx/sites-available/{{ selectedSite.domain }}.conf
            </h3>
          </div>
          <button
            type="button"
            class="text-gray-400 hover:text-gray-600 p-1"
            @click="isNginxModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div class="bg-gray-950 text-gray-200 p-4 rounded-xl font-mono text-[11px] overflow-x-auto shadow-inner border border-gray-800 max-h-96">
          <pre><code>server {
    listen 80;
    server_name {{ selectedSite.domain }};
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name {{ selectedSite.domain }};

    ssl_certificate /etc/letsencrypt/live/{{ selectedSite.domain }}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/{{ selectedSite.domain }}/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:{{ selectedSite.frontendPort }};
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 443 ssl http2;
    server_name {{ selectedSite.apiDomain }};

    location / {
        proxy_pass http://127.0.0.1:{{ selectedSite.kongPort }};
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}</code></pre>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <ReusableButton
            variant="fill"
            color="primary"
            size="sm"
            @click="isNginxModalOpen = false"
          >
            Close
          </ReusableButton>
        </div>
      </div>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import TenantProvisionModal from '~/components/site-generator/TenantProvisionModal.vue'
import ReusableButton from '~/components/shared/ReusableButton.vue'

interface ClientSite {
  id: string
  clientName: string
  slug: string
  domain: string
  apiDomain: string
  industry: string
  status: 'Active' | 'Provisioning' | 'Maintenance'
  brandColor: string
  adminName: string
  adminEmail: string
  frontendPort: number
  kongPort: number
  complianceFramework: string
  createdAt: string
}

const isModalOpen = ref(false)
const isInspectModalOpen = ref(false)
const isNginxModalOpen = ref(false)
const selectedSite = ref<ClientSite | null>(null)

const searchQuery = ref('')
const statusFilter = ref('ALL')

// Initial Showcase Client Tenants: Accenture & Telkom as requested
const clientSites = ref<ClientSite[]>([
  {
    id: 'tenant-1',
    clientName: 'Accenture Indonesia',
    slug: 'accenture',
    domain: 'accenture.auditsphere.id',
    apiDomain: 'api-accenture.auditsphere.id',
    industry: 'Consulting & Professional Services',
    status: 'Active',
    brandColor: '#A100FF',
    adminName: 'Budi Santoso, CIA',
    adminEmail: 'cae@accenture.com',
    frontendPort: 3010,
    kongPort: 8090,
    complianceFramework: 'ISO 27001 + IIA Global Standards',
    createdAt: '2026-08-15'
  },
  {
    id: 'tenant-2',
    clientName: 'Telkom Indonesia',
    slug: 'telkom',
    domain: 'telkom.auditsphere.id',
    apiDomain: 'api-telkom.auditsphere.id',
    industry: 'Telecommunications',
    status: 'Active',
    brandColor: '#ED1D24',
    adminName: 'Siti Rahma, CISA',
    adminEmail: 'audit@telkom.co.id',
    frontendPort: 3011,
    kongPort: 8091,
    complianceFramework: 'COSO ERM + State-Owned (BUMN) Standard',
    createdAt: '2026-08-20'
  },
  {
    id: 'tenant-3',
    clientName: 'Bank Mandiri (Persero) Tbk',
    slug: 'mandiri',
    domain: 'mandiri.auditsphere.id',
    apiDomain: 'api-mandiri.auditsphere.id',
    industry: 'Banking & Financial Services',
    status: 'Active',
    brandColor: '#003D79',
    adminName: 'Hendro Prasetyo, QIA',
    adminEmail: 'internal-audit@bankmandiri.co.id',
    frontendPort: 3012,
    kongPort: 8092,
    complianceFramework: 'Bank Indonesia / OJK Regulatory Standard',
    createdAt: '2026-08-25'
  },
  {
    id: 'tenant-4',
    clientName: 'Pertamina (Persero)',
    slug: 'pertamina',
    domain: 'pertamina.auditsphere.id',
    apiDomain: 'api-pertamina.auditsphere.id',
    industry: 'Energy, Oil & Gas',
    status: 'Active',
    brandColor: '#00853F',
    adminName: 'Agus Setiawan, CRMA',
    adminEmail: 'audit.investigasi@pertamina.com',
    frontendPort: 3013,
    kongPort: 8093,
    complianceFramework: 'ISO 27001 + UU PDP',
    createdAt: '2026-08-28'
  }
])

const activeSitesCount = computed(() => {
  return clientSites.value.filter(s => s.status === 'Active').length
})

const nextPort = computed(() => {
  const maxFePort = Math.max(...clientSites.value.map(s => s.frontendPort), 3009)
  return maxFePort + 1
})

const nextKongPort = computed(() => {
  const maxKong = Math.max(...clientSites.value.map(s => s.kongPort), 8089)
  return maxKong + 1
})

const filteredSites = computed(() => {
  return clientSites.value.filter((site) => {
    const matchesSearch
      = !searchQuery.value
        || site.clientName.toLowerCase().includes(searchQuery.value.toLowerCase())
        || site.slug.toLowerCase().includes(searchQuery.value.toLowerCase())
        || site.domain.toLowerCase().includes(searchQuery.value.toLowerCase())
        || site.industry.toLowerCase().includes(searchQuery.value.toLowerCase())

    const matchesStatus = statusFilter.value === 'ALL' || site.status === statusFilter.value

    return matchesSearch && matchesStatus
  })
})

const openInspectModal = (site: ClientSite) => {
  selectedSite.value = site
  isInspectModalOpen.value = true
}

const openNginxModal = (site: ClientSite) => {
  selectedSite.value = site
  isNginxModalOpen.value = true
}

const onSiteCreated = (newSite: ClientSite) => {
  clientSites.value.unshift(newSite)
}
</script>
