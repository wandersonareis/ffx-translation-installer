import { onMounted, onUnmounted } from 'vue'
import { Events } from '@wailsio/runtime'

/**
 * Assina um evento do Wails v3 durante o ciclo de vida do componente.
 *
 * No v3 o callback do runtime recebe um WailsEvent ({ name, data, sender }),
 * não o payload cru — aqui a store/componente continua recebendo só o data.
 */
export function useWailsEvent(eventName: string, handler: (data: unknown) => void): void {
  let off: (() => void) | null = null

  onMounted(() => {
    off = Events.On(eventName, (event: { data: unknown }) => handler(event.data))
  })

  onUnmounted(() => {
    off?.()
    off = null
  })
}
