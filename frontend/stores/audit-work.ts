// stores/kka.ts
import { defineStore } from 'pinia'
import type { AuditWorkingForm, TestStep, AuditWorkingItem } from '~/types/audit'

export const useAuditWorkingStore = defineStore('audit-work', () => {
  const form = reactive<AuditWorkingForm>({
    assignmentNumber: '',
    teamMembers: '',
    processName: '',
    activityName: '',
    period: '',
    riskDescription: '',
    riskTaxonomy: 'Operational',
    appName: '',
    populationSize: 0,
    testSteps: [],
    items: [],
    findingTitle: '',
    condition: '',
    criteria: '',
    cause: '',
    effect: '',
    evidenceFile: null,
    rcaCategory: 'Process',
    why1: '',
    why2: '',
    why3: '',
    rootCauseConclusion: '',
    auditeeResponse: '',
    actionPlan: '',
    targetDate: '',
    pic: '',
    approver: '',
    status: 'Draft'
  })

  // F-02 Logic: Tambah Langkah Kerja
  const addTestStep = (desc: string) => {
    const newStep: TestStep = {
      id: `step_${Date.now()}`,
      description: desc
    }
    form.testSteps.push(newStep)
    
    // Sinkronisasi: Update existing samples dengan key step baru (default N/A)
    form.items.forEach(item => {
      item.results[newStep.id] = 'N/A'
    })
  }

  // F-03 Logic: Tambah Sampel
  const addItem = (name: string) => {
    // Validasi Populasi
    if (form.items.length >= form.populationSize && form.populationSize > 0) {
      throw new Error("Jumlah sampel tidak boleh melebihi total populasi!")
    }

    const results: Record<string, any> = {}
    form.testSteps.forEach(step => {
      results[step.id] = 'N/A'
    })

    const newItem: AuditWorkingItem = {
      id: `smpl_${Date.now()}`,
      auditWorkingName: name,
      results: results,
      isEffective: true // Default
    }
    form.items.push(newItem)
  }

  // F-03 Logic: Hitung Efektivitas per Sampel
  const updateItemResult = (itemId: string, stepId: string, result: 'Pass' | 'Fail' | 'N/A') => {
    const item = form.items.find(s => s.id === itemId)
    if (item) {
      item.results[stepId] = result
      
      // Auto Calculation: Jika ada 1 Fail, maka Not Effective
      const hasFail = Object.values(item.results).includes('Fail')
      item.isEffective = !hasFail
    }
  }

  return {
    form,
    addTestStep,
    addItem,
    updateItemResult
  }
})