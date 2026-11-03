<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { X } from 'lucide-vue-next'
import { Window } from '@wailsio/runtime'

interface Props {
  title: string
}

defineProps<Props>()

const close = (): void => {
  void Window.Close().catch(() => undefined)
}

const onKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape') {
    event.preventDefault()
    close()
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="window-shell">
    <header class="window-titlebar">
      <span class="window-title">{{ title }}</span>
      <button
        type="button"
        class="window-close"
        aria-label="Fechar"
        title="Fechar"
        @click="close"
      >
        <X :size="14" />
      </button>
    </header>
    <main class="window-body">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.window-shell {
  display: flex;
  flex-direction: column;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: var(--app-bg);
  color: var(--app-fg);
  /* O Wails lê esta propriedade para permitir redimensionar pelas bordas. */
  --wails-resize: all;
}

.window-titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  height: 36px;
  flex-shrink: 0;
  padding: 0 6px 0 14px;
  background-color: var(--app-surface);
  border-bottom: 1px solid var(--app-border);
  user-select: none;
  /* Região de arrasto da janela frameless. */
  --wails-draggable: drag;
}

.window-title {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--app-fg-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.window-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.375rem;
  border: none;
  border-radius: 0.375rem;
  background: transparent;
  color: var(--app-fg-muted);
  cursor: pointer;
  /* Clicável: fora da região de arrasto. */
  --wails-draggable: no-drag;
}

.window-close:hover {
  background-color: rgb(220 38 38);
  color: #fff;
}

.window-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
</style>
