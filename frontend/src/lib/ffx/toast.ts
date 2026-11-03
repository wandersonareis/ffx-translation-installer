import { defineStore } from 'pinia'
import { ref } from 'vue'

interface Toast {
  id: number
  type: 'info' | 'success' | 'warning' | 'error'
  message: string
  duration: number
}

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])
  let toastId = 0

  function showToast(type: Toast['type'], message: string, duration = 4000) {
    const id = ++toastId
    toasts.value.push({ id, type, message, duration })

    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }

    return id
  }

  function removeToast(id: number) {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index !== -1) {
      toasts.value.splice(index, 1)
    }
  }

  return {
    toasts,
    info: (message: string, options?: { duration?: number }) => showToast('info', message, options?.duration ?? 4000),
    success: (message: string, options?: { duration?: number }) => showToast('success', message, options?.duration ?? 4000),
    warning: (message: string, options?: { duration?: number }) => showToast('warning', message, options?.duration ?? 4000),
    error: (message: string, options?: { duration?: number }) => showToast('error', message, options?.duration ?? 4000),
    remove: removeToast,
  }
})