<template>
    <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
      <UTable
        :data="store.filteredPlans"    
        :columns="store.columns"
        :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data rencana audit.' }"
        class="w-full text-sm text-left"
      >
        <template #activity-cell="{ row }">
          <div class="py-2">
            <div class="font-bold text-gray-600">{{ row.original.code }}</div>
            <!-- <div class="font-medium text-gray-900 dark:text-white">{{row.original.name }}</div>
            <UBadge color="primary" variant="soft" size="md" class="mt-1">
              {{ row.original.category }}
            </UBadge> -->
          </div>
        </template>

        <template #department-cell="{ row }">
          <div class="flex gap-1 flex-wrap">
            <UBadge 
              v-for="dept in Array.from(new Set(row.original.activities.map(a => a.department)))" 
              :key="dept"
              color="neutral"
              variant="soft"
              size="md"
            >
              {{ dept }}
            </UBadge>
          </div>
        </template>

        <template #timeline-cell="{ row }">
          <div class="font-bold text-primary-600 mr-1">{{ row.original.year }}</div>
          <UBadge v-for="q in row.original.quarters" :key="q" color="primary" variant="subtle" size="md">
            {{ q }}
          </UBadge>
          <div class="flex gap-1 flex-wrap mt-1">
            <UBadge 
              v-for="idx in row.original.selectedMonths.slice().sort((a, b) => a - b)" 
              :key="idx"
              color="primary" 
              variant="outline" 
              size="md"
            >
              {{ store.monthsList[idx] }}
            </UBadge>
          </div>
        </template>

        <template #progress-cell="{ row }">
          <UProgress v-model="store.progressAudit" color="secondary" status />
        </template>

        <template #status-data="{ row }">
          <span 
            class="w-2.5 h-2.5 rounded-full inline-block"
            :class="store.getStatusColor(row.original.status)"
          />
        </template>

        <template #actions-cell="{ row }">
          <div class="flex">
            <UButton
              label="View"
              color="primary"
              variant="ghost"
              size="lg"
              @click="store.openViewModal(row.original)"
            />
            <h4> | </h4>
            <UButton
              label="Edit"
              color="primary"
              variant="ghost"
              size="lg"
              @click="store.handleEdit(row.original)"
            />
          </div>
        </template>

        <!-- <template #notes-cell="{ row }">
          <div class="max-w-[250px] text-sm text-gray-600 dark:text-gray-400">
            <template v-if="!isNotesLongText(row.original.notes!)">
              <span class="italic">{{ row.original.notes || '-' }}</span>
            </template>

            <template v-else>
              <span class="italic whitespace-normal break-words">
                {{ expandedNotesRows.has(row.original.id!) 
                  ? row.original.notes 
                  : row.original.notes?.slice(0, 50) + '...' 
                }}
              </span>
              
              <UButton
                :label="expandedNotesRows.has(row.original.id!) ? 'Show Less' : 'Read More'"
                variant="link"
                size="xs"
                :padded="false"
                color="primary"
                class="ml-1 font-bold underline"
                @click="toggleNotesReadMore(row.original.id!)"
              />
            </template>
          </div>
        </template> -->
      
      </UTable>
    </UCard>
</template>

<script setup lang="ts">
import { useAnnualPlanStore } from '~/stores/annual-audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
</script>