<script setup lang="ts">
import { Settings, Sun, Moon } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Tooltip from '@/components/ui/Tooltip.vue'
import { useThemeStore } from '@/lib/ffx/theme'

const emit = defineEmits<{
  'open-config': []
}>()

const theme = useThemeStore()

const handleOpenConfig = () => {
  emit('open-config')
}

const toggleTheme = () => {
  void theme.toggle()
}
</script>

<template>
  <header class="app-header">
    <Tooltip>
      <template #default>
        <Button
          variant="ghost"
          size="icon"
          @click="toggleTheme"
          aria-label="Alternar tema"
          title="Alternar tema"
          class="theme-button"
        >
          <Sun v-if="theme.resolved === 'dark'" :size="22" class="header-icon" />
          <Moon v-else :size="22" class="header-icon" />
        </Button>
      </template>
      <template #content>
        {{ theme.resolved === 'dark' ? 'Tema claro' : 'Tema escuro' }}
      </template>
    </Tooltip>

    <Tooltip>
      <template #default>
        <Button
          variant="ghost"
          size="icon"
          @click="handleOpenConfig"
          aria-label="Configurações"
          class="config-button"
        >
          <Settings :size="22" class="header-icon" />
        </Button>
      </template>
      <template #content>
        Configurações (Ctrl+,)
      </template>
    </Tooltip>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background-color: var(--app-surface);
  border-bottom: 1px solid var(--app-border);
  height: 44px;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 10;
  transition: background-color 150ms ease, border-color 150ms ease;
}

.config-button,
.theme-button {
  flex-shrink: 0;
}

.header-icon {
  color: var(--app-fg);
}
</style>
