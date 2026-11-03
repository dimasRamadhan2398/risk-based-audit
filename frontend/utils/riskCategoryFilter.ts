// Risk category filter for the Corporate Risk Profile register.
// Pure helpers: no store or i18n access, so they can be unit tested directly.

export const ALL_RISK_CATEGORIES = 'all'

// Category names the app already knows, mapped to their riskProfile.categories i18n key.
// Unknown categories from the data are shown as stored.
const CATEGORY_I18N_KEYS: Record<string, string> = {
  'Financial': 'financial',
  'Technology': 'technology',
  'Compliance': 'compliance',
  'Governance': 'governance',
  'Operations': 'operations',
  'Human Resources': 'humanResources',
  'Strategic': 'strategic'
}

export const getRiskCategoryI18nKey = (category: string): string | null =>
  CATEGORY_I18N_KEYS[category] ?? null

// Distinct, trimmed, non-empty categories present in the data, sorted for display.
export const getDistinctRiskCategories = (risks: ReadonlyArray<{ category?: string | null }>): string[] => {
  const set = new Set<string>()
  for (const risk of risks) {
    const category = risk.category?.trim()
    if (category) set.add(category)
  }
  return Array.from(set).sort((a, b) => a.localeCompare(b))
}

// 'all' (or an empty selection) keeps every row; otherwise only exact category matches.
export const filterRisksByCategory = <T extends { category?: string | null }>(
  risks: ReadonlyArray<T>,
  category: string | null | undefined
): T[] => {
  if (!category || category === ALL_RISK_CATEGORIES) return [...risks]
  return risks.filter(risk => risk.category?.trim() === category)
}
