<template>
    <div class="flex flex-wrap items-center gap-4">
      <UInput
        v-model="store.searchQuery"
        icon="i-lucide-search"
        placeholder="Search"
        class="w-full max-w-xs"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedType as any)"
        :items="store.qaTypes"
        placeholder="Select QA"
        class="w-full max-w-[200px]"
        size="md"
      />
      <USelectMenu
        v-model="store.selectedPeriod"
        :items="store.periods"
        placeholder="Select Period"
        class="w-full max-w-[200px]"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedStatus as any)"
        :items="store.qaStatuses"
        placeholder="Select Status"
        class="w-full max-w-[200px]"
        size="md"
      />
    </div>
</template>

<script setup lang="ts">
import { useQualityAssuranceStore } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()
</script>