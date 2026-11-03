<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  asChild?: boolean
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
}

const props = defineProps<Props>()

const baseClasses = 'inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-purple-500/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--app-bg)] disabled:pointer-events-none disabled:opacity-50'

const variants = {
  default: 'bg-purple-600 text-white hover:bg-purple-700 shadow-sm',
  destructive: 'bg-red-600 text-white hover:bg-red-700 shadow-sm',
  outline: 'border border-[var(--app-border)] bg-transparent hover:bg-[var(--app-hover)]',
  secondary: 'bg-[var(--app-surface)] text-[var(--app-fg)] hover:bg-[var(--app-hover)]',
  ghost: 'hover:bg-[var(--app-hover)]',
  link: 'text-[#38bdf8] underline-offset-4 hover:underline hover:text-[#7dd3fc]',
} as const

const sizes = {
  default: 'h-10 px-4 py-2',
  sm: 'h-9 rounded-md px-3',
  lg: 'h-11 rounded-md px-8',
  icon: 'h-10 w-10',
} as const

const classes = computed(() => {
  return `${baseClasses} ${variants[props.variant || 'default']} ${sizes[props.size || 'default']}`
})
</script>

<template>
  <button
    :class="classes"
    :disabled="disabled"
    :type="type"
    :aria-disabled="disabled"
  >
    <slot />
  </button>
</template>