import { describe, expect, it } from 'vitest'
import fs from 'node:fs'
import path from 'node:path'

const read = (relativePath: string) => fs.readFileSync(path.resolve(__dirname, '../../', relativePath), 'utf8')

describe('SignatureBox (reusable digital signature)', () => {
  const box = read('components/shared/SignatureBox.vue')

  it('draws with pointer events so mouse, touch and pen share one path', () => {
    expect(box).toContain('@pointerdown="onPointerDown"')
    expect(box).toContain('@pointermove="onPointerMove"')
    expect(box).toContain('@pointerup="onPointerUp"')
    expect(box).toContain('touch-none')
    expect(box).toContain('setPointerCapture')
  })

  it('exports a PNG data URL and clears the value on clear', () => {
    expect(box).toContain('toDataURL(\'image/png\')')
    expect(box).toMatch(/const clear = \(\) => \{[\s\S]*?emit\('update:modelValue', \{ \.\.\.props\.modelValue, image: '', signedAt: '' \}\)/)
  })

  it('sizes the drawing buffer from layout size so the modal zoom-in animation cannot blur it', () => {
    const setup = box.slice(box.indexOf('const setupCanvas'), box.indexOf('const emitValue'))
    expect(setup).toContain('layoutSize(canvas)')
    expect(setup).not.toContain('getBoundingClientRect')
    expect(box).toContain('canvas.clientWidth')
  })

  it('records signer name, title and a timestamp set at the first stroke', () => {
    expect(box).toContain('t(\'common.signatureBox.name\')')
    expect(box).toContain('t(\'common.signatureBox.title\')')
    expect(box).toContain('signedAt: image ? (props.modelValue.signedAt || new Date().toISOString()) : \'\'')
  })

  it('has i18n keys for every signature box label', () => {
    const en = JSON.parse(read('locales/en/common.json'))
    const id = JSON.parse(read('locales/id/common.json'))
    for (const key of ['notSigned', 'signHere', 'hint', 'name', 'namePlaceholder', 'title', 'titlePlaceholder', 'timestamp']) {
      expect(en.common.signatureBox[key]).toBeTruthy()
      expect(id.common.signatureBox[key]).toBeTruthy()
    }
  })
})

describe('ExecutiveSummaryApproveModal (signed finalize)', () => {
  const modal = read('components/audit-result-report/ExecutiveSummaryApproveModal.vue')

  it('requires place, date, signature and signer name before approval is enabled', () => {
    expect(modal).toMatch(/canSubmit = computed\(\(\) => Boolean\(\s*place\.value\.trim\(\) && date\.value && signature\.value\.image && signature\.value\.name\.trim\(\)/)
    expect(modal).toContain(':disabled="!canSubmit"')
  })

  it('sends the approve payload fields the backend expects', () => {
    for (const field of ['signaturePlace', 'signatureDate', 'signatureImage', 'signerName', 'signerRole']) {
      expect(modal).toContain(`${field}:`)
    }
    expect(modal).toContain('store.approveWithSignature(payload)')
  })

  it('confirms through the global confirmation modal before submitting', () => {
    expect(modal).toContain('useGlobalModalStore().confirmSubmit')
  })
})
