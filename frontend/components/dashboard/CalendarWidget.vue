<template>
  <AuditorWidgetState
    :state="loading ? 'loading' : (errorText ? 'error' : 'ready')"
    :error-text="errorText"
    :retry-label="retryLabel"
    @retry="emit('retry')"
  >
    <div class="grid grid-cols-1 lg:grid-cols-[minmax(0,18rem)_minmax(0,1fr)] gap-4">
      <!-- Mini month -->
      <div class="min-w-0">
        <div class="flex items-center justify-between mb-2">
          <UButton
            icon="i-lucide-chevron-left"
            size="xs"
            variant="ghost"
            color="neutral"
            class="cursor-pointer"
            :aria-label="t('dashboard.calendarWidget.previousMonth')"
            @click="viewMonth = shiftMonth(viewMonth, -1)"
          />
          <button
            type="button"
            class="text-sm font-semibold text-gray-900 dark:text-white capitalize cursor-pointer hover:underline"
            :title="t('dashboard.calendarWidget.today')"
            @click="goToToday"
          >
            {{ monthLabel }}
          </button>
          <UButton
            icon="i-lucide-chevron-right"
            size="xs"
            variant="ghost"
            color="neutral"
            class="cursor-pointer"
            :aria-label="t('dashboard.calendarWidget.nextMonth')"
            @click="viewMonth = shiftMonth(viewMonth, 1)"
          />
        </div>

        <div
          class="grid grid-cols-7 gap-0.5 text-center"
          role="grid"
        >
          <div
            v-for="name in weekdayNames"
            :key="name"
            class="text-[10px] font-semibold uppercase text-gray-400 dark:text-gray-500 py-1"
          >
            {{ name }}
          </div>
          <button
            v-for="cell in grid"
            :key="cell.day"
            type="button"
            class="relative h-9 rounded-md text-xs flex flex-col items-center justify-center cursor-pointer transition-colors"
            :class="dayClass(cell)"
            :aria-pressed="selectedDay === cell.day"
            :aria-label="dayAriaLabel(cell.day)"
            @click="toggleDay(cell.day)"
          >
            <span>{{ Number(cell.day.slice(8)) }}</span>
            <span
              v-if="dayDots(cell.day).length"
              class="absolute bottom-1 flex gap-0.5"
            >
              <span
                v-for="kind in dayDots(cell.day)"
                :key="kind"
                class="size-1 rounded-full"
                :class="KIND_META[kind].dot"
              />
            </span>
          </button>
        </div>
      </div>

      <!-- Agenda -->
      <div class="min-w-0">
        <div class="flex flex-wrap items-center justify-between gap-2 mb-2">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ selectedDay ? formatDay(selectedDay) : t('dashboard.calendarWidget.upcoming', { days: windowDays }) }}
          </h3>
          <div class="flex items-center gap-2">
            <UBadge
              v-if="overdueCount > 0"
              color="error"
              variant="soft"
              size="xs"
            >
              {{ t('dashboard.calendarWidget.overdue', { count: overdueCount }) }}
            </UBadge>
            <UButton
              v-if="selectedDay"
              size="xs"
              variant="link"
              color="neutral"
              class="cursor-pointer"
              @click="selectedDay = null"
            >
              {{ t('dashboard.calendarWidget.showUpcoming') }}
            </UButton>
          </div>
        </div>

        <ul
          v-if="agendaRows.length"
          class="divide-y divide-gray-100 dark:divide-gray-800"
        >
          <li
            v-for="event in agendaRows"
            :key="event.id"
            class="py-2 flex items-start gap-3"
          >
            <div
              class="mt-0.5 p-1.5 rounded-lg shrink-0"
              :class="KIND_META[event.kind].iconBg"
            >
              <UIcon
                :name="KIND_META[event.kind].icon"
                class="size-4"
                :class="KIND_META[event.kind].iconClass"
              />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                {{ event.title }}
              </p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 truncate">
                {{ t(`dashboard.calendarWidget.kinds.${event.kind}`) }}
                · {{ eventDateLabel(event) }}
                <span v-if="event.subtitle"> · {{ event.subtitle }}</span>
              </p>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <UBadge
                v-if="event.overdue"
                color="error"
                variant="subtle"
                size="xs"
              >
                {{ t('dashboard.calendarWidget.overdueBadge') }}
              </UBadge>
              <UButton
                v-if="event.to"
                :to="event.to"
                size="xs"
                variant="ghost"
                color="neutral"
                icon="i-lucide-arrow-right"
                :aria-label="event.title"
              />
            </div>
          </li>
        </ul>
        <p
          v-else
          class="py-6 text-center text-xs text-gray-500 dark:text-gray-400"
        >
          {{ selectedDay ? t('dashboard.calendarWidget.noEventsOnDay') : t('dashboard.calendarWidget.noUpcoming', { days: windowDays }) }}
        </p>
        <p
          v-if="hiddenCount > 0"
          class="pt-1 text-[11px] text-gray-500 dark:text-gray-400"
        >
          {{ t('dashboard.calendarWidget.more', { count: hiddenCount }) }}
        </p>
      </div>
    </div>

    <p
      v-if="note"
      class="mt-3 text-[11px] text-gray-400 dark:text-gray-500"
    >
      {{ note }}
    </p>
  </AuditorWidgetState>
</template>

<script setup lang="ts">
// Mini-month plus agenda for the auditor and auditee dashboards.
// Presentational only: each dashboard builds its own events from data it already loads.
import { computed, ref } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { jakartaToday, formatAtrDate } from '~/utils/actionTakenReport'
import {
  addDays,
  buildMonthGrid,
  compareCalendarEvents,
  eventCoversDay,
  eventEnd,
  eventInWindow,
  shiftMonth,
  type CalendarDayCell,
  type CalendarEvent,
  type CalendarEventKind
} from '~/utils/calendarEvents'
import AuditorWidgetState from './AuditorWidgetState.vue'

const props = withDefaults(defineProps<{
  events: CalendarEvent[]
  loading?: boolean
  errorText?: string
  retryLabel?: string
  note?: string
  windowDays?: number
  maxRows?: number
}>(), {
  loading: false,
  errorText: '',
  retryLabel: '',
  note: '',
  windowDays: 14,
  maxRows: 8
})

const emit = defineEmits<{
  retry: []
}>()

const { t, locale } = useI18n()

const KIND_META: Record<CalendarEventKind, { icon: string, dot: string, iconBg: string, iconClass: string }> = {
  engagement: { icon: 'i-lucide-briefcase', dot: 'bg-sky-500', iconBg: 'bg-sky-500/10', iconClass: 'text-sky-600 dark:text-sky-400' },
  atrDue: { icon: 'i-lucide-list-checks', dot: 'bg-amber-500', iconBg: 'bg-amber-500/10', iconClass: 'text-amber-600 dark:text-amber-400' },
  interview: { icon: 'i-lucide-mic', dot: 'bg-violet-500', iconBg: 'bg-violet-500/10', iconClass: 'text-violet-600 dark:text-violet-400' },
  observation: { icon: 'i-lucide-eye', dot: 'bg-indigo-500', iconBg: 'bg-indigo-500/10', iconClass: 'text-indigo-600 dark:text-indigo-400' },
  sample: { icon: 'i-lucide-flask-conical', dot: 'bg-teal-500', iconBg: 'bg-teal-500/10', iconClass: 'text-teal-600 dark:text-teal-400' },
  pbcDue: { icon: 'i-lucide-file-text', dot: 'bg-orange-500', iconBg: 'bg-orange-500/10', iconClass: 'text-orange-600 dark:text-orange-400' },
  testControlDue: { icon: 'i-lucide-shield-check', dot: 'bg-emerald-500', iconBg: 'bg-emerald-500/10', iconClass: 'text-emerald-600 dark:text-emerald-400' }
}

const intlLocale = computed(() => (locale.value === 'id' ? 'id-ID' : 'en-GB'))

const today = ref(jakartaToday())
const viewMonth = ref(today.value.slice(0, 7))
const selectedDay = ref<string | null>(null)

const grid = computed(() => buildMonthGrid(viewMonth.value))

const monthLabel = computed(() =>
  new Intl.DateTimeFormat(intlLocale.value, { month: 'long', year: 'numeric', timeZone: 'UTC' })
    .format(new Date(`${viewMonth.value}-01T00:00:00Z`))
)

// 2024-01-01 is a Monday; the grid starts on Monday.
const weekdayNames = computed(() => {
  const fmt = new Intl.DateTimeFormat(intlLocale.value, { weekday: 'narrow', timeZone: 'UTC' })
  return Array.from({ length: 7 }, (_, i) => fmt.format(new Date(Date.UTC(2024, 0, 1 + i))))
})

const formatDay = (day: string) => formatAtrDate(day, locale.value)

const eventDateLabel = (event: CalendarEvent) =>
  event.end
    ? t('dashboard.calendarWidget.period', { start: formatDay(event.start), end: formatDay(event.end) })
    : formatDay(event.start)

// Per-day lookups for the visible grid only.
const dayIndex = computed(() => {
  const first = grid.value[0]!.day
  const last = grid.value[grid.value.length - 1]!.day
  const map = new Map<string, { dots: CalendarEventKind[], inRange: boolean, overdue: boolean }>()
  for (const cell of grid.value) map.set(cell.day, { dots: [], inRange: false, overdue: false })
  for (const event of props.events) {
    if (!eventInWindow(event, first, last)) continue
    if (event.end) {
      for (const cell of grid.value) {
        if (eventCoversDay(event, cell.day)) map.get(cell.day)!.inRange = true
      }
      continue
    }
    const entry = map.get(event.start)
    if (!entry) continue
    if (!entry.dots.includes(event.kind)) entry.dots.push(event.kind)
    if (event.overdue) entry.overdue = true
  }
  return map
})

const dayDots = (day: string) => (dayIndex.value.get(day)?.dots || []).slice(0, 3)

const dayClass = (cell: CalendarDayCell) => {
  const info = dayIndex.value.get(cell.day)
  const classes: string[] = []
  if (selectedDay.value === cell.day) {
    classes.push('bg-primary-600 text-white font-semibold')
  } else {
    classes.push(cell.inMonth ? 'text-gray-700 dark:text-gray-200' : 'text-gray-300 dark:text-gray-600')
    classes.push(info?.inRange ? 'bg-sky-50 dark:bg-sky-500/10 hover:bg-sky-100 dark:hover:bg-sky-500/20' : 'hover:bg-gray-100 dark:hover:bg-gray-800')
  }
  if (cell.day === today.value) classes.push('ring-1 ring-primary-500 font-semibold')
  return classes
}

const dayAriaLabel = (day: string) => {
  const count = props.events.filter(e => eventCoversDay(e, day)).length
  return count > 0 ? `${formatDay(day)} (${count})` : formatDay(day)
}

const toggleDay = (day: string) => {
  selectedDay.value = selectedDay.value === day ? null : day
}

const goToToday = () => {
  today.value = jakartaToday()
  viewMonth.value = today.value.slice(0, 7)
  selectedDay.value = null
}

// Upcoming window: today .. today + windowDays - 1, plus anything already overdue.
const agendaEvents = computed(() => {
  if (selectedDay.value) {
    const day = selectedDay.value
    return props.events.filter(e => eventCoversDay(e, day)).sort(compareCalendarEvents)
  }
  const from = today.value
  const to = addDays(from, props.windowDays - 1)
  return props.events
    .filter(e => e.overdue || eventInWindow(e, from, to))
    // A long engagement period is listed once it has started only if it ends inside the window.
    .filter(e => e.overdue || !e.end || e.start >= from || eventEnd(e) <= to)
    .sort(compareCalendarEvents)
})

const agendaRows = computed(() => agendaEvents.value.slice(0, props.maxRows))
const hiddenCount = computed(() => Math.max(0, agendaEvents.value.length - agendaRows.value.length))
const overdueCount = computed(() => props.events.filter(e => e.overdue).length)
</script>
