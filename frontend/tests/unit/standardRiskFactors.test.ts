import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Standard Risk Factors CRUD Feature', () => {
  const rootDir = path.resolve(__dirname, '../../')

  it('verifies risk-factors index.vue contains Add, Edit, Delete UI elements and modals', () => {
    const filePath = path.join(rootDir, 'pages/risk-profile/risk-factors/index.vue')
    const content = fs.readFileSync(filePath, 'utf8')

    // Add button and actions
    expect(content).toContain('openAddFactorModal')
    expect(content).toContain('openEditFactorModal')
    expect(content).toContain('openDeleteFactorModal')

    // Modals
    expect(content).toContain('factorModalOpen')
    expect(content).toContain('deleteModalOpen')
    expect(content).toContain('saveFactorForm')
    expect(content).toContain('confirmDeleteFactor')
  })

  it('verifies risk-factors store contains CRUD methods for standard risk factors', () => {
    const filePath = path.join(rootDir, 'stores/risk-factors.ts')
    const content = fs.readFileSync(filePath, 'utf8')

    expect(content).toContain('createStandardFactor')
    expect(content).toContain('updateStandardFactor')
    expect(content).toContain('deleteStandardFactor')
    expect(content).toContain("method: 'POST'")
    expect(content).toContain("method: 'PUT'")
    expect(content).toContain("method: 'DELETE'")
  })

  it('verifies localization keys for Standard Risk Factors CRUD exist in en and id locales', () => {
    const enCommon = JSON.parse(fs.readFileSync(path.join(rootDir, 'locales/en/common.json'), 'utf8'))
    const idCommon = JSON.parse(fs.readFileSync(path.join(rootDir, 'locales/id/common.json'), 'utf8'))

    expect(enCommon.riskFactors.weighting.addFactor).toBe('Add Factor')
    expect(enCommon.riskFactors.weighting.editFactor).toBe('Edit Factor')
    expect(enCommon.riskFactors.weighting.deleteFactor).toBe('Delete Factor')
    expect(enCommon.riskFactors.messages.factorCreated).toBeDefined()
    expect(enCommon.riskFactors.messages.factorUpdated).toBeDefined()
    expect(enCommon.riskFactors.messages.factorDeleted).toBeDefined()

    expect(idCommon.riskFactors.weighting.addFactor).toBe('Tambah Faktor')
    expect(idCommon.riskFactors.weighting.editFactor).toBe('Ubah Faktor')
    expect(idCommon.riskFactors.weighting.deleteFactor).toBe('Hapus Faktor')
    expect(idCommon.riskFactors.messages.factorCreated).toBeDefined()
    expect(idCommon.riskFactors.messages.factorUpdated).toBeDefined()
    expect(idCommon.riskFactors.messages.factorDeleted).toBeDefined()
  })
})
