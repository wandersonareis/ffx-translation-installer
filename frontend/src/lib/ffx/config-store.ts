import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ConfigSnapshot } from '@bindings/ffxinstaller/backend/config/models'
import { GetConfig, SetGameExeDir } from '@bindings/ffxinstaller/backend/services/configservice'
import { GetGamePath, IsValid } from '@bindings/ffxinstaller/backend/service/gameservice'
import { useToastStore } from './toast'

/** Snapshot do config (ConfigService.GetConfig e evento ConfigChanged). */
export type ConfigData = ConfigSnapshot

/** Nome do jogo suportado — o appId do Steam é detalhe interno do backend. */
export const GAME_NAME = 'FINAL FANTASY X/X-2 HD Remaster'

/** Estado da checagem do diretório do jogo pelo GameService.IsValid. */
export type GameDirCheck = 'idle' | 'valid' | 'invalid' | 'error'

const INVALID_DIR_MESSAGE =
  'Diretório do jogo inválido: escolha a pasta que contém data\\FFX_Data.vbf e data\\FFX2_Data.vbf.'

/**
 * Store do config: espelha o backend (config.json), roda o fluxo de
 * inicialização e valida o diretório do jogo com GameService.IsValid.
 *
 * O único caminho do jogo é o diretório: o seletor pede FFX.exe/FFX-2.exe só
 * para garantir a pasta certa e o valor salvo é sempre um diretório — o mesmo
 * formato devolvido pelo serviço que procura o jogo no Steam.
 *
 * Cada janela (principal, modal, filha) tem sua instância; o evento
 * ConfigChanged mantém todas em dia sem re-fetch ao reabrir.
 */
export const useConfigStore = defineStore('config', () => {
  const toast = useToastStore()

  const gameDir = ref('')
  const loading = ref(false)

  /** Resultado do IsValid para o diretório do jogo. */
  const dirCheck = ref<GameDirCheck>('idle')

  // Inicialização e validação: guardas de concorrência para nunca duplicar
  // toast da mesma validação nem aplicar resultado atrasado.
  let initPromise: Promise<void> | null = null
  let validationSeq = 0

  const notifyError = (error: unknown): void => {
    const message = error instanceof Error ? error.message : String(error ?? 'Erro')
    toast.error(message, { duration: 5000 })
  }

  /** Aplica um snapshot cru (carga inicial ou eco do backend). */
  const applySnapshot = (data: ConfigData | undefined | null): void => {
    if (!data) return
    gameDir.value = data.GameExeDir || ''
  }

  /**
   * Confere o diretório do jogo com GameService.IsValid
   * (data\FFX_Data.vbf e data\FFX2_Data.vbf).
   *
   * `notify` dispara NO MÁXIMO um toast por chamada — ou seja, um por
   * validação, seja no start, no blur ou após o diálogo. Revalidações
   * silenciosas (eco do backend) passam com notify=false.
   */
  const validateGameDir = async (
    path: string = gameDir.value,
    options: { notify?: boolean } = {},
  ): Promise<GameDirCheck> => {
    const dir = (path ?? '').trim()
    const seq = ++validationSeq

    if (!dir) {
      dirCheck.value = 'idle'
      return 'idle'
    }

    let check: GameDirCheck
    try {
      check = (await IsValid(dir)) ? 'valid' : 'invalid'
    } catch (error) {
      check = 'error'
      if (options.notify && seq === validationSeq) notifyError(error)
    }

    if (seq === validationSeq) dirCheck.value = check

    if (options.notify && seq === validationSeq && check === 'invalid') {
      toast.error(INVALID_DIR_MESSAGE, { duration: 8000 })
    }

    return check
  }

  /**
   * Carga do snapshot persistido. O backend também garante o diretório fixo
   * de recursos (<exec>/resources); se não for possível criá-lo, GetConfig
   * falha e o erro sobe (toast no init / no commit).
   */
  const load = async (): Promise<void> => {
    loading.value = true
    try {
      applySnapshot(await GetConfig())
    } finally {
      loading.value = false
    }
  }

  // Commit + recarga: o backend é a fonte da verdade (valida a existência e
  // devolve o snapshot atualizado).
  const commit = async (call: () => Promise<unknown>): Promise<void> => {
    await call()
    await load()
  }

  /**
   * Aplica o valor do campo do diretório do jogo: persiste e valida com
   * IsValid. Diretório inexistente não é persistido (o rascunho continua na
   * tela, em vermelho, para correção).
   */
  const applyGameDir = async (
    raw: string,
    options: { notify?: boolean } = {},
  ): Promise<GameDirCheck> => {
    const dir = (raw ?? '').trim()

    if (!dir) {
      if (gameDir.value !== '') {
        try {
          await commit(() => SetGameExeDir(''))
        } catch (error) {
          notifyError(error)
          return 'error'
        }
      }
      dirCheck.value = 'idle'
      return 'idle'
    }

    if (dir !== gameDir.value) {
      try {
        await commit(() => SetGameExeDir(dir))
      } catch (error) {
        dirCheck.value = 'invalid'
        if (options.notify) notifyError(error)
        return 'invalid'
      }
    }

    return validateGameDir(dir, options)
  }

  /** Sincroniza com o eco do backend sem duplicar toasts nem chamadas. */
  const syncFromSnapshot = async (data: ConfigData | undefined | null): Promise<void> => {
    const previousDir = gameDir.value
    applySnapshot(data)
    if (gameDir.value !== previousDir) await validateGameDir(gameDir.value)
  }

  /**
   * Fluxo de inicialização:
   *   1. busca o config;
   *   2. diretório presente (veio do config)? valida com IsValid;
   *   3. vazio? busca a pasta no serviço (Steam);
   *   4. ainda vazio? toast avisando que falta o diretório do jogo.
   */
  const initialize = async (): Promise<void> => {
    await load()

    if (gameDir.value) {
      await validateGameDir(gameDir.value, { notify: true })
      return
    }

    let dir = ''
    try {
      dir = await GetGamePath()
    } catch (error) {
      notifyError(error)
      return
    }

    if (!dir) {
      toast.warning(
        `Diretório do jogo não encontrado. Instale ${GAME_NAME} no Steam ou selecione a pasta do jogo nas Configurações.`,
        { duration: 8000 },
      )
      return
    }

    const check = await applyGameDir(dir, { notify: true })
    if (check === 'valid') {
      toast.success(`Diretório do jogo detectado automaticamente: ${dir}`, { duration: 6000 })
    }
  }

  /** Roda o fluxo de inicialização uma única vez por janela. */
  const init = (): Promise<void> => {
    initPromise ??= initialize().catch((error: unknown) => {
      notifyError(error)
    })
    return initPromise
  }

  return {
    gameDir,
    loading,
    dirCheck,
    applySnapshot,
    load,
    init,
    validateGameDir,
    applyGameDir,
    syncFromSnapshot,
  }
})
