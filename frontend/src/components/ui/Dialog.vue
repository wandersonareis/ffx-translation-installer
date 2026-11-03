<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { X } from 'lucide-vue-next'

interface Props {
  modelValue?: boolean
  class?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const isOpen = computed({
  get: () => props.modelValue ?? false,
  set: (val: boolean) => emit('update:modelValue', val)
})

const handleOverlayClick = () => {
  isOpen.value = false
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
  document.body.style.overflow = 'hidden'
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div v-show="isOpen" class="dialog-overlay fixed inset-0 z-50 flex items-center justify-center">
      <!-- O backdrop cobre o overlay inteiro, então é NELE que mora o
           fechar-por-clique-fora: `@click.self` no overlay nunca dispara,
           porque o alvo do clique é sempre o backdrop. -->
      <div
        class="dialog-backdrop fixed inset-0 bg-black/80 transition-opacity"
        @click="handleOverlayClick"
      />
      <div
        :class="[
          'dialog-content fixed z-50 w-full max-w-lg rounded-lg bg-[var(--app-surface-2)] border border-[var(--app-border)] text-[var(--app-fg)] shadow-xl',
          props.class
        ]"
        role="dialog"
        aria-modal="true"
        @click.stop
      >
        <slot />
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.dialog-overlay {
  animation: fadeIn 150ms ease-out;
}

.dialog-content {
  animation: slideIn 150ms ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>