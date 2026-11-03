<script setup lang="ts">
import { onMounted } from 'vue'
import { useHotkeys } from '@/lib/ffx/hotkeys'
import { useConfigDialogStore } from '@/lib/ffx/config-dialog-store'
import { useConfigStore } from '@/lib/ffx/config-store'
import AppHeader from '@/components/AppHeader.vue'
import AppActions from '@/components/AppActions.vue'
import ConfigDialog from '@/components/dialogs/ConfigDialog.vue'
import ToastContainer from '@/components/ui/ToastContainer.vue'

const configDialogStore = useConfigDialogStore()
const config = useConfigStore()

const handleOpenConfig = () => {
  configDialogStore.openConfigDialog()
}

// Atalho global Ctrl+, para abrir configurações
useHotkeys([
  {
    hotkey: 'Mod+,',
    callback: () => configDialogStore.openConfigDialog(),
    options: {
      meta: { name: 'Abrir Configurações', group: 'Global' }
    }
  }
])

// Configurações da janela principal - 896x504 (70% de 1280x720) não redimensionável
onMounted(() => {
  // O tamanho da janela é controlado pelo backend (main.go)
  // Fluxo de inicialização: config → serviço (Steam) → toast se faltar o .exe.
  void config.init()
})
</script>

<template>
  <div class="app-container">
    <AppHeader @open-config="handleOpenConfig" />

    <main class="app-main">
      <div class="logo-centered">
        <img
          src="/logo_us.png"
          alt="FINAL FANTASY X and X-2 HD Remaster"
          class="main-logo"
        />
      </div>
    </main>

    <!-- Faixa animada de água como fundo da barra de ações -->
    <AppActions />

    <!-- Modal de Configurações (browser/e2e) -->
    <ConfigDialog v-model="configDialogStore.open" />

    <!-- Toast Container -->
    <ToastContainer />
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  /* header no topo, conteúdo no meio e barra de ações no rodapé */
  justify-content: space-between;
  width: 100%;
  height: 100vh;
  background-color: var(--app-bg);
  color: var(--app-fg);
  overflow: hidden;
  /* Altura fixa compartilhada: faixa de água = barra de botões. */
  --bottom-bar-height: 61px;
}

.app-main {
  flex: 1;
  min-height: 0;
  display: flex;
  /* Centraliza na horizontal; na vertical fica no TOPO, afastado apenas
     pelo padding (sem centralização vertical). */
  align-items: flex-start;
  justify-content: center;
  padding: 2rem 2rem 0;
  overflow: auto;
}

.logo-centered {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.main-logo {
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  object-fit: contain;
}
</style>