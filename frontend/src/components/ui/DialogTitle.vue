<script setup lang="ts">
interface Props {
  class?: string
  asChild?: boolean
}

const props = defineProps<Props>()
</script>

<template>
  <h2
    :class="['dialog-title text-lg font-semibold leading-none tracking-tight text-[var(--app-fg)]', props.class]"
    :aria-level="2"
  >
    <slot />
  </h2>
</template>