<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  imageUrl: string
}>()

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'confirm', payload: { file: File; previewUrl: string }): void
}>()

const scale = ref(1)
const offsetX = ref(0)
const offsetY = ref(0)
const dragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const imageRef = ref<HTMLImageElement | null>(null)

watch(() => props.visible, (visible) => {
  if (!visible) return
  scale.value = 1
  offsetX.value = 0
  offsetY.value = 0
})

const imageStyle = computed(() => ({
  transform: `translate(${offsetX.value}px, ${offsetY.value}px) scale(${scale.value})`
}))

function startDrag(event: MouseEvent) {
  dragging.value = true
  dragStartX.value = event.clientX - offsetX.value
  dragStartY.value = event.clientY - offsetY.value
}

function onDrag(event: MouseEvent) {
  if (!dragging.value) return
  offsetX.value = event.clientX - dragStartX.value
  offsetY.value = event.clientY - dragStartY.value
}

function stopDrag() {
  dragging.value = false
}

async function confirmCrop() {
  if (!imageRef.value) return
  const canvas = document.createElement('canvas')
  canvas.width = 320
  canvas.height = 320
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const image = imageRef.value
  const size = 320
  const baseWidth = image.naturalWidth
  const baseHeight = image.naturalHeight
  const drawWidth = baseWidth * scale.value
  const drawHeight = baseHeight * scale.value
  const x = (size - drawWidth) / 2 + offsetX.value
  const y = (size - drawHeight) / 2 + offsetY.value
  ctx.drawImage(image, x, y, drawWidth, drawHeight)

  canvas.toBlob((blob) => {
    if (!blob) return
    const file = new File([blob], 'avatar.png', { type: 'image/png' })
    const previewUrl = URL.createObjectURL(blob)
    emit('confirm', { file, previewUrl })
  }, 'image/png')
}
</script>

<template>
  <div v-if="visible" class="cropper-mask" @mousemove="onDrag" @mouseup="stopDrag" @mouseleave="stopDrag">
    <div class="cropper-dialog card apple-panel">
      <div class="cropper-head">
        <div>
          <p class="apple-label">Avatar Cropper</p>
          <h3>裁剪头像</h3>
        </div>
        <button class="apple-button secondary" type="button" @click="$emit('close')">取消</button>
      </div>
      <div class="cropper-frame">
        <img ref="imageRef" :src="imageUrl" alt="crop preview" class="crop-image" :style="imageStyle" @mousedown.prevent="startDrag" />
      </div>
      <label class="cropper-slider">
        <span>缩放</span>
        <input v-model="scale" type="range" min="0.8" max="2.5" step="0.01" />
      </label>
      <div class="cropper-actions">
        <button class="apple-button" type="button" @click="confirmCrop">确认裁剪</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cropper-mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}

.cropper-dialog {
  width: min(92vw, 520px);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.cropper-head,
.cropper-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.cropper-frame {
  width: 320px;
  height: 320px;
  margin: 0 auto;
  border-radius: 28px;
  overflow: hidden;
  position: relative;
  background: rgba(15, 23, 42, 0.08);
}

.crop-image {
  position: absolute;
  left: 0;
  top: 0;
  width: 320px;
  height: 320px;
  object-fit: cover;
  cursor: grab;
  user-select: none;
}

.cropper-slider {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
