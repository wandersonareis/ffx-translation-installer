/**
 * Detecta execução dentro do WebView do Wails (desktop) vs browser puro
 * (web / server mode).
 *
 * No Wails v3 o @wailsio/runtime escolhe o transporte em tempo de carga:
 * WebView2 (window.chrome.webview), WKWebView (window.webkit.messageHandlers.external)
 * ou fetch HTTP (server mode).
 */
export function isDesktop(): boolean {
  if (typeof window === 'undefined') return false
  const w = window as Window & {
    chrome?: { webview?: { postMessage?: (data: unknown) => void } }
    webkit?: {
      messageHandlers?: {
        external?: { postMessage?: (data: unknown) => void }
      }
    }
    wails?: { invoke?: unknown }
  }
  // Server mode: wails.invoke exists but WebView APIs don't
  // Real desktop: WebView APIs exist (chrome.webview or webkit.messageHandlers)
  const hasWebView = Boolean(
    w.chrome?.webview?.postMessage ||
    w.webkit?.messageHandlers?.external?.postMessage,
  )
  return hasWebView
}