import { describe, it, expect } from 'vitest'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, resolve } from 'node:path'

const ROOT = resolve(__dirname, '../..')
const DIRS = ['components', 'pages', 'stores', 'composables', 'layouts', 'plugins', 'middleware', 'utils']
const EXT = /\.(vue|ts|js)$/

// Bare alert( / confirm( / prompt( or window.alert( etc. Method calls such as
// modal.confirm( or showAlert( are not matched.
const NATIVE_DIALOG = /(?<![\w$.])(?:window\.|globalThis\.)?(?:alert|confirm|prompt)\s*\(/

function walk(dir: string): string[] {
  let out: string[] = []
  for (const name of readdirSync(dir)) {
    const full = join(dir, name)
    if (statSync(full).isDirectory()) out = out.concat(walk(full))
    else if (EXT.test(name)) out.push(full)
  }
  return out
}

describe('no native browser dialogs', () => {
  it('uses toasts / the global confirm modal instead of alert, confirm or prompt', () => {
    const offenders: string[] = []
    for (const d of DIRS) {
      let files: string[] = []
      try {
        files = walk(join(ROOT, d))
      } catch {
        continue
      }
      for (const file of files) {
        readFileSync(file, 'utf8').split('\n').forEach((line, i) => {
          if (/^\s*(\/\/|\*|\/\*)/.test(line)) return
          if (NATIVE_DIALOG.test(line)) offenders.push(`${file.replace(ROOT + '/', '')}:${i + 1}: ${line.trim()}`)
        })
      }
    }
    expect(offenders).toEqual([])
  })
})
