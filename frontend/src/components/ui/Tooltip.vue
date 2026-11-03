<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

interface Props {
  delayDuration?: number
  skipDelayDuration?: number
}

interface Emits {
  'update:open': [value: boolean]
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const open = ref(false)
const triggerRef = ref<HTMLElement | null>(null)
const contentRef = ref<HTMLElement | null>(null)
let hideTimeout: ReturnType<typeof setTimeout> | null = null

const show = () => {
  if (hideTimeout) {
    clearTimeout(hideTimeout)
    hideTimeout = null
  }
  open.value = true
}

const hide = () => {
  if (props.delayDuration && props.delayDuration > 0) {
    hideTimeout = setTimeout(() => {
      open.value = false
    }, props.delayDuration)
  } else {
    open.value = false
  }
}

onUnmounted(() => {
  if (hideTimeout) {
    clearTimeout(hideTimeout)
  }
})
</script>

<template>
  <div
    class="relative inline-flex"
    @mouseenter="show"
    @mouseleave="hide"
    @focusin="show"
    @focusout="hide"
    ref="triggerRef"
  >
    <slot />

    <Teleport to="body">
      <div
        v-show="open"
        ref="contentRef"
        class="absolute z-50"
      >
        <div class="tooltip-content">
          <slot name="content" />
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.tooltip-content {
  background: var(--app-surface);
  border: 1px solid var(--app-border);
  border-radius: 0.375rem;
  padding: 0.5rem;
  font-size: 0.875rem;
  color: var(--app-fg);
  white-space: nowrap;
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
}
</style>