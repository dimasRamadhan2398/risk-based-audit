<template>
  <component :is="as" class="leading-relaxed inline-block max-w-full">
    <template v-if="!text">
      <span :class="textClass">-</span>
    </template>
    <template v-else-if="!isLong">
      <span :class="textClass" class="break-words whitespace-normal">{{ text }}</span>
    </template>
    <template v-else>
      <span :class="textClass" class="break-words whitespace-normal">
        {{ displayText }}
      </span>
      <button
        type="button"
        class="inline-flex items-center gap-0.5 ml-1 text-xs font-semibold text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 transition-colors duration-150 cursor-pointer select-none focus:outline-none hover:underline"
        :class="buttonClass"
        :aria-expanded="isExpanded"
        @click.stop="toggleExpand"
      >
        <span>{{ isExpanded ? displayLessLabel : displayMoreLabel }}</span>
        <UIcon
          :name="isExpanded ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
          class="w-3.5 h-3.5 transition-transform duration-200 inline-block align-middle"
        />
      </button>
    </template>
  </component>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '~/composables/useI18n'

const props = withDefaults(
  defineProps<{
    text?: string | null
    maxLength?: number
    as?: string
    textClass?: string
    buttonClass?: string
    moreLabel?: string
    lessLabel?: string
  }>(),
  {
    text: '',
    maxLength: 80,
    as: 'div',
    textClass: '',
    buttonClass: '',
    moreLabel: '',
    lessLabel: ''
  }
)

const { t } = useI18n()
const isExpanded = ref(false)

const toggleExpand = () => {
  isExpanded.value = !isExpanded.value
}

const isLong = computed(() => {
  if (!props.text) return false
  return props.text.length > props.maxLength
})

const displayText = computed(() => {
  if (!props.text) return ''
  if (!isLong.value || isExpanded.value) {
    return props.text
  }
  const truncated = props.text.slice(0, props.maxLength).trim()
  return `${truncated}...`
})

const displayMoreLabel = computed(() => {
  if (props.moreLabel) return props.moreLabel
  const translated = t('common.readMore')
  return translated !== 'common.readMore' ? translated : 'Baca selengkapnya'
})

const displayLessLabel = computed(() => {
  if (props.lessLabel) return props.lessLabel
  const translated = t('common.readLess')
  return translated !== 'common.readLess' ? translated : 'Sembunyikan'
})
</script>
