<script setup lang="ts">
import WindowChrome from './WindowChrome.vue'
import ConfigDialogContent from '@/components/dialogs/ConfigDialogContent.vue'
import ToastContainer from '@/components/ui/ToastContainer.vue'
</script>

<template>
  <WindowChrome title="Configurações">
    <div class="config-window-body">
      <ConfigDialogContent />
    </div>
  </WindowChrome>

  <ToastContainer />
</template>

<style scoped>
.config-window-body {
  padding: 1rem;
}
</style>
