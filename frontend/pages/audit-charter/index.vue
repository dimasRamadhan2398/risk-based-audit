<template>
  <div class="max-w-full mx-auto space-y-6">
    <UCard :ui="{ body: 'p-0' }" class="overflow-hidden border border-gray-200 dark:border-gray-800">
      <UTabs v-model="activeTab" :items="tabs" class="w-full">
        <template #content="{ item }">
          <div class="p-6 space-y-6">
            <div v-if="item.key === 'charter'">
              <AuditCharterCard />
              <AuditCharterForm />
            </div>

            <div v-else-if="item.key === 'guideline'">
              <GuidelineList />
              <GuidelineForm />
            </div>

            <div v-else-if="item.key === 'sop'">
              <SopList />
              <SopForm />
            </div>
          </div>
        </template>
      </UTabs>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import AuditCharterCard from "~/components/audit-charter/AuditCharterCard.vue";
import AuditCharterForm from "~/components/audit-charter/AuditCharterForm.vue";
import GuidelineList from "~/components/audit-charter/GuidelineList.vue";
import GuidelineForm from "~/components/audit-charter/GuidelineForm.vue";
import SopList from "~/components/audit-charter/SopList.vue";
import SopForm from "~/components/audit-charter/SopForm.vue";
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const activeTab = ref('charter')

const tabs = computed(() => [
  {label: t('auditCharter.tabs.charter'), key: 'charter', value: 'charter', icon: 'i-lucide-file-text'},
  {label: t('auditCharter.tabs.guideline'), key: 'guideline', value: 'guideline', icon: 'i-lucide-book-open'},
  {label: t('auditCharter.tabs.sop'), key: 'sop', value: 'sop', icon: 'i-lucide-file-check'}
])
</script>
