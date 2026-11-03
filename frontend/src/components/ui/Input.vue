<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  value?: string
  placeholder?: string
  disabled?: boolean
  readonly?: boolean
  id?: string
  type?: string
  invalid?: boolean
  ariaLabel?: string
  ariaDescribedBy?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:value': [value: string]
  input: [value: string]
  blur: [event: FocusEvent]
  keydown: [event: KeyboardEvent]
}>()

const classes = computed(() =>
  'flex h-10 w-full rounded-md border bg-[var(--app-input-bg)] px-3 py-2 text-sm text-[var(--app-fg)] ring-offset-[var(--app-bg)] ' +
  'file:border-0 file:bg-transparent file:text-sm file:font-medium ' +
  'placeholder:text-[var(--app-fg-muted)] ' +
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 ' +
  'disabled:cursor-not-allowed disabled:opacity-50 ' +
  'transition-colors ' +
  (props.invalid
    ? 'border-red-500 focus-visible:ring-red-500'
    : 'border-[var(--app-input-border)] focus-visible:ring-purple-500')
)
</script>

<template>
  <input
    :class="classes"
    :value="value"
    :placeholder="placeholder"
    :disabled="disabled"
    :readonly="readonly"
    :id="id"
    :type="type || 'text'"
    :aria-label="ariaLabel"
    :aria-describedby="ariaDescribedBy"
    :aria-invalid="invalid || undefined"
    @input="(event) => { const value = (event.target as HTMLInputElement).value; emit('input', value); emit('update:value', value) }"
    @blur="emit('blur', $event)"
    @keydown="emit('keydown', $event)"
  />
</template>
