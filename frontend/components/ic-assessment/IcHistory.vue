<template>
  <div>
    <div v-if="loading" class="space-y-3">
      <USkeleton v-for="i in 4" :key="i" class="h-14 w-full rounded-xl" />
    </div>
    <p v-else-if="!events.length" class="text-sm text-gray-500 dark:text-gray-400">{{ t('icAssessment.history.empty') }}</p>
    <ol v-else class="relative border-s border-gray-200 dark:border-gray-800 ms-2 space-y-5">
      <li v-for="e in events" :key="e.id" class="ms-5">
        <span class="absolute -start-1.5 mt-1.5 size-3 rounded-full border border-white dark:border-gray-900 bg-primary-500" />
        <div class="flex flex-wrap items-baseline gap-x-2">
          <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ actionLabel(e.action) }}</span>
          <span v-if="e.actorName" class="text-sm text-gray-600 dark:text-gray-300">{{ t('icAssessment.history.by', { name: e.actorName }) }}</span>
          <time v-if="e.createdAt" class="text-xs text-gray-500 dark:text-gray-400">{{ formatDateTime(e.createdAt) }}</time>
        </div>
        <p v-if="e.fromStatus && e.toStatus && e.fromStatus !== e.toStatus" class="text-xs text-gray-600 dark:text-gray-300">
          {{ t('icAssessment.history.transition', { from: statusLabel(e.fromStatus), to: statusLabel(e.toStatus) }) }}
        </p>
        <p v-if="e.reason" class="text-sm text-gray-800 dark:text-gray-200 mt-1 whitespace-pre-line">
          <span class="font-semibold">{{ t('icAssessment.history.reason') }}:</span> {{ e.reason }}
        </p>
        <details v-if="changeKeys(e).length" class="mt-1">
          <summary class="text-xs text-primary-600 dark:text-primary-400 cursor-pointer">{{ t('icAssessment.history.changes') }} ({{ changeKeys(e).length }})</summary>
          <dl class="mt-1 grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5 text-xs">
            <template v-for="k in changeKeys(e)" :key="k">
              <dt class="font-mono text-gray-500 dark:text-gray-400">{{ k }}</dt>
              <dd class="text-gray-800 dark:text-gray-200 break-words">{{ show(e.before?.[k]) }} &rarr; {{ show(e.after?.[k]) }}</dd>
            </template>
          </dl>
        </details>
      </li>
    </ol>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '~/composables/useI18n'
import type { IcEvent } from '~/utils/icAssessment'

defineProps<{
  events: IcEvent[]
  loading: boolean
}>()

const { t, locale } = useI18n()

const translatedOr = (key: string, fallback: string) => {
  const v = t(key)
  return v === key ? fallback : v
}
const actionLabel = (action: string) => translatedOr(`icAssessment.history.action.${action}`, action)
const statusLabel = (status: string) => translatedOr(`icAssessment.status.${status.toUpperCase()}`, status)

const changeKeys = (e: IcEvent) => Array.from(new Set([...Object.keys(e.before ?? {}), ...Object.keys(e.after ?? {})])).slice(0, 50)

const show = (v: unknown): string => {
  if (v === null || v === undefined || v === '') return '-'
  return typeof v === 'object' ? JSON.stringify(v) : String(v)
}

const formatDateTime = (iso: string) => {
  const d = new Date(iso)
  return isNaN(d.getTime()) ? iso : d.toLocaleString(locale.value === 'en' ? 'en-US' : 'id-ID', { dateStyle: 'medium', timeStyle: 'short' })
}
</script>
