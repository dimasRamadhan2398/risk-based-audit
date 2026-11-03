<template>
  <div class="space-y-3">
    <!-- Status row -->
    <div class="flex items-center justify-between gap-2">
      <UBadge
        v-if="isSigned"
        color="success"
        variant="subtle"
        size="xs"
        class="flex items-center gap-1"
      >
        <UIcon
          name="i-heroicons-check-circle"
          class="size-3"
        />
        {{ t('common.signaturePad.signed') }}
      </UBadge>
      <UBadge
        v-else
        color="neutral"
        variant="subtle"
        size="xs"
      >
        {{ t('common.signatureBox.notSigned') }}
      </UBadge>
      <UButton
        v-if="hasInk || isSigned"
        size="xs"
        color="neutral"
        variant="ghost"
        icon="i-heroicons-trash"
        :label="t('common.signaturePad.clear')"
        @click="clear"
      />
    </div>

    <!-- Drawing area: mouse, touch and pen through pointer events -->
    <div
      class="relative rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-600 bg-white overflow-hidden hover:border-primary-400 transition"
      :style="{ height: `${height}px` }"
    >
      <canvas
        ref="canvasRef"
        class="block w-full h-full touch-none cursor-crosshair"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
      />
      <div class="absolute bottom-6 left-6 right-6 border-b border-gray-200 pointer-events-none flex justify-between items-center text-[10px] text-gray-400">
        <span>{{ t('common.signatureBox.signHere') }}</span>
        <UIcon
          name="i-heroicons-pencil"
          class="size-3 opacity-40"
        />
      </div>
    </div>
    <p class="text-[11px] text-gray-400 dark:text-gray-500 italic">
      {{ t('common.signatureBox.hint') }}
    </p>

    <!-- Signer details -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <UFormField
        :label="t('common.signatureBox.name')"
        required
      >
        <UInput
          :model-value="modelValue.name"
          :placeholder="t('common.signatureBox.namePlaceholder')"
          class="w-full"
          @update:model-value="updateField('name', $event)"
        />
      </UFormField>
      <UFormField :label="t('common.signatureBox.title')">
        <UInput
          :model-value="modelValue.title"
          :placeholder="t('common.signatureBox.titlePlaceholder')"
          class="w-full"
          @update:model-value="updateField('title', $event)"
        />
      </UFormField>
    </div>

    <!-- Timestamp: set when the first stroke is drawn -->
    <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
      <UIcon
        name="i-heroicons-clock"
        class="size-4 shrink-0"
      />
      <span>{{ t('common.signatureBox.timestamp') }}:</span>
      <span class="font-mono text-gray-700 dark:text-gray-300">{{ formattedSignedAt || '-' }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { SignatureBoxValue } from '~/types/signature'

const { t, locale } = useI18n()

const props = withDefaults(
  defineProps<{
    modelValue: SignatureBoxValue
    height?: number
  }>(),
  { height: 160 }
)

const emit = defineEmits<{
  'update:modelValue': [value: SignatureBoxValue]
}>()

// Strokes are stored in normalized coordinates (0..1 of the drawing area) so they can be redrawn after a resize.
interface Point { x: number, y: number }

const canvasRef = ref<HTMLCanvasElement | null>(null)
const hasInk = ref(false)
const isSigned = computed(() => Boolean(props.modelValue.image))

let ctx: CanvasRenderingContext2D | null = null
let strokes: Point[][] = []
let currentStroke: Point[] | null = null
let resizeObserver: ResizeObserver | null = null

const INK = '#0f172a'
const LINE_WIDTH = 2.5

const clamp01 = (n: number) => Math.min(1, Math.max(0, n))

const pointFromEvent = (e: PointerEvent): Point | null => {
  const canvas = canvasRef.value
  if (!canvas) return null
  const rect = canvas.getBoundingClientRect()
  if (rect.width === 0 || rect.height === 0) return null
  return {
    x: clamp01((e.clientX - rect.left) / rect.width),
    y: clamp01((e.clientY - rect.top) / rect.height)
  }
}

// Layout size of the canvas, unaffected by CSS transforms (the modal zoom-in animation scales getBoundingClientRect)
const layoutSize = (canvas: HTMLCanvasElement) => ({ w: canvas.clientWidth, h: canvas.clientHeight })

const drawDot = (p: Point, w: number, h: number) => {
  if (!ctx) return
  ctx.beginPath()
  ctx.arc(p.x * w, p.y * h, LINE_WIDTH / 2, 0, Math.PI * 2)
  ctx.fillStyle = INK
  ctx.fill()
}

const drawSegment = (a: Point, b: Point, w: number, h: number) => {
  if (!ctx) return
  if (a.x === b.x && a.y === b.y) {
    drawDot(a, w, h)
    return
  }
  ctx.beginPath()
  ctx.moveTo(a.x * w, a.y * h)
  ctx.lineTo(b.x * w, b.y * h)
  ctx.stroke()
}

const redraw = () => {
  const canvas = canvasRef.value
  if (!ctx || !canvas) return
  const { w, h } = layoutSize(canvas)
  ctx.clearRect(0, 0, w, h)
  for (const stroke of strokes) {
    for (let i = 1; i < stroke.length; i++) {
      drawSegment(stroke[i - 1]!, stroke[i]!, w, h)
    }
    if (stroke.length === 1) drawDot(stroke[0]!, w, h)
  }
}

// Sizes the backing store to the CSS size times devicePixelRatio. Skipped while the box is hidden (size 0).
const setupCanvas = () => {
  const canvas = canvasRef.value
  if (!canvas) return
  const { w, h } = layoutSize(canvas)
  if (w === 0 || h === 0) return
  const dpr = window.devicePixelRatio || 1
  canvas.width = Math.round(w * dpr)
  canvas.height = Math.round(h * dpr)
  ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.strokeStyle = INK
  ctx.fillStyle = INK
  ctx.lineWidth = LINE_WIDTH
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  redraw()
}

const emitValue = () => {
  const image = hasInk.value && canvasRef.value ? canvasRef.value.toDataURL('image/png') : ''
  emit('update:modelValue', {
    ...props.modelValue,
    image,
    // The signing time is fixed at the first stroke; a later edit keeps the original time
    signedAt: image ? (props.modelValue.signedAt || new Date().toISOString()) : ''
  })
}

const onPointerDown = (e: PointerEvent) => {
  if (!ctx) setupCanvas()
  const p = pointFromEvent(e)
  if (!ctx || !p || !canvasRef.value) return
  canvasRef.value.setPointerCapture(e.pointerId)
  currentStroke = [p]
  strokes.push(currentStroke)
  hasInk.value = true
  const { w, h } = layoutSize(canvasRef.value)
  drawDot(p, w, h)
}

const onPointerMove = (e: PointerEvent) => {
  if (!currentStroke || !ctx || !canvasRef.value) return
  const p = pointFromEvent(e)
  const last = currentStroke[currentStroke.length - 1]
  if (!p || !last || (p.x === last.x && p.y === last.y)) return
  currentStroke.push(p)
  const { w, h } = layoutSize(canvasRef.value)
  drawSegment(last, p, w, h)
}

const onPointerUp = (e: PointerEvent) => {
  if (!currentStroke) return
  if (canvasRef.value?.hasPointerCapture(e.pointerId)) {
    canvasRef.value.releasePointerCapture(e.pointerId)
  }
  currentStroke = null
  emitValue()
}

const clear = () => {
  strokes = []
  currentStroke = null
  hasInk.value = false
  if (ctx && canvasRef.value) {
    const { w, h } = layoutSize(canvasRef.value)
    ctx.clearRect(0, 0, w, h)
  }
  emit('update:modelValue', { ...props.modelValue, image: '', signedAt: '' })
}

const updateField = (key: 'name' | 'title', value: string | number | undefined) => {
  emit('update:modelValue', { ...props.modelValue, [key]: String(value ?? '') })
}

const formattedSignedAt = computed(() => {
  if (!props.modelValue.signedAt) return ''
  const d = new Date(props.modelValue.signedAt)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleString(locale.value === 'en' ? 'en-US' : 'id-ID', { dateStyle: 'medium', timeStyle: 'short' })
})

// The parent resets the value (e.g. after a cancelled or completed approval): drop the ink as well.
watch(
  () => props.modelValue.image,
  (image) => {
    if (!image && strokes.length > 0 && !currentStroke) {
      strokes = []
      hasInk.value = false
      redraw()
    }
  }
)

onMounted(() => {
  nextTick(() => setupCanvas())
  if (typeof ResizeObserver !== 'undefined' && canvasRef.value) {
    resizeObserver = new ResizeObserver(() => setupCanvas())
    resizeObserver.observe(canvasRef.value)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
})
</script>
