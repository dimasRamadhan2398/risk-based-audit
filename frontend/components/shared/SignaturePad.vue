<template>
  <div class="signature-pad-container space-y-2">
    <!-- Mode Tabs -->
    <div class="flex items-center justify-between gap-2 border-b border-gray-100 dark:border-gray-800 pb-2">
      <div class="flex items-center gap-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg text-xs">
        <button
          type="button"
          class="px-2.5 py-1 rounded-md font-medium transition-all"
          :class="mode === 'draw' ? 'bg-white dark:bg-gray-700 text-primary-600 dark:text-primary-400 shadow-sm' : 'text-gray-500 hover:text-gray-900 dark:hover:text-white'"
          @click="setMode('draw')"
        >
          <UIcon name="i-heroicons-pencil" class="size-3.5 inline mr-1" />
          Gambar
        </button>
        <button
          type="button"
          class="px-2.5 py-1 rounded-md font-medium transition-all"
          :class="mode === 'upload' ? 'bg-white dark:bg-gray-700 text-primary-600 dark:text-primary-400 shadow-sm' : 'text-gray-500 hover:text-gray-900 dark:hover:text-white'"
          @click="setMode('upload')"
        >
          <UIcon name="i-heroicons-arrow-up-tray" class="size-3.5 inline mr-1" />
          Upload
        </button>
      </div>

      <div class="flex items-center gap-1.5">
        <UBadge
          v-if="hasSignature"
          color="success"
          variant="subtle"
          size="xs"
          class="flex items-center gap-1"
        >
          <UIcon name="i-heroicons-check-circle" class="size-3" />
          Ditandatangani
        </UBadge>
        <UButton
          v-if="hasSignature"
          size="xs"
          color="neutral"
          variant="ghost"
          icon="i-heroicons-trash"
          label="Hapus"
          @click="clearSignature"
        />
      </div>
    </div>

    <!-- Draw Mode Canvas -->
    <div v-show="mode === 'draw'" class="relative">
      <div
        class="border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-xl bg-white dark:bg-gray-900 overflow-hidden relative touch-none hover:border-primary-400 dark:hover:border-primary-600 transition"
        :style="{ height: `${height}px` }"
      >
        <canvas
          ref="canvasRef"
          class="w-full h-full cursor-crosshair block"
          @pointerdown="handlePointerDown"
          @pointermove="handlePointerMove"
          @pointerup="handlePointerUp"
          @pointercancel="handlePointerUp"
        />
        <!-- Signature line guide -->
        <div class="absolute bottom-6 left-6 right-6 border-b border-gray-200 dark:border-gray-800 pointer-events-none flex justify-between items-center text-[10px] text-gray-400">
          <span>Tanda tangan di atas garis</span>
          <UIcon name="i-heroicons-pencil" class="size-3 opacity-40" />
        </div>
      </div>
      <p class="text-[11px] text-gray-400 mt-1 italic">
        Gunakan mouse atau sentuhan jari / stylus untuk menggambar tanda tangan.
      </p>
    </div>

    <!-- Upload Mode Input -->
    <div v-show="mode === 'upload'" class="space-y-3">
      <div class="flex items-center gap-2">
        <input
          ref="fileInputRef"
          type="file"
          accept="image/png,image/jpeg,image/svg+xml"
          class="hidden"
          @change="handleFileUpload"
        />
        <UButton
          size="sm"
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-up-tray"
          label="Pilih File Gambar (.png, .jpg)"
          @click="fileInputRef?.click()"
        />
        <span class="text-xs text-gray-400">Maksimal 2MB, format transparan direkomendasikan</span>
      </div>
      <div
        v-if="modelValue"
        class="border border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-white dark:bg-gray-900 flex items-center justify-center min-h-[90px]"
      >
        <img :src="modelValue" alt="Preview Tanda Tangan" class="max-h-20 max-w-full object-contain" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    height?: number
    defaultName?: string
  }>(),
  {
    modelValue: '',
    height: 120,
    defaultName: ''
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}>()

const mode = ref<'draw' | 'upload'>('draw')
const canvasRef = ref<HTMLCanvasElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const isDrawing = ref(false)
const hasSignature = ref(Boolean(props.modelValue))

let ctx: CanvasRenderingContext2D | null = null
let lastX = 0
let lastY = 0
let lastExportedDataUrl = ''

const setMode = (newMode: 'draw' | 'upload') => {
  mode.value = newMode
  if (newMode === 'draw') {
    nextTick(() => {
      initCanvas()
      if (props.modelValue && props.modelValue !== lastExportedDataUrl) {
        renderImageToCanvas(props.modelValue)
      }
    })
  }
}

const initCanvas = () => {
  const canvas = canvasRef.value
  if (!canvas) return

  const rect = canvas.getBoundingClientRect()
  const dpr = window.devicePixelRatio || 1

  if (rect.width === 0 || rect.height === 0) return

  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr

  ctx = canvas.getContext('2d')
  if (ctx) {
    ctx.scale(dpr, dpr)
    ctx.strokeStyle = '#0f172a'
    ctx.lineWidth = 2.5
    ctx.lineCap = 'round'
    ctx.lineJoin = 'round'
  }
}

const renderImageToCanvas = (dataUrl: string) => {
  if (!dataUrl || !ctx || !canvasRef.value) return
  const img = new Image()
  img.onload = () => {
    if (!ctx || !canvasRef.value) return
    const rect = canvasRef.value.getBoundingClientRect()
    ctx.clearRect(0, 0, rect.width, rect.height)
    // Scale image proportionally to fit canvas without artificial downscale
    const scale = Math.min(rect.width / img.width, rect.height / img.height, 1)
    const w = img.width * scale
    const h = img.height * scale
    const x = (rect.width - w) / 2
    const y = (rect.height - h) / 2
    ctx.drawImage(img, x, y, w, h)
    hasSignature.value = true
  }
  img.src = dataUrl
}

const getPointerPos = (e: PointerEvent) => {
  const canvas = canvasRef.value
  if (!canvas) return { x: 0, y: 0 }
  const rect = canvas.getBoundingClientRect()
  return {
    x: e.clientX - rect.left,
    y: e.clientY - rect.top
  }
}

const handlePointerDown = (e: PointerEvent) => {
  if (!ctx || !canvasRef.value) initCanvas()
  if (!ctx || !canvasRef.value) return

  const pos = getPointerPos(e)
  lastX = pos.x
  lastY = pos.y
  isDrawing.value = true
  canvasRef.value.setPointerCapture(e.pointerId)

  // Draw initial dot
  ctx.beginPath()
  ctx.arc(lastX, lastY, 1.25, 0, Math.PI * 2)
  ctx.fillStyle = ctx.strokeStyle
  ctx.fill()
}

const handlePointerMove = (e: PointerEvent) => {
  if (!isDrawing.value || !ctx) return
  const pos = getPointerPos(e)

  ctx.beginPath()
  ctx.moveTo(lastX, lastY)
  ctx.lineTo(pos.x, pos.y)
  ctx.stroke()

  lastX = pos.x
  lastY = pos.y
  hasSignature.value = true
}

const handlePointerUp = (e: PointerEvent) => {
  if (!isDrawing.value) return
  isDrawing.value = false
  if (canvasRef.value && canvasRef.value.hasPointerCapture(e.pointerId)) {
    canvasRef.value.releasePointerCapture(e.pointerId)
  }

  exportCanvas()
}

const exportCanvas = () => {
  if (!canvasRef.value) return
  const dataUrl = canvasRef.value.toDataURL('image/png')
  lastExportedDataUrl = dataUrl
  hasSignature.value = true
  emit('update:modelValue', dataUrl)
  emit('change', dataUrl)
}

const clearSignature = () => {
  lastExportedDataUrl = ''
  if (ctx && canvasRef.value) {
    const rect = canvasRef.value.getBoundingClientRect()
    ctx.clearRect(0, 0, rect.width, rect.height)
  }
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
  hasSignature.value = false
  emit('update:modelValue', '')
  emit('change', '')
}

const handleFileUpload = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > 2 * 1024 * 1024) {
    alert('File gambar terlalu besar (maksimal 2MB).')
    return
  }

  const reader = new FileReader()
  reader.onload = (event) => {
    const dataUrl = event.target?.result as string
    if (dataUrl) {
      lastExportedDataUrl = dataUrl
      hasSignature.value = true
      emit('update:modelValue', dataUrl)
      emit('change', dataUrl)
    }
  }
  reader.readAsDataURL(file)
}

watch(
  () => props.modelValue,
  (val) => {
    hasSignature.value = Boolean(val)
    // Avoid re-rendering into canvas if this is the exact data URL that the canvas just exported
    if (val && mode.value === 'draw' && val !== lastExportedDataUrl) {
      nextTick(() => {
        renderImageToCanvas(val)
      })
    }
  }
)

onMounted(() => {
  nextTick(() => {
    initCanvas()
    if (props.modelValue) {
      renderImageToCanvas(props.modelValue)
    }
  })
})
</script>
