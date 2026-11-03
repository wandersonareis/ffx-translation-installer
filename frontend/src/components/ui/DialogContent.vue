<script setup lang="ts">
interface Props {
  class?: string
}

const props = defineProps<Props>()
</script>

<template>
  <div :class="['dialog-content-inner p-6', props.class]">
    <slot />
  </div>
</template>

<style scoped>
.dialog-content-inner {
  max-height: 85vh;
  overflow-y: auto;
}
</style>