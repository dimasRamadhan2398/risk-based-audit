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
        <div class="sm:flex sm:flex-row-reverse gap-4">
          <UButton
            v-if="store.activeCharter.fileUrl && store.activeCharter.fileUrl !== '#'"
            @click="store.downloadCharter(store.activeCharter.id, store.activeCharter.fileName)"
            icon="i-lucide-download"
            size="md"
            color="primary"
            variant="solid"
            :label="t('auditCharter.card.download')"
          />
          <UButton
            v-if="canManageCharter"
            :label="t('auditCharter.card.edit')"
            @click="store.handleEdit(store.activeCharter)"
            color="primary"
            icon="i-lucide-edit"
            variant="outline"
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
      <div>
        <UCard class="relative group" variant="soft">
          <h3 class="text-lg font-semibold text-gray-700 mb-4">
            {{ t('auditCharter.card.historyTitle') }}
          </h3>
          <UTable
            :data="store.historyCharters"
            :columns="store.columns"
            :empty-state="{
              icon: 'i-lucide-folder-open',
              label: t('auditCharter.card.emptyHistory'),
            }"
            class="w-full text-sm text-left"
          >
            <template #version-cell="{ row }">
              <span class="font-bold text-gray-800">{{
                row.original.version
              }}</span>
            </template>
            <template #title-cell="{ row }">
              <span class="font-bold text-gray-800">{{
                row.original.title
              }}</span>
            </template>
            <template #date-cell="{ row }">
              <span class="font-bold text-gray-800">{{
                row.original.date
              }}</span>
            </template>
            <template #approvedBy-cell="{ row }">
              <span class="font-bold text-gray-800">{{
                row.original.approvedBy
              }}</span>
            </template>
            <template #uploadedBy-cell="{ row }">
              <span class="font-bold text-gray-800">{{
                row.original.uploadedBy
              }}</span>
            </template>
            <template #fileName-cell="{ row }">
              <UButton
                v-if="row.original.fileUrl && row.original.fileUrl !== '#'"
                @click="store.downloadCharter(row.original.id, row.original.fileName)"
                icon="i-lucide-external-link"
                color="primary"
                variant="link"
                size="sm"
                class="p-0 font-bold"
                :label="t('auditCharter.card.viewDocument')"
              />

              <span v-else class="text-gray-400 italic">
                {{ t('auditCharter.card.noFile') }}
              </span>
            </template>
            <template #actions-cell="{ row }">
              <div v-if="canManageCharter" class="flex justify-end gap-2">
                <UButton
                  :label="t('auditCharter.card.edit')"
                  size="md"
                  color="primary"
                  variant="outline"
                  icon="i-lucide-edit"
                  @click="store.handleEdit(row.original)"
                />
                <UButton
                  :label="t('auditCharter.card.delete')"
                  size="md"
                  color="error"
                  variant="outline"
                  icon="i-lucide-trash"
                  @click="confirmDelete(row.original)"
                />
              </div>
            </template>
          </UTable>
        </UCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCharterStore } from '~/stores/charter'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'

const { t } = useI18n()
const store = useCharterStore()
const { canManageCharter } = useRbac()

const confirmDelete = async (item: any) => {
  if (confirm(t('auditCharter.card.deleteConfirm', { title: item.title, version: item.version }))) {
    await store.deleteCharter(item.id || '')
  }
}

onMounted(async () => {
  await store.fetchCharters()
})
</script>