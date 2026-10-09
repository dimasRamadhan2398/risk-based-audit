import { describe, it, expect } from 'vitest'
import SignaturePad from '~/components/shared/SignaturePad.vue'

describe('SignaturePad Component Import', () => {
  it('parses and exports component cleanly without syntax errors', () => {
    expect(SignaturePad).toBeDefined()
    expect(SignaturePad.__name).toBe('SignaturePad')
  })
})
