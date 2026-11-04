import { describe, it, expect } from 'vitest'
import fs from 'node:fs'
import path from 'node:path'

const read = (relativePath: string) => fs.readFileSync(path.resolve(__dirname, '../../', relativePath), 'utf8')

const individual = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
const compilation = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')
const cards = read('components/audit-result-report/SignificantFindingCards.vue')

describe('significant findings cards', () => {
  it('renders each finding as a collapsible card with the five required fields', () => {
    expect(cards).toContain('<UCollapsible')
    expect(cards).toContain('v-for="(finding, idx) in findings"')
    for (const key of ['table.unitDivision', 'table.findingTitle', 'table.riskValue', 'table.followUpStatus', 'table.proposedAction']) {
      expect(cards).toContain(`t('executiveSummaryForm.${key}')`)
    }
  })

  it('tracks open state by finding identity and starts collapsed', () => {
    expect(cards).toContain(':open="isExpanded(finding)"')
    expect(cards).toContain('createFindingExpansion(reactive(new Set<object>()))')
    expect(cards).toContain('defineExpose({ expand })')
  })

  it('both forms use the shared card list and open only the newly added card', () => {
    for (const form of [individual, compilation]) {
      expect(form).toContain('<SignificantFindingCards')
      expect(form).toContain('findingCardsRef.value?.expand(store.form.topFindings[store.form.topFindings.length - 1])')
      expect(form).not.toMatch(/<tr v-for="\(finding, idx\) in store\.form\.topFindings"/)
    }
  })

  it('the findings section heading no longer carries the "5" count', () => {
    const en = JSON.parse(read('locales/en/common.json'))
    const id = JSON.parse(read('locales/id/common.json'))
    expect(en.executiveSummaryForm.topFindings.title).toBe('Top Significant Findings')
    expect(id.executiveSummaryForm.topFindings.title).toBe('Temuan Signifikan Teratas')
    for (const locale of [en, id]) {
      for (const key of ['cardTitle', 'untitled', 'remove']) {
        expect(locale.executiveSummaryForm.topFindings[key]).toBeTruthy()
      }
    }
  })
})

describe('executive summary stepper header', () => {
  it('scrolls horizontally only and keeps the active step in view', () => {
    for (const form of [individual, compilation]) {
      expect(form).toContain('overflow-x-auto overflow-y-hidden')
      expect(form).toContain('ref="stepScrollRef"')
      expect(form).toContain('watch(currentStepIndex, () => scrollActiveStepIntoView(), { flush: \'post\' })')
      expect(form).toContain('scrollStepIntoView(container, step)')
    }
  })
})

describe('approval entry point', () => {
  it('both forms open the signed approval modal from the Approve button', () => {
    for (const form of [individual, compilation]) {
      expect(form).toContain('@click="openApproveSignature"')
      expect(form).toContain('<ExecutiveSummaryApproveModal')
      expect(form).toContain('v-model:open="showApproveModal"')
    }
  })
})
