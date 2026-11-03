import { test, expect } from '@playwright/test'

test.describe('Tema', () => {
  test('botão alterna claro/escuro e persiste a escolha no config', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    const themeButton = page.locator('button[aria-label="Alternar tema"]')
    await expect(themeButton).toBeVisible()

    const before = await page.evaluate(() => document.documentElement.dataset.theme)
    expect(['light', 'dark']).toContain(before)

    await themeButton.click()
    const after = await page.evaluate(() => document.documentElement.dataset.theme)
    expect(after).not.toBe(before)

    // Persistiu no config.json: recarregar mantém a escolha (e não volta
    // para o tema do sistema).
    await page.reload()
    await page.waitForLoadState('networkidle')
    await expect
      .poll(() => page.evaluate(() => document.documentElement.dataset.theme))
      .toBe(after)

    // Restaura o estado inicial para não deixar o config alterado.
    await page.locator('button[aria-label="Alternar tema"]').click()
    await expect
      .poll(() => page.evaluate(() => document.documentElement.dataset.theme))
      .toBe(before)
  })
})
