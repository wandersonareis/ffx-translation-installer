<script setup lang="ts">
import { ref, computed } from 'vue'
import { Check } from 'lucide-vue-next'

interface Props {
  modelValue?: boolean
  disabled?: boolean
  id?: string
  name?: string
  value?: string
  required?: boolean
  ariaLabel?: string
  ariaDescribedBy?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  change: [value: boolean]
}>()

const checked = computed({
  get: () => props.modelValue ?? false,
  set: (value) => emit('update:modelValue', value)
})

const classes = computed(() =>
  'peer relative flex h-4 w-4 shrink-0 cursor-pointer items-center justify-center ' +
  'rounded-sm border border-[var(--app-border)] ring-offset-[var(--app-bg)] focus-visible:outline-none ' +
  'focus-visible:ring-2 focus-visible:ring-purple-500 focus-visible:ring-offset-2 ' +
  'disabled:cursor-not-allowed disabled:opacity-50 ' +
  'data-[state=checked]:bg-purple-600 data-[state=checked]:border-purple-600 ' +
  'data-[state=unchecked]:bg-[var(--app-input-bg)] ' +
  'transition-colors'
)
</script>

<template>
  <div class="relative flex items-center">
    <input
      :id="id"
      :name="name"
      :value="value"
      :required="required"
      type="checkbox"
      class="peer h-4 w-4 shrink-0 rounded-sm border border-[var(--app-border)] appearance-none cursor-pointer"
      :checked="checked"
      :disabled="props.disabled"
      :aria-label="ariaLabel"
      :aria-describedby="ariaDescribedBy"
      @change="emit('change', ($event.target as HTMLInputElement).checked)"
    />
    <div
      :class="classes"
      :data-state="checked ? 'checked' : 'unchecked'"
      aria-hidden="true"
    >
      <Check class="h-3.5 w-3.5 text-white" v-show="checked" />
    </div>
  </div>
</template>

<style scoped>
.peer:checked + div {
  background-color: #9333ea;
  border-color: #9333ea;
}

.peer:focus-visible + div {
  outline: none;
  ring: 2px;
  ring-color: #a855f7;
  ring-offset: 2px;
  ring-offset-color: #030712;
}
</style>