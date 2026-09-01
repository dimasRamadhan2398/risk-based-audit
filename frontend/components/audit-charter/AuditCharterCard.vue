<template>
  <div>
    <UAlert
      v-if="store.errorMsg"
      color="error"
      variant="soft"
      icon="i-lucide-circle-alert"
      :title="t('auditCharter.card.loadError')"
      :description="store.errorMsg"
      class="mb-4"
    />

    <div v-if="store.loading" class="space-y-3 py-6">
      <USkeleton class="h-24 w-full" />
      <USkeleton class="h-48 w-full" />
    </div>

    <!-- Empty State when there are no charters at all -->
    <div
      v-else-if="store.charters.length === 0"
      class="flex flex-col items-center justify-center p-12 bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-2xl text-center space-y-6 shadow-sm my-4"
    >
      <div class="w-16 h-16 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500">
        <UIcon name="i-lucide-file-text" class="w-8 h-8" />
      </div>
      <div class="space-y-2 max-w-md">
        <h2 class="text-xl font-bold text-[var(--text-main)]">{{ t('auditCharter.card.emptyTitle') }}</h2>
        <p class="text-sm text-[var(--text-muted)] leading-relaxed">
          {{ t('auditCharter.card.emptyDesc') }}
        </p>
      </div>
      <UButton
        v-if="canManageCharter"
        :label="t('auditCharter.card.uploadNew')"
        @click="() => { store.showModal = true }"
        color="primary"
        size="lg"
        class="rounded-xl px-6 py-3 font-semibold transition-all duration-200"
        icon="i-lucide-plus"
      />
    </div>

    <!-- Normal State when charters exist -->
    <div v-else class="space-y-8">
      <!-- Active Charter Card -->
      <UCard
        v-if="store.activeCharter"
        class="relative group"
        variant="soft"
      >
        <div class="flex justify-between items-center">
          <h1 class="text-2xl font-bold text-gray-900">
            {{ t('auditCharter.card.activeTitle') }}
          </h1>
          <UButton
            v-if="canManageCharter"
            :label="t('auditCharter.card.addCharter')"
            @click="() => { store.showModal = true }"
            color="primary"
            icon="i-lucide-plus"
          >
          </UButton>
        </div>

        <div class="border-t border-gray-400 my-4"></div>

        <div class="flex justify-between items-start">
          <div>
            <UBadge
              :label="t('auditCharter.card.currentlyActive')"
              class="px-2.5 py-0.5 mb-4 rounded inline-block"
              size="xl"
              color="success"
            >
            </UBadge>
            <h2 class="text-3xl font-bold text-[var(--text-main)] mb-4">
              {{ store.activeCharter.title }}
            </h2>
            <div class="flex items-center gap-4 text-sm text-gray-500 mb-6">
              <UBadge class="rounded inline-block" size="lg" color="error">
                v{{ store.activeCharter.version }}
              </UBadge>
              <UIcon name="i-lucide-calendar" size="md"></UIcon>
              <span class="flex items-center gap-1">{{
                store.activeCharter.date
              }}</span>
              <UIcon name="i-lucide-file-text" size="md"></UIcon>
              <span class="flex items-center gap-1">{{
                store.activeCharter.fileName
              }}</span>
            </div>

            <div class="grid grid-cols-2 bg-[var(--bg-surface)] gap-8 mb-6 p-4 rounded-lg transition-colors duration-300">
              <div>
                <p class="text-md text-[var(--text-muted)] uppercase tracking-wider">
                  {{ t('auditCharter.card.uploadedBy') }}
                </p>
                <p class="font-medium text-[var(--text-main)]">
                  {{ store.activeCharter.uploadedBy }}
                </p>
              </div>
              <div>
                <p class="text-md text-[var(--text-muted)] uppercase tracking-wider">
                  {{ t('auditCharter.card.approvedBy') }}
                </p>
                <p class="font-medium text-[var(--text-main)]">
                  {{ store.activeCharter.approvedBy }}
                </p>
              </div>
            </div>
          </div>
        </div>
        <div class="sm:flex sm:flex-row-reverse">
          <UButton
            v-if="store.activeCharter.fileUrl && store.activeCharter.fileUrl !== '#'"
            :to="store.activeCharter.fileUrl"
            target="_blank"
            icon="i-lucide-eye"
            size="xl"
            color="neutral"
            variant="ghost"
          />
          <UButton
            v-if="canManageCharter"
            :label="t('auditCharter.card.edit')"
            @click="store.handleEdit(store.activeCharter)"
            color="warning"
            icon="i-lucide-edit"
            variant="ghost"
            size="xl"
          >
          </UButton>
        </div>
      </UCard>

      <!-- Warning card when there is no active charter (but there is history) -->
      <UCard v-else class="bg-secondary-50 border-l-4 border-secondary-400 p-4">
        <div class="flex justify-between items-center">
          <UIcon name="i-lucide-alert-triangle" size="lg" class="text-warning-600"></UIcon>
          <p class="text-secondary-700">
            {{ t('auditCharter.card.noActiveCharterWarning') }}
          </p>
          <UButton
            v-if="canManageCharter"
            :label="t('auditCharter.card.addCharter')"
            @click="() => { store.showModal = true }"
            color="primary"
            icon="i-lucide-plus"
          >
          </UButton>
        </div>
      </UCard>

      <!-- History Table -->
      <div class="space-y-3">
        <h3 class="text-lg font-semibold text-[var(--text-main)]">
          {{ t('auditCharter.card.historyTitle') }}
        </h3>
        <TableEntities
          :data="store.historyCharters"
          :columns="store.columns"
          :empty-state="{
            icon: 'i-lucide-folder-open',
            label: t('auditCharter.card.emptyHistory'),
          }"
          class="w-full"
        >
          <template #version-cell="{ row }">
            <span class="font-bold text-[var(--text-main)] whitespace-nowrap">{{
              row.original.version
            }}</span>
          </template>
          <template #title-cell="{ row }">
            <div class="font-bold text-[var(--text-main)] break-words whitespace-normal leading-relaxed min-w-0">
              {{ row.original.title }}
            </div>
          </template>
          <template #content-cell="{ row }">
            <div class="font-normal text-[var(--text-muted)] text-sm break-words whitespace-normal leading-relaxed min-w-0">
              {{ row.original.content || '-' }}
            </div>
          </template>
          <template #date-cell="{ row }">
            <span class="font-medium text-[var(--text-main)] whitespace-nowrap">{{
              row.original.date
            }}</span>
          </template>
          <template #approvedBy-cell="{ row }">
            <div class="font-medium text-[var(--text-main)] break-words whitespace-normal leading-relaxed min-w-0">
              {{ row.original.approvedBy }}
            </div>
          </template>
          <template #uploadedBy-cell="{ row }">
            <div class="font-medium text-[var(--text-main)] break-words whitespace-normal leading-relaxed min-w-0">
              {{ row.original.uploadedBy }}
            </div>
          </template>
          <template #actions-cell="{ row }">
            <div class="flex justify-end gap-1.5 whitespace-nowrap">
              <UButton
                v-if="row.original.fileUrl && row.original.fileUrl !== '#'"
                :to="row.original.fileUrl"
                target="_blank"
                size="md"
                color="neutral"
                variant="ghost"
                icon="i-lucide-eye"
              />
              <UButton
                v-if="canManageCharter"
                size="md"
                color="warning"
                variant="ghost"
                icon="i-lucide-edit"
                @click="store.handleEdit(row.original)"
              />
              <UButton
                v-if="canManageCharter"
                size="md"
                color="error"
                variant="ghost"
                icon="i-lucide-trash-2"
                @click="confirmDelete(row.original)"
              />
            </div>
          </template>
        </TableEntities>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCharterStore } from '~/stores/charter'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import TableEntities from '~/components/shared/TableEntities.vue'

const { t } = useI18n()
const store = useCharterStore()
const { canManageCharter } = useRbac()

const confirmDelete = async (item: any) => {
  if (await useGlobalModalStore().confirmDelete({ description: t('auditCharter.card.deleteConfirm', { title: item.title, version: item.version }) })) {
    await store.deleteCharter(item.id || '')
  }
}

onMounted(async () => {
  await store.fetchCharters()
})
</script>