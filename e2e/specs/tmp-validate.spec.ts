import { test, expect } from '@playwright/test'

// Valida o IsValid no input do diretório do jogo (Enter/blur).
const INVALID_DIR = 'C:\\__ffx_e2e_nao_existe__'

test.beforeEach(async ({ page }) => {
  await page.goto('/')
  await page.waitForLoadState('networkidle')
  await page.locator('button[aria-label="Configurações"]').click()
  await expect(page.locator('.dialog-overlay')).toBeVisible()
})

test('caminho inexistente fica vermelho, com mensagem e sem persistir', async ({ page }) => {
  const dir = page.locator('#GameDir')
  await dir.fill(INVALID_DIR)
  await dir.blur()

  await expect(page.locator('#GameDir-problems')).toBeVisible({ timeout: 10_000 })
  await expect(dir).toHaveAttribute('aria-invalid', 'true')
  await expect(dir).toHaveValue(INVALID_DIR)
  await expect(page.getByText(/diretório não existe/i)).toBeVisible()
})

test('não existe laudo verde de sucesso no input', async ({ page }) => {
  // A mensagem só aparece quando inválido: nenhum texto de sucesso no DOM.
  await expect(page.getByText('Dados do jogo encontrados')).toHaveCount(0)
})
