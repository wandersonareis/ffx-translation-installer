import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './styles/globals.css'
import App from './App.vue'
import ConfigWindow from './components/window/ConfigWindow.vue'
import { isConfigWindow } from './lib/ffx/window-host'
import { useThemeStore } from './lib/ffx/theme'

// A janela principal é servida em "/" (sem hash). O backend abre janelas
// filhas em "/#config": o hash monta outro conteúdo no mesmo index.html.
const RootComponent = isConfigWindow() ? ConfigWindow : App

const app = createApp(RootComponent)
const pinia = createPinia()
app.use(pinia)

// Tema em TODAS as janelas: sistema por padrão, config quando existir.
void useThemeStore(pinia).init()

app.mount('#app')
