import { describe, it, expect } from 'vitest'
import { extractTextFromClientFile } from '~/utils/documentTextExtractor'

describe('documentTextExtractor client-side extraction', () => {
  it('extracts plain text file correctly', async () => {
    const content = 'Hasil audit kepatuhan regulasi POJK pada cabang Surabaya.'
    const file = new File([content], 'audit_kepatuhan.txt', { type: 'text/plain' })
    const result = await extractTextFromClientFile(file)
    expect(result.text).toBe(content)
    expect(result.method).toBe('Direct Client Text Reader')
  })

  it('extracts CSV structured rows correctly', async () => {
    const csvContent = 'ID,Temuan,Rekomendasi\n1,Kelemahan kontrol kas,Perketat otorisasi'
    const file = new File([csvContent], 'laporan_temuan.csv', { type: 'text/csv' })
    const result = await extractTextFromClientFile(file)
    expect(result.text).toContain('Kelemahan kontrol kas')
  })
})

