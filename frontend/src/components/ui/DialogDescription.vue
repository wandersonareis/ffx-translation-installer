<script setup lang="ts">
interface Props {
  class?: string
  asChild?: boolean
}

const props = defineProps<Props>()
</script>

<template>
  <p
    :class="['dialog-description text-sm text-[var(--app-fg-muted)]', props.class]"
    aria-live="polite"
  >
    <slot />
  </p>
</template>