import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const DETAIL_SFC = resolve(__dirname, '../../components/annual-audit/AnnualAuditDetail.vue')

const NEUTRAL_COLORS = 'gray|slate|zinc|neutral|stone'
// Accent colours only need a dark counterpart at the very light (backgrounds/borders)
// or very dark (text) ends; mid shades such as icon `text-primary-500` read fine in both modes.
const ACCENT_COLORS = 'primary|secondary|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|pink|rose'

const LIGHT_ONLY_COLOR = new RegExp(
  `^((?:[a-z-]+:)*)(text|bg|border)-(?:`
  + `(?:${NEUTRAL_COLORS})-\\d{2,3}`
  + `|black|white`
  + `|(?:${ACCENT_COLORS})-(?:50|100|200|600|700|800|900|950)`
  + `)(?:\\/\\d+)?$`
)
const BARE_BORDER = /^border(?:-[trblxyse])?$/

function extractTemplate(src: string): string {
  const start = src.indexOf('<template>')
  const end = src.lastIndexOf('</template>')
  return src.slice(start, end)
}

/** Static class attributes plus every quoted string inside a `:ui="{ ... }"` binding. */
function collectClassLists(template: string): string[] {
  const lists: string[] = []
  for (const m of template.matchAll(/(?<![:\w-])class="([^"]*)"/g)) lists.push(m[1]!)
  for (const ui of template.matchAll(/:ui="(\{[\s\S]*?\})"/g)) {
    for (const s of ui[1]!.matchAll(/'([^']*)'/g)) lists.push(s[1]!)
  }
  return lists
}

/** Returns the light-mode colour tokens in a class list that have no matching `dark:` override. */
function missingDarkCounterparts(classList: string): string[] {
  const tokens = classList.split(/\s+/).filter(Boolean)
  const darkTokens = tokens.filter(t => t.startsWith('dark:'))
  const missing: string[] = []

  for (const token of tokens) {
    if (token.startsWith('dark:')) continue

    const color = token.match(LIGHT_ONLY_COLOR)
    if (color) {
      const [, variants = '', prop] = color
      if (!darkTokens.some(d => d.startsWith(`dark:${variants}${prop}-`))) missing.push(token)
      continue
    }

    // A bare `border` uses currentColor, which turns near-white on a dark surface.
    if (BARE_BORDER.test(token) && !darkTokens.some(d => d.startsWith('dark:border-'))) {
      missing.push(token)
    }
  }
  return missing
}

describe('AnnualAuditDetail.vue dark mode styling', () => {
  const src = readFileSync(DETAIL_SFC, 'utf-8')
  const template = extractTemplate(src)
  const classLists = collectClassLists(template)

  it('checker flags the original light-only label and accepts the fixed one', () => {
    expect(missingDarkCounterparts('font-bold text-gray-600  w-44 text-sm')).toEqual(['text-gray-600'])
    expect(missingDarkCounterparts('font-bold text-gray-600 dark:text-gray-400 w-44 text-sm')).toEqual([])
    expect(missingDarkCounterparts('p-4 border rounded-lg bg-gray-50')).toEqual(['border', 'bg-gray-50'])
    expect(missingDarkCounterparts('text-gray-400 hover:text-gray-600 dark:text-gray-500')).toEqual(['hover:text-gray-600'])
    expect(missingDarkCounterparts('w-4 h-4 text-primary-500 text-sm text-center')).toEqual([])
  })

  it('finds the template class lists (guards against a vacuous pass)', () => {
    expect(classLists.length).toBeGreaterThan(40)
  })

  it('gives every light text/background/border colour a dark: counterpart', () => {
    const offenders = classLists
      .map(list => ({ list, missing: missingDarkCounterparts(list) }))
      .filter(entry => entry.missing.length > 0)
    expect(offenders).toEqual([])
  })

  it('has no hardcoded hex colours or inline styles in the template', () => {
    for (const list of classLists) {
      expect(list, `hex colour in "${list}"`).not.toMatch(/#[0-9a-fA-F]{3,8}\b/)
    }
    expect(template).not.toMatch(/\s:?style="/)
    expect(src).not.toMatch(/<style[\s\S]*#[0-9a-fA-F]{3,8}\b[\s\S]*<\/style>/)
  })

  it('keeps the risk level badge on the shared helper', () => {
    expect(template).toContain(':class="getRiskLevelColorClass(activity.riskLevel)"')
    expect(src).toContain('import { getRiskLevelColorClass } from \'~/utils/riskLevelBadge\'')
  })
})
