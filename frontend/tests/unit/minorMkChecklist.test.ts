import { describe, expect, it } from 'vitest'
import fs from 'node:fs'
import path from 'node:path'

const read = (relativePath: string) => fs.readFileSync(path.resolve(__dirname, '../../', relativePath), 'utf8')

describe('PROD minor checklist items assigned to Mk', () => {
  it('uses pointer cursors for all enabled interactive controls', () => {
    const css = read('assets/css/main.css')

    for (const selector of [
      'button:not(:disabled)',
      'select:not(:disabled)',
      'input[type="checkbox"]:not(:disabled)',
      '[role="tab"]:not([aria-disabled="true"])',
      '[role="combobox"]:not([aria-disabled="true"])'
    ]) {
      expect(css).toContain(selector)
    }
  })

  it('uses the required black dark-mode fields in the Audit Charter form', () => {
    const form = read('components/audit-charter/AuditCharterForm.vue')

    expect(form.match(/bg-gray-50 dark:bg-black/g)).toHaveLength(2)
    expect(form).toContain('class="dark:!bg-black"')
  })

  it('sorts PIC and Unit in Charge options alphabetically', () => {
    const store = read('stores/mitigation-risk.ts')

    expect(store).toMatch(/const picOptions =[\s\S]*?\.sort\(\(a, b\) => a\.localeCompare\(b, 'id'/)
    expect(store).toMatch(/const unitInChargeOptions =[\s\S]*?\.sort\(\(a, b\) => a\.localeCompare\(b, 'id'/)
  })

  it('provides an explicit close action in the RCM add/edit modal', () => {
    const rcm = read('pages/risk-profile/risk-control-matrix/index.vue')

    expect(rcm).toContain('aria-label="Tutup form Risk Control Matrix"')
    expect(rcm).toContain('@click="isModalOpen = false"')
  })

  it('marks Annual Audit scheduling fields as required and keeps generated fields read-only', () => {
    const form = read('components/annual-audit/AnnualAuditForm.vue')

    for (const label of [
      'label="Select Months"',
      'label="Number of Auditors (1-10)"',
      'label="Duration (Days)"',
      'label="Supervisor"'
    ]) {
      const labelPosition = form.indexOf(label)
      expect(labelPosition).toBeGreaterThan(-1)
      expect(form.slice(labelPosition, labelPosition + 160)).toContain('required')
    }

    expect(form).toMatch(/v-model="store\.form\.code"[\s\S]*?disabled/)
    expect(form).toMatch(/v-model="store\.form\.attachmentUploadedBy"[\s\S]*?disabled/)
    expect(form).toContain('admin: \'System Administrator\'')
    expect(form).toContain('store.form.attachmentUploadedBy = signedInRoleLabel.value')
  })

  it('groups Audit Execution Status under Assignment Letter instead of Annual Audit Plan', () => {
    const layout = read('layouts/default.vue')
    const annualStart = layout.indexOf('label: t(\'navigation.annualAuditPlan\')')
    const activityPlanStart = layout.indexOf('label: t(\'navigation.auditActivityPlan\')', annualStart)
    const assignmentStart = layout.indexOf('label: t(\'navigation.assignmentLetter\')', activityPlanStart)
    const workingPaperStart = layout.indexOf('label: t(\'navigation.workingPaper\')', assignmentStart)

    expect(layout.slice(annualStart, activityPlanStart)).not.toContain('to: \'/audit-execution-status\'')
    expect(layout.slice(assignmentStart, workingPaperStart)).toContain('to: \'/audit-execution-status\'')
  })

  it('generates and disables the Performance Report document title', () => {
    const upload = read('pages/kpi-performance/upload.vue')

    expect(upload).toContain('const updateGeneratedTitle = () =>')
    expect(upload).toContain('watch([() => form.value.period, () => form.value.year, locale], updateGeneratedTitle)')
    expect(upload).toMatch(/v-model="form\.title"[\s\S]*?disabled/)
  })
})
