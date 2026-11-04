import { RISK_FACTOR_CATALOG } from '~/utils/riskFactorCatalog'
import { useI18n } from '~/composables/useI18n'

const seedKey = (id?: string) => {
  const m = id?.match(/^10000000-0000-0000-0000-0*(\d+)$/)
  return m ? m[1] : undefined
}

// Seeded standard risk factors are stored in English; show the Indonesian text when the UI is
// Indonesian. Custom or edited rows (no catalog match, or text differs from the seed) stay as-is.
export const useRiskFactorText = () => {
  const { locale } = useI18n()

  const entry = (id?: string) => {
    const key = seedKey(id)
    return key && locale.value === 'id' ? RISK_FACTOR_CATALOG[key] : undefined
  }

  const factorName = (f?: { id?: string; name?: string } | null) => {
    const e = entry(f?.id)
    return e && f?.name === e.en.name ? e.id.name : (f?.name ?? '')
  }

  const factorDescription = (f?: { id?: string; description?: string } | null) => {
    const e = entry(f?.id)
    return e && f?.description === e.en.description ? e.id.description : (f?.description ?? '')
  }

  const guidelineDesc = (f: { id?: string } | null | undefined, score: number, desc: string) => {
    const e = entry(f?.id)
    const i = 5 - score
    return e && e.en.g[i] === desc ? e.id.g[i] : desc
  }

  return { factorName, factorDescription, guidelineDesc }
}
