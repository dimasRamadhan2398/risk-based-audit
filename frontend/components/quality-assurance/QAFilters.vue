<template>
  <div class="flex flex-wrap items-center gap-3 sm:gap-4">
    <UInput
      v-model="store.searchQuery"
      icon="i-lucide-search"
      placeholder="Search"
      class="w-full sm:max-w-md"
      size="md"
    />

    <USelectMenu
      v-model="(store.selectedType as any)"
      :items="store.qaTypes"
      placeholder="Select QA"
      class="w-full sm:w-48"
      size="md"
    />

    <USelectMenu
      v-model="store.selectedPeriod"
      :items="store.periods"
      placeholder="Select Period"
      class="w-full sm:w-48"
      size="md"
    />

    <USelectMenu
      v-model="(store.selectedStatus as any)"
      :items="store.qaStatuses"
      placeholder="Select Status"
      class="w-full sm:w-48"
      size="md"
    />

    <UButton
      label="Reset Filter"
      icon="i-lucide-rotate-ccw"
      color="neutral"
      variant="outline"
      size="md"
      class="w-full sm:w-auto"
      @click="store.resetFilters"
    />
  </div>
</template>

<script setup lang="ts">
import { useQualityAssuranceStore } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()
</script>