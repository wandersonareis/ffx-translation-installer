<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'

interface Props {
  delayDuration?: number
  skipDelayDuration?: number
  class?: string
  side?: 'top' | 'right' | 'bottom' | 'left'
  align?: 'start' | 'center' | 'end'
  sideOffset?: number
  alignOffset?: number
}

const props = defineProps<Props>()
const contentRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLElement | null>(null)

const positionContent = () => {
  if (!triggerRef.value || !contentRef.value) return

  const triggerRect = triggerRef.value.getBoundingClientRect()
  const contentRect = contentRef.value.getBoundingClientRect()
  const side = props.side || 'top'
  const align = props.align || 'center'
  const sideOffset = props.sideOffset || 8
  const alignOffset = props.alignOffset || 0

  let top = 0
  let left = 0

  switch (side) {
    case 'top':
      top = triggerRect.top - contentRect.height - sideOffset
      break
    case 'bottom':
      top = triggerRect.bottom + sideOffset
      break
    case 'left':
      left = triggerRect.left - contentRect.width - sideOffset
      break
    case 'right':
      left = triggerRect.right + sideOffset
      break
  }

  switch (align) {
    case 'start':
      if (side === 'top' || side === 'bottom') {
        left = triggerRect.left + alignOffset
      } else {
        top = triggerRect.top + alignOffset
      }
      break
    case 'end':
      if (side === 'top' || side === 'bottom') {
        left = triggerRect.right - contentRect.width + alignOffset
      } else {
        top = triggerRect.bottom - contentRect.height + alignOffset
      }
      break
    case 'center':
    default:
      if (side === 'top' || side === 'bottom') {
        left = triggerRect.left + (triggerRect.width - contentRect.width) / 2 + alignOffset
      } else {
        top = triggerRect.top + (triggerRect.height - contentRect.height) / 2 + alignOffset
      }
      break
  }

  contentRef.value.style.top = `${top + window.scrollY}px`
  contentRef.value.style.left = `${left + window.scrollX}px`
}

onMounted(() => {
  positionContent()
  window.addEventListener('scroll', positionContent)
  window.addEventListener('resize', positionContent)
})

onUnmounted(() => {
  window.removeEventListener('scroll', positionContent)
  window.removeEventListener('resize', positionContent)
})

watch(() => props.delayDuration, positionContent)
</script>

<template>
  <div
    ref="contentRef"
    :class="[
      'tooltip-content',
      props.class
    ]"
    role="tooltip"
    aria-hidden="false"
  >
    <div class="tooltip-arrow" data-side="top" />
    <slot />
  </div>
</template>

<style scoped>
.tooltip-content {
  background-color: var(--app-surface);
  color: var(--app-fg);
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  white-space: nowrap;
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  border: 1px solid var(--app-border);
}

.tooltip-arrow {
  position: absolute;
  width: 8px;
  height: 8px;
  background-color: var(--app-surface);
  border: 1px solid var(--app-border);
  border-right: none;
  border-bottom: none;
  transform: rotate(45deg);
}
</style>