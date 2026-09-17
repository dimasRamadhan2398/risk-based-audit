<template>
  <div class="relative inline-block w-full">
    <UPopover v-model:open="isOpen" :disabled="disabled" class="w-full">
      <!-- Input Field Trigger -->
      <div class="relative flex items-center w-full">
        <input
          :id="id"
          :name="name"
          type="text"
          :value="displayValue"
          :placeholder="placeholder"
          :disabled="disabled"
          :required="required"
          maxlength="10"
          @input="handleTextInput"
          @blur="handleBlur"
          @keydown.down.prevent="isOpen = true"
          :class="[
            'block w-full rounded-md border text-sm transition-colors duration-150',
            'bg-white dark:bg-gray-900',
            'text-gray-900 dark:text-gray-100',
            'placeholder:text-gray-400 dark:placeholder:text-gray-500',
            'border-gray-300 dark:border-gray-700',
            'focus:border-primary-500 focus:ring-1 focus:ring-primary-500 focus:outline-none',
            disabled ? 'opacity-50 cursor-not-allowed bg-gray-100 dark:bg-gray-800' : 'cursor-text',
            sizeClasses,
            $attrs.class
          ]"
        />

        <!-- Action Icons (Clear + Calendar Button) -->
        <div class="absolute inset-y-0 right-2.5 flex items-center gap-1 pointer-events-none">
          <button
            v-if="!disabled && modelValue"
            type="button"
            @click.stop="clearDate"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 p-0.5 rounded focus:outline-none cursor-pointer pointer-events-auto flex items-center justify-center"
            tabindex="-1"
            title="Clear date"
          >
            <UIcon name="i-lucide-x" class="w-3.5 h-3.5" />
          </button>
          <div
            class="text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 p-0.5 rounded focus:outline-none disabled:cursor-not-allowed cursor-pointer pointer-events-auto flex items-center justify-center"
            title="Open calendar"
          >
            <UIcon name="i-lucide-calendar" class="w-4 h-4" />
          </div>
        </div>
      </div>

      <template #content>
        <div class="w-72 p-3 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-xl">
          <!-- Calendar Header (Month/Year Navigation) -->
          <div class="flex items-center justify-between mb-3 px-1">
            <div class="flex items-center gap-1">
              <select
                v-model="viewMonth"
                class="text-xs font-semibold bg-transparent text-gray-800 dark:text-gray-200 rounded px-1.5 py-1 hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer focus:outline-none"
              >
                <option v-for="(m, idx) in monthNames" :key="idx" :value="idx" class="dark:bg-gray-900">
                  {{ m }}
                </option>
              </select>
              <select
                v-model="viewYear"
                class="text-xs font-semibold bg-transparent text-gray-800 dark:text-gray-200 rounded px-1.5 py-1 hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer focus:outline-none"
              >
                <option v-for="y in yearOptions" :key="y" :value="y" class="dark:bg-gray-900">
                  {{ y }}
                </option>
              </select>
            </div>

            <div class="flex items-center gap-0.5">
              <button
                type="button"
                @click.stop="prevMonth"
                class="p-1 rounded-md text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors cursor-pointer"
                title="Previous month"
              >
                <UIcon name="i-lucide-chevron-left" class="w-4 h-4" />
              </button>
              <button
                type="button"
                @click.stop="nextMonth"
                class="p-1 rounded-md text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors cursor-pointer"
                title="Next month"
              >
                <UIcon name="i-lucide-chevron-right" class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Days of Week Header -->
          <div class="grid grid-cols-7 gap-1 mb-1 text-center">
            <span
              v-for="(day, idx) in dayNames"
              :key="idx"
              class="text-[11px] font-medium text-gray-400 dark:text-gray-500 py-0.5"
            >
              {{ day }}
            </span>
          </div>

          <!-- Days Grid -->
          <div class="grid grid-cols-7 gap-1 text-center">
            <button
              v-for="(dayObj, idx) in calendarDays"
              :key="idx"
              type="button"
              :disabled="dayObj.disabled"
              @click.stop="selectDate(dayObj.date)"
              :class="[
                'h-7 w-7 mx-auto rounded-lg text-xs font-medium flex items-center justify-center transition-all duration-100',
                dayObj.isCurrentMonth ? 'text-gray-800 dark:text-gray-200' : 'text-gray-300 dark:text-gray-600',
                dayObj.isSelected
                  ? 'bg-primary-600 text-white font-bold shadow-sm hover:bg-primary-700 dark:bg-primary-500'
                  : dayObj.isToday
                  ? 'border border-primary-500 text-primary-600 dark:text-primary-400 font-semibold hover:bg-primary-50 dark:hover:bg-primary-950/40'
                  : 'hover:bg-gray-100 dark:hover:bg-gray-800',
                dayObj.disabled ? 'opacity-30 cursor-not-allowed hover:bg-transparent' : 'cursor-pointer'
              ]"
            >
              {{ dayObj.dayNumber }}
            </button>
          </div>

          <!-- Footer Shortcuts -->
          <div class="mt-3 pt-2 border-t border-gray-100 dark:border-gray-800 flex justify-between items-center px-1">
            <button
              type="button"
              @click.stop="selectToday"
              class="text-xs font-semibold text-primary-600 dark:text-primary-400 hover:underline cursor-pointer"
            >
              Hari Ini
            </button>
            <button
              v-if="modelValue"
              type="button"
              @click.stop="clearDate"
              class="text-xs text-gray-400 hover:text-error-600 dark:hover:text-error-400 cursor-pointer"
            >
              Hapus
            </button>
          </div>
        </div>
      </template>
    </UPopover>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  format,
  parseISO,
  isValid,
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  eachDayOfInterval,
  isSameDay,
  isSameMonth,
  isToday as checkIsToday,
  addMonths,
  subMonths,
  setMonth,
  setYear,
  parse
} from 'date-fns'

const props = withDefaults(
  defineProps<{
    modelValue?: string | null // Standard ISO format: 'YYYY-MM-DD'
    placeholder?: string
    disabled?: boolean
    required?: boolean
    id?: string
    name?: string
    size?: 'sm' | 'md' | 'lg'
    minDate?: string // 'YYYY-MM-DD'
    maxDate?: string // 'YYYY-MM-DD'
  }>(),
  {
    modelValue: '',
    placeholder: 'dd/mm/yyyy',
    disabled: false,
    required: false,
    size: 'md'
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}>()

const isOpen = ref(false)

// Internal text shown in input (formatted as DD/MM/YYYY)
const displayValue = ref('')

// Calendar view state
const currentDate = ref(new Date())
const viewMonth = ref(currentDate.value.getMonth())
const viewYear = ref(currentDate.value.getFullYear())

// Sync modelValue -> displayValue
watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal) {
      const parsed = parseISO(newVal)
      if (isValid(parsed)) {
        displayValue.value = format(parsed, 'dd/MM/yyyy')
        viewMonth.value = parsed.getMonth()
        viewYear.value = parsed.getFullYear()
        return
      }
    }
    displayValue.value = ''
  },
  { immediate: true }
)

// When popover opens, sync viewing calendar to selected date if exists
watch(
  () => isOpen.value,
  (open) => {
    if (open && props.modelValue) {
      const parsed = parseISO(props.modelValue)
      if (isValid(parsed)) {
        viewMonth.value = parsed.getMonth()
        viewYear.value = parsed.getFullYear()
      }
    }
  }
)

// Size classes
const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'py-1 pl-2.5 pr-14 text-xs'
    case 'lg':
      return 'py-2.5 pl-3.5 pr-16 text-base'
    default:
      return 'py-1.5 pl-3 pr-14 text-sm'
  }
})

// Localization & Calendar constants
const monthNames = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
]
const dayNames = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']

// Year options (-50 to +20 years)
const yearOptions = computed(() => {
  const currentY = new Date().getFullYear()
  const years: number[] = []
  for (let y = currentY - 50; y <= currentY + 20; y++) {
    years.push(y)
  }
  return years
})

// Current viewing month date object
const viewDate = computed(() => {
  return setYear(setMonth(new Date(), viewMonth.value), viewYear.value)
})

// Generate calendar days
const calendarDays = computed(() => {
  const vDate = viewDate.value
  const monthStart = startOfMonth(vDate)
  const monthEnd = endOfMonth(vDate)
  const startDate = startOfWeek(monthStart, { weekStartsOn: 0 })
  const endDate = endOfWeek(monthEnd, { weekStartsOn: 0 })

  const days = eachDayOfInterval({ start: startDate, end: endDate })

  const selectedDateObj = props.modelValue ? parseISO(props.modelValue) : null
  const minDateObj = props.minDate ? parseISO(props.minDate) : null
  const maxDateObj = props.maxDate ? parseISO(props.maxDate) : null

  return days.map(day => {
    let disabled = false
    if (minDateObj && isValid(minDateObj) && day < minDateObj) disabled = true
    if (maxDateObj && isValid(maxDateObj) && day > maxDateObj) disabled = true

    return {
      date: day,
      dayNumber: format(day, 'd'),
      isCurrentMonth: isSameMonth(day, vDate),
      isToday: checkIsToday(day),
      isSelected: selectedDateObj && isValid(selectedDateObj) ? isSameDay(day, selectedDateObj) : false,
      disabled
    }
  })
})

const prevMonth = () => {
  const newDate = subMonths(viewDate.value, 1)
  viewMonth.value = newDate.getMonth()
  viewYear.value = newDate.getFullYear()
}

const nextMonth = () => {
  const newDate = addMonths(viewDate.value, 1)
  viewMonth.value = newDate.getMonth()
  viewYear.value = newDate.getFullYear()
}

const selectDate = (date: Date) => {
  const isoStr = format(date, 'yyyy-MM-dd')
  emit('update:modelValue', isoStr)
  emit('change', isoStr)
  isOpen.value = false
}

const selectToday = () => {
  const today = new Date()
  viewMonth.value = today.getMonth()
  viewYear.value = today.getFullYear()
  selectDate(today)
}

const clearDate = () => {
  emit('update:modelValue', '')
  emit('change', '')
  displayValue.value = ''
  isOpen.value = false
}

// Auto format & mask DD/MM/YYYY when typing
const handleTextInput = (e: Event) => {
  const input = e.target as HTMLInputElement
  let raw = input.value.replace(/\D/g, '') // Only digits

  if (raw.length > 8) {
    raw = raw.slice(0, 8)
  }

  // Format as DD/MM/YYYY
  let formatted = raw
  if (raw.length >= 5) {
    formatted = `${raw.slice(0, 2)}/${raw.slice(2, 4)}/${raw.slice(4)}`
  } else if (raw.length >= 3) {
    formatted = `${raw.slice(0, 2)}/${raw.slice(2)}`
  }

  displayValue.value = formatted

  // If complete (10 characters: DD/MM/YYYY), validate & emit
  if (formatted.length === 10) {
    const parsed = parse(formatted, 'dd/MM/yyyy', new Date())
    if (isValid(parsed)) {
      const iso = format(parsed, 'yyyy-MM-dd')
      viewMonth.value = parsed.getMonth()
      viewYear.value = parsed.getFullYear()
      emit('update:modelValue', iso)
      emit('change', iso)
    }
  } else if (formatted.length === 0) {
    emit('update:modelValue', '')
    emit('change', '')
  }
}

const handleBlur = () => {
  // If user left incomplete date, revert to current modelValue
  if (props.modelValue) {
    const parsed = parseISO(props.modelValue)
    if (isValid(parsed)) {
      displayValue.value = format(parsed, 'dd/MM/yyyy')
      return
    }
  }
  if (!props.modelValue) {
    displayValue.value = ''
  }
}
</script>
