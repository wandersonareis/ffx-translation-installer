<script setup lang="ts">
import Button from '@/components/ui/Button.vue'

const emit = defineEmits<{
  install: []
  update: []
  uninstall: []
}>()

const actions = [
  { key: 'install', label: 'Instalar tradução' },
  { key: 'update', label: 'Atualizar tradução' },
  { key: 'uninstall', label: 'Desinstalar tradução' },
] as const

type ActionKey = (typeof actions)[number]['key']

const run = (key: ActionKey): void => {
  if (key === 'install') {
    emit('install')
  } else if (key === 'update') {
    emit('update')
  } else {
    emit('uninstall')
  }
}
</script>

<template>
  <footer class="app-actions" aria-label="Ações da tradução">
    <!-- water_slow.webm (AV1) em loop como fundo da barra de botões. -->
    <video
      class="water-video"
      autoplay
      muted
      loop
      playsinline
      preload="auto"
      aria-hidden="true"
    >
      <source src="/water.webm" type="video/webm">
    </video>
    <Button
      v-for="action in actions"
      :key="action.key"
      variant="link"
      size="sm"
      type="button"
      :class="`action-link action-link--${action.key}`"
      @click="run(action.key)"
    >
      {{ action.label }}
    </Button>
  </footer>
</template>

<style scoped>
.app-actions {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2rem;
  flex-shrink: 0;
  height: var(--bottom-bar-height, 61px);
  padding: 0.75rem 1rem;
  overflow: hidden;
  border-top: 1px solid rgb(63 63 70);
}

/* water.webm em loop como fundo da barra de botões. */
.water-video {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;

  transform: translateZ(0);
  backface-visibility: hidden;
  will-change: transform;
}

/* Escurece a água o suficiente para o texto azul continuar legível. */
.app-actions::before {
  content: '';
  position: absolute;
  inset: 0;
  background-color: rgb(21 22 26 / 0.55);
  pointer-events: none;
  z-index: 1;
}

.action-link {
  position: relative;
  z-index: 2;
  text-decoration: none;
  text-underline-offset: 4px;
  text-shadow: 0 1px 2px rgb(0 0 0 / 0.9);
}

.action-link:hover {
  text-decoration: underline;
}
</style>
