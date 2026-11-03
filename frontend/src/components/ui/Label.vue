<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  htmlFor?: string
  class?: string
}

const props = defineProps<Props>()

const classes = computed(() =>
  'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 ' + (props.class || '')
)
</script>

<template>
  <label
    :for="htmlFor"
    :class="classes"
  >
    <slot />
  </label>
</template>