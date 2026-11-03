/// <reference types="vite/client" />

interface Window {
  wails?: {
    invoke: (method: string, args?: unknown[]) => Promise<unknown>
    Events: {
      On: (eventName: string, handler: (data: unknown) => void) => void
      Off: (eventName: string, handler: (data: unknown) => void) => void
      Emit: (eventName: string, data: unknown) => void
    }
  }
}