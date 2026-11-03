import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { Events } from '@wailsio/runtime'
import type { ConfigSnapshot } from '@bindings/ffxinstaller/backend/config/models'
import { GetConfig, SetTheme } from '@bindings/ffxinstaller/backend/services/configservice'
import { ConfigChanged } from '@bindings/events'

export type ThemePreference = 'system' | 'light' | 'dark'
export type ResolvedTheme = 'light' | 'dark'

const normalizePreference = (value: string | null | undefined): ThemePreference => {
  const v = (value ?? '').trim().toLowerCase()
  return v === 'light' || v === 'dark' ? v : 'system'
}

/**
 * Tema do app.
 *
 * Regra: sem escolha salva no config.json → segue o tema do sistema
 * (prefers-color-scheme). Com "light"/"dark" salvo → usa o do config.
 * O botão do header alterna claro/escuro e persiste a escolha no backend;
 * o evento ConfigChanged mantém todas as janelas em sincronia.
 */
export const useThemeStore = defineStore('theme', () => {
  const preference = ref<ThemePreference>('system')
  const systemTheme = ref<ResolvedTheme>('dark')
  let media: MediaQueryList | null = null
  let initialized = false

  const resolved = computed<ResolvedTheme>(() =>
    preference.value === 'system' ? systemTheme.value : preference.value,
  )

  const apply = (): void => {
    document.documentElement.dataset.theme = resolved.value
    document.documentElement.style.colorScheme = resolved.value
  }

  /** Aplica um snapshot do backend (carga inicial ou eco do ConfigChanged). */
  const applySnapshot = (data: Pick<ConfigSnapshot, 'Theme'> | null | undefined): void => {
    preference.value = normalizePreference(data?.Theme)
    apply()
  }

  const commit = async (next: ThemePreference): Promise<void> => {
    const previous = preference.value
    preference.value = next
    apply()
    try {
      await SetTheme(next)
    } catch (error) {
      console.error('Erro ao salvar o tema:', error)
      preference.value = previous
      apply()
    }
  }

  /** Botão: alterna entre claro e escuro e salva a escolha. */
  const toggle = (): Promise<void> => commit(resolved.value === 'dark' ? 'light' : 'dark')

  /** Define explicitamente (inclui "system"), salvando no config. */
  const setPreference = (value: ThemePreference): Promise<void> => commit(value)

  /** Inicializa: sistema por padrão, depois o que estiver salvo no config. */
  const init = async (): Promise<void> => {
    if (initialized) return
    initialized = true

    media = window.matchMedia?.('(prefers-color-scheme: dark)') ?? null
    systemTheme.value = media?.matches ? 'dark' : 'light'
    media?.addEventListener?.('change', (event) => {
      systemTheme.value = event.matches ? 'dark' : 'light'
      apply()
    })
    apply() // tema do sistema até o config responder

    try {
      applySnapshot(await GetConfig())
    } catch (error) {
      console.error('Erro ao carregar o tema do config:', error)
    }

    Events.On(ConfigChanged, (event: { data: unknown }) => {
      applySnapshot(event.data as ConfigSnapshot)
    })
  }

  return {
    preference,
    resolved,
    systemTheme,
    init,
    toggle,
    setPreference,
    applySnapshot,
  }
})
