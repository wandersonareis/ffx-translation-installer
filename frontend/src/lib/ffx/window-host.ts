/**
 * Host das janelas filhas do Wails.
 *
 * O backend abre a janela com URL "/#config" (hash na raiz) porque o asset
 * server do Wails v3 só serve o index.html em "/" — não existe fallback de
 * SPA para rotas como /w/. O hash escolhe o conteúdo montado pelo main.ts.
 */
export function windowNameFromHash(hash: string = window.location.hash): string | null {
  const raw = hash.replace(/^#\/?/, '').trim()
  return raw === '' ? null : raw
}

export function isConfigWindow(hash?: string): boolean {
  return windowNameFromHash(hash) === 'config'
}
