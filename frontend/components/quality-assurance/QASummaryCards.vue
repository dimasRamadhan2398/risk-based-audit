<template>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Reguler Asesmen -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-amber-400"></div>
            <h3 class="text-lg font-bold">Regular Assessment</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Latest Result</span>
              <span class="font-bold">{{ (store.summary.regular as any).result }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Period</span>
              <span class="font-bold">{{ (store.summary.regular as any).period }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Status</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.regular as any).status)]"></div>
                <span class="font-bold text-sm">{{ (store.summary.regular as any).status }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            label="View Detail"
            icon="i-lucide-chevron-right"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.regular)"
          />
        </div>
      </UCard>

      <!-- QAR (PROFESSIONAL REVIEW) -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-white border border-gray-300"></div>
            <h3 class="text-lg font-bold">QAR (Professional Review)</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Overall Conclusion</span>
              <span class="font-bold">{{ formatOverallConclusion((store.summary.qar as any).result) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Cycle</span>
              <span class="font-bold">{{ (store.summary.qar as any).period }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Status</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.qar as any).status)]"></div>
                <span class="font-bold text-sm">{{ (store.summary.qar as any).status === QAStatus.COMPLETED ? 'Verified' : (store.summary.qar as any).status }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            label="View Detail"
            icon="i-lucide-certificate"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.qar)"
          />
        </div>
      </UCard>

      <!-- SAIV (GIAS STANDARD) -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-blue-500"></div>
            <h3 class="text-lg font-bold">SAIV (GIAS STANDARD)</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Overall Conclusion</span>
              <span class="font-bold">{{ formatOverallConclusion((store.summary.saiv as any).result) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Validator</span>
              <span class="font-bold text-sm truncate max-w-[120px]">{{ (store.summary.saiv as any).validator }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Status</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.saiv as any).status)]"></div>
                <span class="font-bold text-sm">{{ (store.summary.saiv as any).status }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            label="View Detail"
            icon="i-lucide-chevron-right"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.saiv)"
          />
        </div>
      </UCard>

      <!-- BUMN IACM (IACM SCALE) -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-indigo-500"></div>
            <h3 class="text-lg font-bold">BUMN IACM Assessment</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Capability Level</span>
              <span class="font-bold text-indigo-600 text-lg">{{ (store.summary.iacm as any).result }} / 5</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Period</span>
              <span class="font-bold">{{ (store.summary.iacm as any).period }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">Status</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.iacm as any).status)]"></div>
                <span class="font-bold text-sm">{{ (store.summary.iacm as any).status }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            label="View Detail"
            icon="i-lucide-chevron-right"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.iacm)"
          />
        </div>
      </UCard>
    </div>
</template>

<script setup lang="ts">
import { useQualityAssuranceStore, QAStatus } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()

const openCardDetail = (report: any) => {
  if (report && report.id) {
    store.openDetail(report)
  }
}

const formatOverallConclusion = (result: string) => {
  if (!result) return '-'
  const res = result.trim().toLowerCase()
  if (res === 'g/c*' || res === 'gc' || res.includes('generally')) {
    return 'Generally Conformed'
  }
  if (res === 'fc' || res.includes('fully') || res.includes('conformance')) {
    return 'Fully Conformance'
  }
  if (res === 'pc' || res.includes('partially')) {
    return 'Partially Conform'
  }
  if (res === 'dnc' || res.includes('does not') || res.includes('doesnot')) {
    return 'Does not Conform'
  }
  return result
}
</script>