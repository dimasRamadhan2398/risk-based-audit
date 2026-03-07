// stores/assignment-letter.ts
import type { TableColumn } from '@nuxt/ui';
import { defineStore } from 'pinia'
import { type SuratTugas, type SuratTugasForm, type SuratTugasStatus, AuditCategory } from '~/types/audit'

export interface AssignmentLetterState {
  isModalOpen: boolean;
  assignmentLetterList: SuratTugas[];
  form: SuratTugasForm;
  columns: TableColumn<SuratTugas>[];
  options: {
    auditTeam: string[];
    workingUnit: string[];
    role: string[];
  };
}

export const useAssignmentLetterStore = defineStore('surat-tugas', {
  state: (): AssignmentLetterState => ({
    isModalOpen: false,
    assignmentLetterList: [
      {
        id: crypto.randomUUID(),
        letterNumber: 'ST-001/SKAI/2026',
        status: 'Published' as SuratTugasStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Real Auditore',
        leader: 'Zeta Ramadhani',
        category: AuditCategory.ASSURANCE,
        auditYear: new Date().getFullYear().toString(),
        auditTeam: 'SKAI',
        startPeriod: '2026-03-01',
        finishPeriod: '2026-03-31',
        workingUnit: 'Finance',
        executionPeriod: '2026-03-01 to 2026-03-31',
        membersList: [
          { name: 'Zeta Ramadhani',   role: 'Chairperson' },
          { name: 'Budi Santoso',     role: 'Supervisor' },
          { name: 'Rina Wulandari',   role: 'Member' },
          { name: 'Andi Firmansyah',  role: 'Member' },
          { name: 'Dewi Kusumawati',  role: 'Person in Charge' }
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
    form: {
      auditTitle: '',
      leader: '',
      category: AuditCategory.ASSURANCE,
      auditYear: new Date().getFullYear().toString(),
      auditTeam: 'SKAI',
      startPeriod: '',
      finishPeriod: '',
      workingUnit: '',
      membersList: [
        { name: '', role: 'Chairperson' },
        { name: '', role: 'Member' }
      ],
      purposeList: [''],
      scopeList: [''],
      ccList: ['President Director']
    },
    columns: [
      { accessorKey: 'letterNumber',    header: 'Letter Number' },
      { accessorKey: 'assignmentTitle', header: 'Audit Title / Object' },
      { accessorKey: 'workingUnit',     header: 'Work Unit' },
      { accessorKey: 'executionPeriod', header: 'Execution Period' },
      { accessorKey: 'auditTeam',       header: 'Audit Team' },
      { accessorKey: 'status',          header: 'Status' }
    ],
    options: {
      auditTeam: ['SKAI', 'DAI', 'CAE'],
      workingUnit: ['Production', 'Marketing', 'Finance'],
      role: ['Person in Charge', 'Supervisor', 'Chairperson', 'Member'],
    }
  }),

  // ==========================================
  // GETTERS
  // ==========================================
  getters: {
    getColumns:              (state) => state.columns,
    getAssignmentLetterList: (state) => state.assignmentLetterList,
    getForm:                 (state) => state.form,
    getOptions:              (state) => state.options,

    // dateError moved here as a getter — computed equivalent in Options API store
    dateError: (state): string | null => {
      if (state.form.startPeriod && state.form.finishPeriod) {
        const start = new Date(state.form.startPeriod)
        const end   = new Date(state.form.finishPeriod)
        if (end < start) {
          return "Error: End date cannot be before start date."
        }
      }
      return null
    }
  },
  actions: {
    // ── HELPER FORM FUNCTIONS ──
    addItem(list: any[], defaultItem: any) {
      list.push(typeof defaultItem === 'object' ? { ...defaultItem } : defaultItem)
    },

    removeItem(list: any[], index: number) {
      list.splice(index, 1)
    },

    // ── LETTER NUMBER GENERATOR ──
    generateNomorSurat(auditTeam: string, year: string): string {
      const nextCount  = this.assignmentLetterList.length + 1
      const paddedCount = nextCount.toString().padStart(3, '0')
      return `ST-${paddedCount}/${auditTeam}/${year}`
    },

    // ── MODAL CONTROLS ──
    openModal() {
      Object.assign(this.form, {
        auditTitle:  '',
        leader:      '',
        category:    AuditCategory.ASSURANCE,
        auditYear:   new Date().getFullYear().toString(),
        auditTeam:   'SKAI',
        startPeriod: '',
        finishPeriod:'',
        workingUnit: '',
        membersList: [{ name: '', role: 'Chairperson' }],
        purposeList: [''],
        scopeList:   [''],
        ccList:      ['President Director']
      })
      this.isModalOpen = true
    },

    closeModal() {
      this.isModalOpen = false
    },

    // ── SUBMIT ──
    handleSubmit() {
      // 1. Validate Work Unit
      if (!this.form.workingUnit) {
        alert("Work unit must be filled in.")
        return
      }

      // 2. Validate Dates
      if (this.dateError || !this.form.startPeriod || !this.form.finishPeriod) {
        alert(this.dateError || "Please fill in start and end period.")
        return
      }

      // 3. Validate Team Members (min 3)
      if (this.form.membersList.length < 3) {
        const proceed = confirm("Template suggests at least 3 team members. Continue saving?")
        if (!proceed) return
      }

      // 4. Build & save entry
      this.assignmentLetterList.unshift({
        id:              crypto.randomUUID(),
        letterNumber:    this.generateNomorSurat(this.form.auditTeam, this.form.auditYear),
        status:          'Draft',
        createdAt:       new Date().toISOString(),
        auditTitle:      this.form.auditTitle,
        leader:          this.form.leader,
        category:        this.form.category as AuditCategory,
        auditYear:       this.form.auditYear,
        auditTeam:       this.form.auditTeam,
        startPeriod:     this.form.startPeriod,
        finishPeriod:    this.form.finishPeriod,
        workingUnit:     this.form.workingUnit,
        executionPeriod: `${this.form.startPeriod} to ${this.form.finishPeriod}`,
        membersList:     JSON.parse(JSON.stringify(this.form.membersList)),
        purposeList:     JSON.parse(JSON.stringify(this.form.purposeList)),
        scopeList:       JSON.parse(JSON.stringify(this.form.scopeList)),
        ccList:          JSON.parse(JSON.stringify(this.form.ccList))
      })

      // 5. Close modal
      this.closeModal()
    },

    // ── CRUD ACTIONS ──
    addSuratTugas(form: SuratTugasForm) {
      const newEntry: SuratTugas = {
        ...form,
        id:              crypto.randomUUID(),
        letterNumber:    this.generateNomorSurat(form.auditTeam, form.auditYear),
        executionPeriod: `${form.startPeriod} to ${form.finishPeriod}`,
        status:          'Draft',
        createdAt:       new Date().toISOString()
      }
      this.assignmentLetterList.unshift(newEntry)
    },

    deleteSuratTugas(id: string) {
      this.assignmentLetterList = this.assignmentLetterList.filter(s => s.id !== id)
    },

    changeStatus(id: string, status: SuratTugasStatus) {
      const surat = this.assignmentLetterList.find(s => s.id === id)
      if (surat) {
        surat.status = status
      }
    }
  }
})