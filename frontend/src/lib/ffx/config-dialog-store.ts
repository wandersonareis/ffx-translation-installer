import { defineStore } from 'pinia'
import { OpenWindow } from '@bindings/ffxinstaller/backend/services/configservice'
import { isDesktop } from './env'

/**
 * Diálogo de Configurações.
 *
 * ABERTURA UNIFICADA: no desktop vira JANELA FILHA frameless (OpenWindow),
 * arrastável e com foco próprio; no browser/server mode cai no modal, que é
 * o caminho usado pelo teste e2e.
 */
export const useConfigDialogStore = defineStore('configDialog', {
  state: () => ({
    open: false,
  }),
  actions: {
    setOpen(open: boolean) {
      this.open = open
    },
    openConfigDialog() {
      if (isDesktop()) {
        void OpenWindow('config').catch((error: unknown) => {
          const message = error instanceof Error ? error.message : String(error ?? 'Erro')
          console.error(message)
        })
        return
      }
      this.open = true
    },
  },
})
