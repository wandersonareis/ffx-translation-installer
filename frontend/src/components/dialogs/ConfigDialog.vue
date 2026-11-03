<script setup lang="ts">
import { computed } from 'vue'
import ConfigDialogContent from './ConfigDialogContent.vue'
import Dialog from '@/components/ui/Dialog.vue'
import DialogContent from '@/components/ui/DialogContent.vue'
import DialogHeader from '@/components/ui/DialogHeader.vue'
import DialogTitle from '@/components/ui/DialogTitle.vue'
import DialogDescription from '@/components/ui/DialogDescription.vue'
import { useConfigDialogStore } from '@/lib/ffx/config-dialog-store'

const configDialogStore = useConfigDialogStore()
</script>

<template>
  <Dialog v-model="configDialogStore.open">
    <DialogContent class="sm:max-w-2xl">
      <DialogHeader>
        <DialogTitle>Configurações</DialogTitle>
        <DialogDescription class="sr-only">Diretórios usados pelo instalador</DialogDescription>
      </DialogHeader>
      <ConfigDialogContent v-if="configDialogStore.open" />
    </DialogContent>
  </Dialog>
</template>