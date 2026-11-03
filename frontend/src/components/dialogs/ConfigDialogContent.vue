<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { FolderOpen } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { type ConfigData, useConfigStore } from '@/lib/ffx/config-store'
import { useToastStore } from '@/lib/ffx/toast'
import { useWailsEvent } from '@/lib/ffx/use-wails-event'
import { SelectGameDir } from '@bindings/ffxinstaller/backend/services/configservice'
import { ConfigChanged } from '@bindings/events'

const config = useConfigStore()
const toast = useToastStore()

// Rascunho do diretório do jogo enquanto o usuário digita.
const editing = ref<string | null>(null)
const pending = ref(false)

const shownDir = computed<string>(() => editing.value ?? config.gameDir)

// Validação única com GameService.IsValid: só há mensagem (e borda vermelha)
// quando inválido — sem laudo verde de sucesso.
const dirInvalid = computed<boolean>(() => config.dirCheck === 'invalid')

const notifyError = (error: unknown): void => {
  const message = error instanceof Error ? error.message : String(error ?? 'Erro')
  toast.error(message, { duration: 5000 })
}

/**
 * Blur/Enter do diretório do jogo: valida o valor existente do input com
 * IsValid. Se inválido, o texto digitado fica na tela (vermelho) para correção.
 */
const applyDraft = async (draft: string): Promise<void> => {
  if (pending.value) return
  pending.value = true
  try {
    const result = await config.applyGameDir(draft, { notify: true })
    if ((result === 'valid' || result === 'idle') && editing.value === draft) {
      editing.value = null
    }
  } catch (error) {
    notifyError(error)
  } finally {
    pending.value = false
  }
}

const onBlur = (): void => {
  const draft = editing.value
  if (draft === null) return
  void applyDraft(draft)
}

const onEnter = (): void => {
  const draft = editing.value
  if (draft === null) return
  void applyDraft(draft)
}

/**
 * Botão da pasta do jogo: o seletor exige FFX.exe/FFX-2.exe e o backend
 * devolve só o diretório — o executável nunca é persistido.
 */
const selectGameDir = async (): Promise<void> => {
  try {
    const dir = await SelectGameDir()
    if (!dir) return
    editing.value = null
    await applyDraft(dir)
  } catch (error) {
    notifyError(error)
  }
}

// Eco do backend: outra janela (ou um setter desta) mudou o config. Revalida
// em silêncio para não repetir o toast da validação que originou a mudança.
useWailsEvent(ConfigChanged, (data) => {
  void config.syncFromSnapshot(data as ConfigData)
})

// Fluxo de inicialização (config → serviço Steam → toast). Idempotente: a
// primeira janela montada roda; as demais reaproveitam a mesma promise.
onMounted(() => {
  void config.init()
})
</script>

<template>
  <div class="flex flex-col gap-3 pt-2 min-w-0">
    <div class="flex items-center gap-2">
      <div class="grid flex-1 gap-1.5">
        <Label htmlFor="GameDir">Diretório do Jogo</Label>
        <Input
          id="GameDir"
          :value="shownDir"
          :invalid="dirInvalid"
          :aria-describedby="`GameDir-hint${dirInvalid ? ' GameDir-problems' : ''}`"
          placeholder="Pasta do jogo (onde está FFX.exe)"
          @input="text => editing = text"
          @blur="onBlur"
          @keydown="e => { if (e.key === 'Enter') { e.preventDefault(); onEnter() } }"
        />
        <p id="GameDir-hint" class="text-xs text-(--app-fg-muted)">
          O seletor pede FFX.exe ou FFX-2.exe e salva apenas a pasta do jogo.
        </p>
        <p v-if="dirInvalid" id="GameDir-problems" class="text-xs text-(--app-danger)">
          Diretório do jogo inválido: escolha a pasta que contém data\FFX_Data.vbf e
          data\FFX2_Data.vbf.
        </p>
      </div>
      <Button
        variant="ghost"
        size="icon"
        @mousedown="e => e.preventDefault()"
        @click="selectGameDir"
        aria-label="Selecionar a pasta do jogo"
        title="Procurar a pasta do jogo (escolha FFX.exe)"
      >
        <FolderOpen :size="20" />
      </Button>
    </div>
  </div>
</template>
