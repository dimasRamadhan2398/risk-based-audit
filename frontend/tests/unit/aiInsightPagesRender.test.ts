import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const readFrontendFile = (relPath: string) =>
  readFileSync(resolve(process.cwd(), relPath), 'utf-8')

describe('AI Insight pages template/script binding safety', () => {
  it('destructures riskCategoryColor and uses toast.add in risk-scoring.vue', () => {
    const source = readFrontendFile('pages/analytics/ai/risk-scoring.vue')
    const scriptMatch = source.match(/<script setup[^>]*>([\s\S]*?)<\/script>/)
    expect(scriptMatch).toBeTruthy()
    const script = scriptMatch![1] || ''

    // Template calls riskCategoryColor(row.riskCategory) inside mobile cards
    expect(source).toContain('riskCategoryColor(row.riskCategory)')
    expect(script).toMatch(/\briskCategoryColor\b/)

    // @nuxt/ui useToast() exposes .add(), not .success()/.info()/.error()
    expect(script).not.toMatch(/\btoast\.(success|info|error)\(/)
    expect(script).toContain('toast.add(')
  })

  it('declares isSavingTaxonomy ref and uses valid Nuxt UI semantic colors in nlp.vue', () => {
    const source = readFrontendFile('pages/analytics/ai/nlp.vue')
    const scriptMatch = source.match(/<script setup[^>]*>([\s\S]*?)<\/script>/)
    expect(scriptMatch).toBeTruthy()
    const script = scriptMatch![1] || ''

    expect(script).toMatch(/const\s+isSavingTaxonomy\s*=\s*ref\(false\)/)
    expect(source).not.toContain('color="rose"')
  })

  it('does not declare duplicate unit field in AuditeeKpiItem in kpi-forecast.vue', () => {
    const source = readFrontendFile('pages/analytics/ai/kpi-forecast.vue')
    const ifaceMatch = source.match(/interface\s+AuditeeKpiItem\s*\{([\s\S]*?)\}/)
    expect(ifaceMatch).toBeTruthy()
    const unitMatches = (ifaceMatch![1] || '').match(/\bunit\?:/g) || []
    expect(unitMatches).toHaveLength(1)
  })

  it('routes sync-profile through /risk-score/sync-profile with python-ai fallback', () => {
    const source = readFrontendFile('composables/useAiAnalytics.ts')
    expect(source).toContain("endpoint === '/risk-score/sync-profile'")
    expect(source).toContain("safeApiFetch('/risk-score/sync-profile'")
    expect(source).not.toContain("safeApiFetch('/predict/risk-score/sync-profile'")
  })
})
