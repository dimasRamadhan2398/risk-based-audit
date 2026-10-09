import { describe, it, expect, vi } from 'vitest'
import { useI18n } from '~/composables/useI18n'

describe('Fieldwork and Common Translations', () => {
  it('should correctly resolve common.actions translations without raw keys', () => {
    const { t, setLocale } = useI18n()
    
    // Test ID locale
    setLocale('id')
    expect(t('common.actions.view')).toBe('Lihat')
    expect(t('common.actions.download')).toBe('Unduh')
    expect(t('common.actions.edit')).toBe('Ubah')
    expect(t('common.actions.delete')).toBe('Hapus')
    expect(t('common.cancel')).toBe('Batal')
    expect(t('common.close')).toBe('Tutup')
    expect(t('common.submit')).toBe('Simpan')
    expect(t('auditFieldwork.sample.title')).toBe('Data Sampel')
    expect(t('auditFieldwork.sample.columns.file')).toBe('Dokumen Lampiran')

    // Test EN locale
    setLocale('en')
    expect(t('common.actions.view')).toBe('View')
    expect(t('common.actions.download')).toBe('Download')
    expect(t('common.actions.edit')).toBe('Edit')
    expect(t('common.actions.delete')).toBe('Delete')
    expect(t('common.cancel')).toBe('Cancel')
    expect(t('common.close')).toBe('Close')
    expect(t('common.submit')).toBe('Submit')
    expect(t('auditFieldwork.sample.title')).toBe('Sample Data')
    expect(t('auditFieldwork.sample.columns.file')).toBe('Attached Document')
  })
})
