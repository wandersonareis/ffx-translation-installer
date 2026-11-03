<script setup lang="ts">
import { computed } from 'vue'
import { useToastStore } from '@/lib/ffx/toast'
import { X, Info, CheckCircle, AlertCircle, AlertTriangle } from 'lucide-vue-next'

const toast = useToastStore()

const getIcon = (type: string) => {
  switch (type) {
    case 'info': return Info
    case 'success': return CheckCircle
    case 'warning': return AlertTriangle
    case 'error': return AlertCircle
    default: return Info
  }
}

const getClasses = (type: string) => {
  const base = 'flex items-start gap-3 p-4 rounded-lg border shadow-lg min-w-[300px] max-w-md animate-slide-in'
  switch (type) {
    case 'info': return `${base} bg-blue-900/50 border-blue-700 text-blue-100`
    case 'success': return `${base} bg-green-900/50 border-green-700 text-green-100`
    case 'warning': return `${base} bg-yellow-900/50 border-yellow-700 text-yellow-100`
    case 'error': return `${base} bg-red-900/50 border-red-700 text-red-100`
    default: return base
  }
}
</script>

<template>
  <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 pointer-events-none" aria-live="polite">
    <div
      v-for="toastItem in toast.toasts"
      :key="toastItem.id"
      :class="getClasses(toastItem.type)"
      class="pointer-events-auto"
    >
      <component :is="getIcon(toastItem.type)" class="h-5 w-5 shrink-0 mt-0.5" />
      <div class="flex-1 text-sm">{{ toastItem.message }}</div>
      <button
        @click="toast.remove(toastItem.id)"
        class="shrink-0 opacity-50 hover:opacity-100 transition-opacity"
        aria-label="Fechar"
      >
        <X class="h-4 w-4" />
      </button>
    </div>
  </div>
</template>

<style scoped>
@keyframes slide-in {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.animate-slide-in {
  animation: slide-in 300ms ease-out;
}
</style>