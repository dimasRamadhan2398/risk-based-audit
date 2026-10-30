<template>
  <div class="flex flex-wrap items-center gap-3 sm:gap-4">
    <CommonSearchInput
      v-model="store.searchQuery"
      :placeholder="t('qualityAssurance.dashboard.filters.search')"
      class="w-full sm:flex-1 sm:min-w-64"
      size="md"
    />

    <USelectMenu
      v-model="(store.selectedType as any)"
      :items="qaTypeOptions"
      value-key="value"
      :placeholder="t('qualityAssurance.dashboard.filters.type')"
      class="w-full sm:w-48"
      size="md"
    />

    <USelectMenu
      v-model="store.selectedPeriod"
      :items="store.periods"
      :placeholder="t('qualityAssurance.dashboard.filters.period')"
      class="w-full sm:w-48"
      size="md"
    />

    <USelectMenu
      v-model="(store.selectedStatus as any)"
      :items="qaStatusOptions"
      value-key="value"
      :placeholder="t('qualityAssurance.dashboard.filters.status')"
      class="w-full sm:w-48"
      size="md"
    />

    <UButton
      :label="t('qualityAssurance.dashboard.filters.reset')"
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
import { useQualityAssuranceStore, QAStatus, QAType } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()
const { t } = useI18n()

const qaTypeOptions = computed(() => [
  { label: t('qualityAssurance.types.regular'), value: QAType.REGULAR },
  { label: t('qualityAssurance.types.saiv'), value: QAType.SAIV },
  { label: t('qualityAssurance.types.qar'), value: QAType.QAR },
  { label: t('qualityAssurance.types.iacm'), value: QAType.IACM }
])

const qaStatusOptions = computed(() => [
  { label: t('qualityAssurance.statuses.completed'), value: QAStatus.COMPLETED },
  { label: t('qualityAssurance.statuses.verified'), value: QAStatus.VERIFIED },
  { label: t('qualityAssurance.statuses.inProgress'), value: QAStatus.IN_PROGRESS },
  { label: t('qualityAssurance.statuses.planned'), value: QAStatus.PLANNED }
])
</script>
