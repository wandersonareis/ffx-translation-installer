import { test, expect } from '@playwright/test'

test.describe('Config Modal', () => {
  test('should open config modal when clicking config button', async ({ page }) => {
    // Go to the app
    await page.goto('/')

    // Wait for the app to load
    await page.waitForLoadState('networkidle')

    // The config button should be in the top right corner (gear icon)
    // It has aria-label="Configurações"
    const configButton = page.locator('button[aria-label="Configurações"]')

    // Verify the button exists
    await expect(configButton).toBeVisible()

    // Click the config button
    await configButton.click()

    // The config modal should open
    // Wait for the dialog overlay to become visible
    const dialogOverlay = page.locator('.dialog-overlay')
    await expect(dialogOverlay).toBeVisible()

    // The modal must actually be a real overlay: fixed, covering the viewport
    // and centered (regressão do modal que renderizava em fluxo estático,
    // sem CSS utilitário, e ficava fora da viewport).
    await expect(dialogOverlay).toHaveCSS('position', 'fixed')
    const viewport = page.viewportSize()
    if (viewport) {
      const box = await dialogOverlay.boundingBox()
      expect(box).not.toBeNull()
      if (box) {
        expect(box.x).toBe(0)
        expect(box.y).toBe(0)
        expect(Math.round(box.width)).toBe(viewport.width)
        expect(Math.round(box.height)).toBe(viewport.height)
      }
    }

    // The dark backdrop must be visible
    const backdrop = page.locator('.dialog-backdrop')
    await expect(backdrop).toBeVisible()
    await expect(backdrop).toHaveCSS('position', 'fixed')

    // The modal title
    const modalTitle = page.locator('h2.dialog-title:has-text("Configurações")')
    await expect(modalTitle).toBeVisible()

    // O único caminho editável é o diretório do jogo.
    const gameDirLabel = page.locator('label:has-text("Diretório do Jogo")')
    await expect(gameDirLabel).toBeVisible()
    await expect(page.locator('#GameDir')).toBeVisible()

    // O diretório de recursos é fixo (<exec>/resources) e não aparece na UI.
    await expect(page.locator('text=Diretório de Recursos')).toHaveCount(0)
    await expect(page.locator('#ResourcesDir')).toHaveCount(0)
    await expect(page.locator('text=Executável Específico')).toHaveCount(0)
  })

  test('config button should be in top right corner', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    const configButton = page.locator('button[aria-label="Configurações"]')

    // Button should be visible
    await expect(configButton).toBeVisible()

    // Button should be in the header (top area)
    const buttonBox = await configButton.boundingBox()
    expect(buttonBox).not.toBeNull()
    if (buttonBox) {
      // Should be in top portion of the screen (header is at top)
      expect(buttonBox.y).toBeLessThan(100)

      // Should be on the right side (flex-end alignment)
      const viewport = page.viewportSize()
      if (viewport) {
        expect(buttonBox.x).toBeGreaterThan(viewport.width / 2)
      }
    }
  })

  test('config modal should close with Escape and reopen', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    const configButton = page.locator('button[aria-label="Configurações"]')
    await configButton.click()

    const dialogOverlay = page.locator('.dialog-overlay')
    await expect(dialogOverlay).toBeVisible()

    await page.keyboard.press('Escape')
    await expect(dialogOverlay).toBeHidden()

    await configButton.click()
    await expect(dialogOverlay).toBeVisible()
    await expect(page.locator('h2.dialog-title')).toBeVisible()
  })

  test('modal opens with Ctrl+, and renders the game directory', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    await page.keyboard.press('Control+,')
    const dialogOverlay = page.locator('.dialog-overlay')
    await expect(dialogOverlay).toBeVisible()

    // Snapshot do backend (config.json) já aplicado no input do jogo.
    await expect(page.locator('#GameDir')).toBeVisible()
    await expect(page.locator('#GameDir-hint')).toContainText('salva apenas a pasta do jogo')
  })

  test('invalid game directory turns the input red with a message', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    await page.locator('button[aria-label="Configurações"]').click()
    await expect(page.locator('.dialog-overlay')).toBeVisible()

    const dir = page.locator('#GameDir')

    // Pasta inexistente: o backend recusa persistir e o input fica vermelho.
    const missing = 'C:\\__ffx_e2e_nao_existe__'
    await dir.fill(missing)
    await dir.press('Enter')

    await expect(page.getByText(/diretório não existe/i)).toBeVisible()
    await expect(dir).toHaveAttribute('aria-invalid', 'true')
    await expect(page.locator('#GameDir-problems')).toBeVisible()

    // O caminho recusado fica na tela (para correção) e o config não muda.
    await expect(dir).toHaveValue(missing)
  })

  test('directory picker is a no-op in server mode and does not break the modal', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    await page.locator('button[aria-label="Configurações"]').click()
    const dialogOverlay = page.locator('.dialog-overlay')
    await expect(dialogOverlay).toBeVisible()

    // No server mode os pickers nativos devolvem "" (cancelamento): nenhum
    // diálogo do SO, nenhum crash — o modal segue utilizável.
    await page.getByRole('button', { name: 'Selecionar a pasta do jogo' }).click()

    await expect(dialogOverlay).toBeVisible()
    await expect(page.locator('h2.dialog-title')).toBeVisible()
  })
})
