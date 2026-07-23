import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import type { VisionMissionGoals, Company } from '~/types/master'

export const useVisionMissionGoalsStore = defineStore('vision-mission-goals', () => {
  // Config
  const config = useRuntimeConfig()
  const getMasterServiceBaseUrl = () => {
    return config.public.masterServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  // State
  const activeVmg = ref<VisionMissionGoals | null>(null)
  const companies = ref<Company[]>([])
  const selectedCompanyId = ref<string>('')
  const loading = ref(false)
  const saving = ref(false)
  const errorMsg = ref('')
  const isModalOpen = ref(false)
  const isEditMode = ref(false)

  // Form State
  const form = reactive({
    id: '',
    visis: [''] as string[],
    misis: [''] as string[],
    goals: [{ id: '', goal_code: 'G-001', goal_name: '' }] as Array<{ id?: string; goal_code: string; goal_name: string }>,
    yearStart: new Date().getFullYear(),
    yearEnd: new Date().getFullYear() + 4
  })

  // Year options for dropdown (e.g. from current year - 10 to current year + 20)
  const currentYear = new Date().getFullYear()
  const yearOptions = Array.from({ length: 31 }, (_, i) => {
    const year = currentYear - 10 + i
    return { label: String(year), value: year }
  })

  // Fetch all companies and then fetch the active VMG
  const fetchCompaniesAndVmg = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const response = await $fetch<any>(`${baseUrl}/companies`, {
        method: 'GET',
        params: { page: 1, page_size: 100 }
      })

      const list = Array.isArray(response.data)
        ? response.data
        : (response.data?.companies || response.companies || [])
      companies.value = Array.isArray(list) ? list : []

      if (companies.value.length > 0) {
        // Set default selected company (prefer holding company if exists, else first)
        const holding = companies.value.find(c => c.company_type === 'HOLDING')
        selectedCompanyId.value = holding?.id || companies.value[0]?.id || ''
        await fetchActiveVmg()
      }
    } catch (error: any) {
      console.error('Failed to fetch companies:', error)
      errorMsg.value = 'Failed to load company data.'
    } finally {
      loading.value = false
    }
  }

  // Fetch VMGs for the selected company and choose the active one
  const fetchActiveVmg = async () => {
    if (!selectedCompanyId.value) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const response = await $fetch<any>(`${baseUrl}/vision-mission-goals/company/${selectedCompanyId.value}`, {
        method: 'GET'
      })

      const items = response.data || response || []
      if (Array.isArray(items) && items.length > 0) {
        // Select published first, else just the latest one
        const published = items.find((i: any) => i.status === 'Published')
        activeVmg.value = published || items[0]
      } else {
        activeVmg.value = null
      }
    } catch (error: any) {
      console.error('Failed to fetch VMG:', error)
      activeVmg.value = null
    } finally {
      loading.value = false
    }
  }

  // Open Modal (Add or Edit Mode)
  const openModal = () => {
    isEditMode.value = !!activeVmg.value
    errorMsg.value = ''

    if (activeVmg.value) {
      form.id = activeVmg.value.id
      // Split vision string by newline
      form.visis = activeVmg.value.vision ? activeVmg.value.vision.split('\n') : ['']
      // Split mission string by newline
      form.misis = activeVmg.value.mission ? activeVmg.value.mission.split('\n') : ['']
      
      // Populate goals
      form.goals = activeVmg.value.goals && activeVmg.value.goals.length > 0 
        ? activeVmg.value.goals.map(g => ({
            id: g.id,
            goal_code: g.goal_code,
            goal_name: g.goal_name
          }))
        : [{ goal_code: 'G-001', goal_name: '' }]
      
      // Parse period string e.g., "2026 - 2030"
      if (activeVmg.value.period) {
        const parts = activeVmg.value.period.split('-')
        if (parts.length === 2) {
          const first = parts[0]
          const second = parts[1]
          if (first && second) {
            form.yearStart = parseInt(first.trim(), 10) || currentYear
            form.yearEnd = parseInt(second.trim(), 10) || (currentYear + 4)
          }
        }
      }
    } else {
      form.id = ''
      form.visis = ['']
      form.misis = ['']
      form.goals = [{ goal_code: 'G-001', goal_name: '' }]
      form.yearStart = currentYear
      form.yearEnd = currentYear + 4
    }
    
    isModalOpen.value = true
  }

  // Save VMG
  const saveVmg = async () => {
    errorMsg.value = ''
    
    // Clean empty values
    const cleanVisis = form.visis.map(v => v.trim()).filter(Boolean)
    const cleanMisis = form.misis.map(m => m.trim()).filter(Boolean)
    const cleanGoals = form.goals.filter(g => g.goal_name.trim())

    if (cleanVisis.length === 0) {
      errorMsg.value = 'Visi harus diisi minimal 1.'
      return false
    }
    if (cleanMisis.length === 0) {
      errorMsg.value = 'Misi harus diisi minimal 1.'
      return false
    }
    if (cleanGoals.length === 0) {
      errorMsg.value = 'Goal harus diisi minimal 1.'
      return false
    }
    if (form.yearEnd < form.yearStart) {
      errorMsg.value = 'Tahun target akhir tidak boleh kurang dari tahun target awal.'
      return false
    }

    saving.value = true
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const payload = {
        company_id: selectedCompanyId.value,
        period: `${form.yearStart} - ${form.yearEnd}`,
        effective_date: `${form.yearStart}-01-01`,
        vision: cleanVisis.join('\n'),
        mission: cleanMisis.join('\n'),
        status: 'Published', // Set directly to Published so it displays
        version: 'v1.0',
        notes: isEditMode.value ? 'Updated via Strategic Audit Plan UI' : 'Created via Strategic Audit Plan UI',
        goals: cleanGoals.map((g, idx) => ({
          id: g.id || undefined,
          goal_code: g.goal_code || `G-00${idx + 1}`,
          goal_name: g.goal_name.trim()
        }))
      }

      if (isEditMode.value && form.id) {
        await $fetch(`${baseUrl}/vision-mission-goals/${form.id}`, {
          method: 'PUT',
          body: payload
        })
      } else {
        await $fetch(`${baseUrl}/vision-mission-goals`, {
          method: 'POST',
          body: payload
        })
      }

      // Re-fetch to update local state
      await fetchActiveVmg()
      isModalOpen.value = false
      return true
    } catch (error: any) {
      console.error('Failed to save VMG:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal menyimpan data Visi, Misi dan Goals.'
      return false
    } finally {
      saving.value = false
    }
  }

  return {
    activeVmg,
    companies,
    selectedCompanyId,
    loading,
    saving,
    errorMsg,
    isModalOpen,
    isEditMode,
    form,
    yearOptions,
    fetchCompaniesAndVmg,
    fetchActiveVmg,
    openModal,
    saveVmg
  }
})
