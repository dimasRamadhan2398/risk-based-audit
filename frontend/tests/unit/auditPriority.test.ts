import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Audit Priority Feature Restructuring', () => {
  const rootDir = path.resolve(__dirname, '../../')

  it('verifies Audit Priority tab is removed from risk-factors/index.vue', () => {
    const filePath = path.join(rootDir, 'pages/risk-profile/risk-factors/index.vue')
    const content = fs.readFileSync(filePath, 'utf8')

    expect(content).not.toContain("slot: 'priority'")
    expect(content).not.toContain('<template #priority>')
    expect(content).toContain("slot: 'weighting'")
    expect(content).toContain("slot: 'scoring'")
  })

  it('verifies Audit Priority tab is removed from audit-universe/index.vue', () => {
    const filePath = path.join(rootDir, 'pages/risk-profile/audit-universe/index.vue')
    const content = fs.readFileSync(filePath, 'utf8')

    expect(content).not.toContain("value: 'priority'")
    expect(content).not.toContain("slot: 'priority'")
    expect(content).not.toContain('<template #priority>')
    expect(content).toContain("value: 'library'")
    expect(content).toContain("value: 'establish'")
  })

  it('verifies default.vue submenu navigation points to /risk-profile/audit-priority', () => {
    const filePath = path.join(rootDir, 'layouts/default.vue')
    const content = fs.readFileSync(filePath, 'utf8')

    expect(content).toContain("label: t('navigation.auditPriority')")
    expect(content).toContain("to: '/risk-profile/audit-priority'")
    expect(content).not.toContain("to: '/risk-profile/audit-universe?tab=priority'")
  })

  it('verifies dedicated Audit Priority page exists with core features', () => {
    const filePath = path.join(rootDir, 'pages/risk-profile/audit-priority/index.vue')
    expect(fs.existsSync(filePath)).toBe(true)

    const content = fs.readFileSync(filePath, 'utf8')

    // Year selection & Refresh
    expect(content).toContain('selectedYear')
    expect(content).toContain('loadData')

    // Table & Columns: priority and sorting go through the shared scoring helpers (unscored rows last, never prioritized)
    expect(content).toContain('filteredYearlyUniverse')
    expect(content).toContain('isUniverseYearPrioritized(ent)')
    expect(content).toContain('sortUniverseYearsByRisk(')
    expect(content).toContain('ent.risk_level')
    expect(content).toContain('riskFactors.priority.notScoredBadge')

    // Risk Index Level Reference comes from the one shared threshold table, not a local copy
    expect(content).toContain('levelInfoTitle')
    expect(content).toContain('RISK_LEVEL_THRESHOLDS')
    expect(content).toContain('formatRiskLevelRange(row)')
    expect(content).not.toMatch(/>=\s*80/)

    // Store integration
    expect(content).toContain('useAuditUniverseStore')
  })

  it('keeps the reference ranges of the shared thresholds', async () => {
    const { RISK_LEVEL_THRESHOLDS, formatRiskLevelRange } = await import('~/utils/riskLevelBadge')
    expect(RISK_LEVEL_THRESHOLDS.map(formatRiskLevelRange)).toEqual(['80 - 100%', '60 - 79%', '40 - 59%', '20 - 39%', '0 - 19%'])
    expect(RISK_LEVEL_THRESHOLDS.filter(th => th.prioritized).map(th => th.level)).toEqual(['High', 'Medium to High', 'Medium'])
  })

  describe('pagination', () => {
    const rows = Array.from({ length: 32 }, (_, i) => ({ id: i + 1 }))

    it('slices the filtered list into pages of 10', async () => {
      const { ref } = await import('vue')
      const { usePagedList } = await import('~/composables/usePagedList')
      const list = ref(rows)
      const { page, pageData, from, to, pageSize } = usePagedList(list)

      expect(pageSize).toBe(10)
      expect(pageData.value.items).toHaveLength(10)
      expect(pageData.value.items.map(r => r.id)).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
      expect(pageData.value.totalPages).toBe(4)
      expect([from.value, to.value]).toEqual([1, 10])

      page.value = 2
      expect(pageData.value.items).toHaveLength(10)
      expect(pageData.value.items[0]?.id).toBe(11)

      page.value = 4
      expect(pageData.value.items.map(r => r.id)).toEqual([31, 32])
      expect([from.value, to.value]).toEqual([31, 32])

      // A list that shrank below the current page is clamped to its last page
      list.value = rows.slice(0, 15)
      expect(pageData.value.page).toBe(2)
      expect(pageData.value.items.map(r => r.id)).toEqual([11, 12, 13, 14, 15])

      list.value = []
      expect([from.value, to.value]).toEqual([0, 0])
    })

    it('goes back to page 1 when search, status filter or year changes', async () => {
      const { ref, nextTick } = await import('vue')
      const { usePagedList } = await import('~/composables/usePagedList')
      const search = ref('')
      const status = ref('all')
      const year = ref(2026)
      const { page, pageData } = usePagedList(() => rows, [search, status, year])

      const changes = [
        () => {
          search.value = 'x'
        },
        () => {
          status.value = 'prioritized'
        },
        () => {
          year.value = 2025
        },
        // Reset button
        () => {
          search.value = ''
          status.value = 'all'
        }
      ]
      for (const change of changes) {
        page.value = 3
        await nextTick()
        change()
        await nextTick()
        expect(page.value).toBe(1)
        expect(pageData.value.items[0]?.id).toBe(1)
      }
    })

    it('renders a pager over the paged rows in the page', () => {
      const content = fs.readFileSync(path.join(rootDir, 'pages/risk-profile/audit-priority/index.vue'), 'utf8')
      expect(content).toContain('<UPagination')
      expect(content).toContain(':data="priorityPage.items"')
      // Fixed page size of 10
      expect(content).toContain('usePagedList(filteredYearlyUniverse, [searchQuery, statusFilter, selectedYear], 10)')
    })
  })

  it('verifies localization keys for auditPriority exist in en and id locales', () => {
    const enCommon = JSON.parse(fs.readFileSync(path.join(rootDir, 'locales/en/common.json'), 'utf8'))
    const idCommon = JSON.parse(fs.readFileSync(path.join(rootDir, 'locales/id/common.json'), 'utf8'))

    expect(enCommon.auditPriority).toBeDefined()
    expect(enCommon.auditPriority.title).toBe('Audit Priority')

    expect(idCommon.auditPriority).toBeDefined()
    expect(idCommon.auditPriority.title).toBe('Audit Priority')
  })
})
