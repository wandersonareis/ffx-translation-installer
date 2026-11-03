<script setup lang="ts">
defineProps<{
  asChild?: boolean
}>()

defineSlots<{
  default: [{}]
}>()
</script>

<template>
  <slot />
</template>