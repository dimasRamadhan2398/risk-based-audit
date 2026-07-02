// stores/assignment-letter.ts
import type { TableColumn } from '@nuxt/ui';
import { defineStore } from 'pinia'
import { type AssignmentLetter, type AssignmentLetterForm, type AssignmentLetterStatus, AuditCategory } from '~/types/audit'

export interface AssignmentLetterState {
  isModalOpen: boolean;
  assignmentLetterList: AssignmentLetter[];
  form: AssignmentLetterForm;
  columns: TableColumn<AssignmentLetter>[];
  options: {
    auditTeam: string[];
    workingUnit: string[];
    role: string[];
  };
  loading: boolean;
  errorMsg: string;
}

export const useAssignmentLetterStore = defineStore('assignment-letter', {
  state: (): AssignmentLetterState => ({
    isModalOpen: false,
    assignmentLetterList: [
      {
        id: 'mock-uuid-zeta-1',
        letterNumber: 'ST-001/SKAI/2026',
        status: 'Published' as AssignmentLetterStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Real Auditore',
        leader: 'Zeta Ramadhani',
        category: AuditCategory.ASSURANCE,
        auditYear: new Date().getFullYear().toString(),
        auditTeam: 'SKAI',
        startPeriod: '2026-03-01',
        finishPeriod: '2026-03-31',
        workingUnit: 'Finance',
        auditPurpose: 'Annual Audit',
        letterDate: '2026-01-01',
        caeSignature: 'System',
        executionPeriod: '2026-03-01 to 2026-03-31',
        membersList: [
          { name: 'Zeta Ramadhani', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' },
          { name: 'Rina Wulandari', role: 'Member' },
          { name: 'Andi Firmansyah', role: 'Member' },
          { name: 'Dewi Kusumawati', role: 'Person in Charge' }
        ],
        purposeList: [
          'Menilai efektivitas pengendalian internal pada divisi keuangan',
          'Memastikan kepatuhan terhadap kebijakan dan prosedur perusahaan',
          'Mengidentifikasi potensi risiko yang dapat mempengaruhi operasional keuangan'
        ],
        scopeList: [
          'Proses pencatatan dan pelaporan keuangan periode Januari - Desember 2025',
          'Pengelolaan kas dan setara kas',
          'Rekonsiliasi bank dan laporan arus kas',
          'Kepatuhan terhadap standar akuntansi yang berlaku (PSAK)'
        ],
        ccList: [
          'President Director',
          'Chief Financial Officer',
          'Head of Internal Audit',
          'Board of Commissioners'
        ]
      }
    ],
    loading: false,
    errorMsg: '',
    form: {
      auditTitle: '',
      leader: '',
      category: AuditCategory.ASSURANCE,
      auditYear: new Date().getFullYear().toString(),
      auditTeam: 'SKAI',
      startPeriod: '',
      finishPeriod: '',
      workingUnit: '',
      auditPurpose: '',
      letterDate: '',
      caeSignature: '',
      membersList: [
        { name: '', role: 'Chairperson' },
        { name: '', role: 'Member' }
      ],
      purposeList: [''],
      scopeList: [''],
      ccList: ['President Director']
    },
    columns: [
      { accessorKey: 'letterNumber', header: 'Letter Number' },
      { accessorKey: 'assignmentTitle', header: 'Audit Title / Object' },
      { accessorKey: 'workingUnit', header: 'Work Unit' },
      { accessorKey: 'executionPeriod', header: 'Execution Period' },
      { accessorKey: 'auditTeam', header: 'Audit Team' },
      { accessorKey: 'status', header: 'Status' }
    ],
    options: {
      auditTeam: ['SKAI', 'DAI', 'CAE'],
      workingUnit: ['Production', 'Marketing', 'Finance'],
      role: ['Person in Charge', 'Supervisor', 'Chairperson', 'Member'],
    }
  }),

  getters: {
    getColumns: (state) => state.columns,
    getAssignmentLetterList: (state) => state.assignmentLetterList,
    getForm: (state) => state.form,
    getOptions: (state) => state.options,

    dateError: (state): string | null => {
      if (state.form.startPeriod && state.form.finishPeriod) {
        const start = new Date(state.form.startPeriod)
        const end = new Date(state.form.finishPeriod)
        if (end < start) {
          return "Error: End date cannot be before start date."
        }
      }
      return null
    }
  },
  actions: {
    getAuditServiceBaseUrl() {
      const config = useRuntimeConfig()
      return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
    },

    async fetchAssignmentLetters() {
      this.loading = true
      this.errorMsg = ''
      const mockList = [
        {
          id: 'mock-uuid-zeta-1',
          letterNumber: 'ST-001/SKAI/2026',
          status: 'Published' as AssignmentLetterStatus,
          createdAt: new Date().toISOString(),
          auditTitle: 'Real Auditore',
          leader: 'Zeta Ramadhani',
          category: AuditCategory.ASSURANCE,
          auditYear: new Date().getFullYear().toString(),
          auditTeam: 'SKAI',
          startPeriod: '2026-03-01',
          finishPeriod: '2026-03-31',
          workingUnit: 'Finance',
          auditPurpose: 'Annual Audit',
          letterDate: '2026-01-01',
          caeSignature: 'System',
          executionPeriod: '2026-03-01 to 2026-03-31',
          membersList: [
            { name: 'Zeta Ramadhani', role: 'Chairperson' },
            { name: 'Budi Santoso', role: 'Supervisor' },
            { name: 'Rina Wulandari', role: 'Member' },
            { name: 'Andi Firmansyah', role: 'Member' },
            { name: 'Dewi Kusumawati', role: 'Person in Charge' }
          ],
          purposeList: [
            'Menilai efektivitas pengendalian internal pada divisi keuangan',
            'Memastikan kepatuhan terhadap kebijakan dan prosedur perusahaan',
            'Mengidentifikasi potensi risiko yang dapat mempengaruhi operasional keuangan'
          ],
          scopeList: [
            'Proses pencatatan dan pelaporan keuangan periode Januari - Desember 2025',
            'Pengelolaan kas dan setara kas',
            'Rekonsiliasi bank dan laporan arus kas',
            'Kepatuhan terhadap standar akuntansi yang berlaku (PSAK)'
          ],
          ccList: [
            'President Director',
            'Chief Financial Officer',
            'Head of Internal Audit',
            'Board of Commissioners'
          ]
        }
      ] as AssignmentLetter[]

      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        const response: any = await $fetch(`${baseUrl}/assignment-letters`, {
          method: 'GET'
        })
        let items: AssignmentLetter[] = []
        if (response && response.data && Array.isArray(response.data.items)) {
        items = response.data.items
      } else if (response && Array.isArray(response.items)) {
        items = response.items
        } else if (Array.isArray(response)) {
          items = response
        }

        if (items.length > 0) {
          this.assignmentLetterList = items
        } else {
          this.assignmentLetterList = mockList
        }
      } catch (error: any) {
        console.error('Failed to fetch assignment letters:', error)
        this.errorMsg = 'Failed to load assignment letters.'
        this.assignmentLetterList = mockList
      } finally {
        this.loading = false
      }
    },

    addItem(list: any[], defaultItem: any) {
      list.push(typeof defaultItem === 'object' ? { ...defaultItem } : defaultItem)
    },

    removeItem(list: any[], index: number) {
      list.splice(index, 1)
    },

    generateNomorSurat(auditTeam: string, year: string): string {
      const nextCount = this.assignmentLetterList.length + 1
      const paddedCount = nextCount.toString().padStart(3, '0')
      return `ST-${paddedCount}/${auditTeam}/${year}`
    },

    openModal() {
      Object.assign(this.form, {
        auditTitle: '',
        leader: '',
        category: AuditCategory.ASSURANCE,
        auditYear: new Date().getFullYear().toString(),
        auditTeam: 'SKAI',
        startPeriod: '',
        finishPeriod: '',
        workingUnit: '',
        auditPurpose: '',
        letterDate: '',
        caeSignature: '',
        membersList: [{ name: '', role: 'Chairperson' }],
        purposeList: [''],
        scopeList: [''],
        ccList: ['President Director']
      })
      this.isModalOpen = true
    },

    closeModal() {
      this.isModalOpen = false
    },

    async handleSubmit() {
      if (!this.form.workingUnit) {
        alert("Work unit must be filled in.")
        return
      }

      if (this.dateError || !this.form.startPeriod || !this.form.finishPeriod) {
        alert(this.dateError || "Please fill in start and end period.")
        return
      }

      if (this.form.membersList.length < 3) {
        const proceed = confirm("Template suggests at least 3 team members. Continue saving?")
        if (!proceed) return
      }

      this.loading = true
      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        const letterNumber = this.generateNomorSurat(this.form.auditTeam, this.form.auditYear)
        const executionPeriod = `${this.form.startPeriod} to ${this.form.finishPeriod}`

        const payload = {
          letterNumber,
          status: 'Draft',
          auditTitle: this.form.auditTitle,
          leader: this.form.leader,
          category: this.form.category,
          auditYear: this.form.auditYear,
          auditTeam: this.form.auditTeam,
          startPeriod: this.form.startPeriod,
          finishPeriod: this.form.finishPeriod,
          workingUnit: this.form.workingUnit,
          auditPurpose: this.form.auditPurpose,
          letterDate: this.form.letterDate,
          caeSignature: this.form.caeSignature,
          executionPeriod,
          membersList: this.form.membersList,
          purposeList: this.form.purposeList,
          scopeList: this.form.scopeList,
          ccList: this.form.ccList
        }

        await $fetch(`${baseUrl}/assignment-letters`, {
          method: 'POST',
          body: payload
        })
        this.closeModal()
        await this.fetchAssignmentLetters()
      } catch (error: any) {
        console.error('Failed to create assignment letter:', error)
        alert('Failed to save assignment letter.')
      } finally {
        this.loading = false
      }
    },

    async addAssignmentLetter(form: AssignmentLetterForm) {
      this.loading = true
      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        const letterNumber = this.generateNomorSurat(form.auditTeam, form.auditYear)
        const executionPeriod = `${form.startPeriod} to ${form.finishPeriod}`

        const payload = {
          ...form,
          letterNumber,
          executionPeriod,
          status: 'Draft'
        }

        await $fetch(`${baseUrl}/assignment-letters`, {
          method: 'POST',
          body: payload
        })
        await this.fetchAssignmentLetters()
      } catch (error) {
        console.error(error)
      } finally {
        this.loading = false
      }
    },

    async deleteSuratTugas(id: string) {
      if (!confirm("Are you sure you want to delete this assignment letter?")) return
      this.loading = true
      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/assignment-letters/${id}`, {
          method: 'DELETE'
        })
        await this.fetchAssignmentLetters()
      } catch (error) {
        console.error(error)
      } finally {
        this.loading = false
      }
    },

    async changeStatus(id: string, status: AssignmentLetterStatus) {
      this.loading = true
      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        // Find existing to construct payload
        const existing = this.assignmentLetterList.find(s => s.id === id)
        if (existing) {
          const payload = { ...existing, status }
          await $fetch(`${baseUrl}/assignment-letters/${id}`, {
            method: 'PUT',
            body: payload
          })
          await this.fetchAssignmentLetters()
        }
      } catch (error) {
        console.error(error)
      } finally {
        this.loading = false
      }
    }
  }
})