import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const read = (path: string) => readFileSync(resolve(__dirname, '../../', path), 'utf8')

describe('Disabled, read-only, and view-only field styling', () => {
  it('configures darker disabled and read-only backgrounds across Nuxt UI form components in app.config.ts', () => {
    const appConfig = read('app.config.ts')
    for (const component of ['input', 'textarea', 'inputNumber', 'inputMenu', 'selectMenu', 'select', 'checkbox', 'radioGroup', 'switch', 'fileUpload']) {
      expect(appConfig).toContain(`${component}:`)
    }
    expect(appConfig).toContain('disabled:bg-slate-200')
    expect(appConfig).toContain('read-only:bg-slate-200')
    expect(appConfig).toContain('data-[disabled]:bg-slate-200')
    expect(appConfig).toContain('dark:disabled:bg-slate-950/80')
    expect(appConfig).toContain('dark:read-only:bg-slate-950/80')
  })

  it('styles native disabled, readonly, and aria-readonly form fields with a darker surface in main.css', () => {
    const mainCss = read('assets/css/main.css')
    expect(mainCss).toContain('textarea:disabled')
    expect(mainCss).toContain('textarea[readonly]')
    expect(mainCss).toContain('textarea[aria-readonly="true"]')
    expect(mainCss).toContain('select:disabled')
    expect(mainCss).toContain('[role="combobox"][aria-disabled="true"]')
    expect(mainCss).toContain('[role="combobox"][data-disabled]')
    expect(mainCss).toContain('.input-field[readonly]')
    expect(mainCss).toContain('background-color: #e2e8f0 !important;')
    expect(mainCss).toContain('background-color: #020617 !important;')
  })

  it('styles disabled AppDatePicker inputs with the darker slate surface', () => {
    const datePicker = read('components/shared/AppDatePicker.vue')
    expect(datePicker).toContain('bg-slate-200 dark:bg-slate-950/80')
    expect(datePicker).not.toContain('opacity-50 cursor-not-allowed bg-gray-100')
  })
})
