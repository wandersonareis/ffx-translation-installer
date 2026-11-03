import { onMounted, onUnmounted } from 'vue'

interface HotkeyOptions {
  meta?: {
    name: string
    group: string
  }
}

interface HotkeyEntry {
  hotkey: string | string[]
  callback: () => void
  options?: HotkeyOptions
}

const normalizeHotkey = (hotkey: string): string[] => {
  const isMac = /mac/i.test(navigator.platform) || /mac os/i.test(navigator.userAgent)
  return hotkey.toLowerCase().split('+').map(k => {
    const key = k.trim()
    // "Mod" é Cmd no macOS e Ctrl nos demais — o mesmo atalho nas duas pontas.
    if (key === 'mod') return isMac ? 'meta' : 'ctrl'
    return keyMap[key] || key
  })
}

const keysMatch = (pressed: Set<string>, required: string[]): boolean => {
  return required.every(k => pressed.has(k))
}

const keyMap: Record<string, string> = {
  'control': 'ctrl',
  'ctrl': 'ctrl',
  'meta': 'meta',
  'mod': 'meta',
  'shift': 'shift',
  'alt': 'alt',
  'option': 'alt',
  'command': 'meta',
  'cmd': 'meta',
  'escape': 'esc',
  'esc': 'esc',
  'space': ' ',
  'plus': '+',
  'minus': '-',
  'comma': ',',
  'period': '.',
}

export function useHotkeys(hotkeys: HotkeyEntry[]) {
  const pressed = new Set<string>()

  const handleKeydown = (e: KeyboardEvent) => {
    // Ignorar se estiver digitando em input/textarea
    const target = e.target as HTMLElement
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
      return
    }

    const key = keyMap[e.key.toLowerCase()] || e.key.toLowerCase()
    pressed.add(key)

    // Verificar hotkeys
    for (const hk of hotkeys) {
      const hotkeyArray = Array.isArray(hk.hotkey) ? hk.hotkey : [hk.hotkey]
      for (const hkStr of hotkeyArray) {
        const required = normalizeHotkey(hkStr)
        if (keysMatch(pressed, required)) {
          e.preventDefault()
          hk.callback()
          break
        }
      }
    }
  }

  const handleKeyup = (e: KeyboardEvent) => {
    const key = keyMap[e.key.toLowerCase()] || e.key.toLowerCase()
    pressed.delete(key)
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
    window.addEventListener('keyup', handleKeyup)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
    window.removeEventListener('keyup', handleKeyup)
  })
}