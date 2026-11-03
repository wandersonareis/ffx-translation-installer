import { test, expect } from '@playwright/test'

test.describe('App shell', () => {
  test('barra de config no topo e ações de tradução no rodapé', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    const viewport = page.viewportSize()
    expect(viewport).not.toBeNull()
    if (!viewport) return

    // Barra superior (config) no topo.
    const header = page.locator('header.app-header')
    await expect(header).toBeVisible()
    const headerBox = await header.boundingBox()
    expect(headerBox).not.toBeNull()
    if (headerBox) {
      expect(headerBox.y).toBeLessThan(100)
      expect(headerBox.y + headerBox.height).toBeLessThan(viewport.height / 2)
    }
    await expect(page.locator('button[aria-label="Configurações"]')).toBeVisible()

    // As três ações em estilo link, no rodapé.
    const install = page.getByRole('button', { name: 'Instalar tradução', exact: true })
    const update = page.getByRole('button', { name: 'Atualizar tradução', exact: true })
    const uninstall = page.getByRole('button', { name: 'Desinstalar tradução', exact: true })
    await expect(install).toBeVisible()
    await expect(update).toBeVisible()
    await expect(uninstall).toBeVisible()

    // Mesmo azul do título do FFX_Resources (sky-400 no tema escuro).
    await expect(install).toHaveCSS('color', 'rgb(56, 189, 248)')

    const footer = page.locator('footer.app-actions')
    await expect(footer).toBeVisible()
    const footerBox = await footer.boundingBox()
    expect(footerBox).not.toBeNull()
    if (footerBox) {
      expect(footerBox.y).toBeGreaterThan(viewport.height / 2)
      expect(footerBox.y + footerBox.height).toBeGreaterThan(viewport.height * 0.9)
      // Altura fixa da barra de ações.
      expect(Math.round(footerBox.height)).toBe(61)
    }

    // water.webm em loop como fundo da barra de botões.
    const waterVideo = footer.locator('video.water-video')
    await expect(waterVideo).toBeAttached()
    await expect(waterVideo).toHaveAttribute('src', '/water.webm')
    await expect(waterVideo).toHaveAttribute('loop', '')
    await expect(waterVideo).toHaveCSS('object-fit', 'cover')

    // Conteúdo (logo) fica entre a barra e o rodapé.
    const logo = page.locator('img.main-logo')
    await expect(logo).toBeVisible()
    const logoBox = await logo.boundingBox()
    expect(logoBox).not.toBeNull()
    if (logoBox) {
      expect(logoBox.y).toBeGreaterThan(headerBox?.y ?? 0)
      expect(logoBox.y).toBeLessThan(footerBox?.y ?? viewport.height)
    }
  })
})
