#!/usr/bin/env node
// Type-check do frontend com o tsgoblin: paridade com o vue-tsc usando o
// compilador nativo do TypeScript 7 (tsgo). Chamado por `npm run typecheck`,
// que por sua vez roda antes de cada build (inclusive o `wails3 dev`).
//
// Executamos os módulos do tsgoblin diretamente em vez do bin `tsgoblin`
// (0.1.0) por três problemas específicos de Windows:
//
//  1. `bin/cli.mjs` monta o caminho com `new URL(import.meta.url).pathname`,
//     virando "\F:\...\...", então o script nem carregava.
//  2. `check` spawna `node_modules/.bin/tsgo` sem shell; no Windows esse arquivo
//     é um shim `sh`, não um executável, então o tsgo nunca rodava e o check
//     "passava" sem verificar nada (0 erros sempre).
//  3. O manifesto `.tsgoblin-maps.json` é gravado com chaves "F:/..." e o check
//     resolve "F:\...", então nenhuma entrada casava e o filtro de paridade
//     (que descarta os falsos positivos do template gerado) era ignorado.
//
// Aqui resolvemos os três sem alterar node_modules. Fora do Windows o script é
// um wrapper fino.
import childProcess from 'node:child_process'
import fs from 'node:fs'
import { createRequire, syncBuiltinESMExports } from 'node:module'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const require = createRequire(import.meta.url)
const originalSpawnSync = childProcess.spawnSync
// Diretório do frontend, onde ficam tsconfig.json e os artefatos do tsgoblin.
const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const tsgoblinDir = path.dirname(require.resolve('tsgoblin/package.json'))
const tsgoblinScript = (name) => pathToFileURL(path.join(tsgoblinDir, 'src', `${name}.mjs`)).href

// (2) Troca o spawn do shim `.bin/tsgo` pelo binário nativo do tsgo.
function resolveNativeTsgo() {
  if (process.platform !== 'win32') return null
  try {
    const platformPkg = `@typescript/native-preview-${process.platform}-${process.arch}`
    return path.join(
      path.dirname(require.resolve(`${platformPkg}/package.json`)),
      'lib',
      'tsgo.exe',
    )
  } catch {
    return null
  }
}

const nativeTsgo = resolveNativeTsgo()
if (nativeTsgo) {
  childProcess.spawnSync = (command, ...rest) =>
    path.basename(String(command)) === 'tsgo'
      ? originalSpawnSync(nativeTsgo, ...rest)
      : originalSpawnSync(command, ...rest)
  // Faz os `import ... from 'node:child_process'` de outros módulos (o próprio
  // check.mjs) enxergarem a versão interceptada.
  syncBuiltinESMExports()
}

// Os módulos do tsgoblin leem os argumentos de process.argv na importação.
async function run(script, args) {
  const originalArgv = process.argv
  process.argv = [originalArgv[0], originalArgv[1], ...args]
  try {
    await import(tsgoblinScript(script))
  } finally {
    process.argv = originalArgv
  }
}

// (3) Normaliza as chaves do manifesto para o mesmo formato que o check resolve.
function normalizeMapsKeys() {
  const mapsPath = path.join(rootDir, '.tsgoblin-maps.json')
  if (!fs.existsSync(mapsPath)) return
  const maps = JSON.parse(fs.readFileSync(mapsPath, 'utf8'))
  const files = {}
  for (const [key, entry] of Object.entries(maps.files)) {
    files[path.resolve(rootDir, key)] = { ...entry, vue: path.resolve(rootDir, entry.vue) }
  }
  fs.writeFileSync(mapsPath, JSON.stringify({ ...maps, files }))
}

await run('generate', [path.join(rootDir, 'tsconfig.json')])
normalizeMapsKeys()
await run('check', [path.join(rootDir, 'tsconfig.tsgo.json'), `--repo-root=${rootDir}`])
