// stores/assignment-letter.ts
import type { TableColumn } from '@nuxt/ui';
import { defineStore } from 'pinia'
import { type AssignmentLetter, type AssignmentLetterForm, type AssignmentLetterStatus, AuditCategory } from '~/types/audit'

export interface AssignmentLetterState {
  isModalOpen: boolean;
  editingId: string | null;
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
    editingId: null,
    assignmentLetterList: [
      {
        id: 'mock-uuid-zeta-1',
        letterNumber: 'ST-001/SKAI/2026',
        status: 'Published' as AssignmentLetterStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Audit Pengendalian Keuangan & Pengeluaran Kas Q1 2026',
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
      },
      {
        id: 'mock-uuid-andi-2',
        letterNumber: 'ST-002/SKAI/2026',
        status: 'Published' as AssignmentLetterStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Audit Keamanan Sistem Informasi & Infrastruktur ERP 2026',
        leader: 'Andi Firmansyah',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-04-01',
        finishPeriod: '2026-04-30',
        workingUnit: 'IT',
        auditPurpose: 'IT Security Audit',
        letterDate: '2026-01-10',
        caeSignature: 'System',
        executionPeriod: '2026-04-01 to 2026-04-30',
        membersList: [
          { name: 'Andi Firmansyah', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' },
          { name: 'Dedi Prasetyo', role: 'Member' }
        ],
        purposeList: ['Evaluasi tata kelola akses pengguna dan keamanan database ERP'],
        scopeList: ['Hak akses user ERP dan database pusat'],
        ccList: ['Chief Information Officer', 'Head of Internal Audit']
      },
      {
        id: 'mock-uuid-rina-3',
        letterNumber: 'ST-003/SKAI/2026',
        status: 'Published' as AssignmentLetterStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Audit Operasional Branch, Gudang & Logistik 2026',
        leader: 'Rina Wulandari',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-07-01',
        finishPeriod: '2026-07-31',
        workingUnit: 'Operations',
        auditPurpose: 'Operational Audit',
        letterDate: '2026-01-25',
        caeSignature: 'System',
        executionPeriod: '2026-07-01 to 2026-07-31',
        membersList: [
          { name: 'Rina Wulandari', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' }
        ],
        purposeList: ['Memastikan pengelolaan persediaan dan stok gudang sesuai SOP'],
        scopeList: ['Opname stok persediaan gudang pusat'],
        ccList: ['Chief Operating Officer', 'Head of Internal Audit']
      },
      {
        id: 'mock-uuid-budi-4',
        letterNumber: 'ST-004/SKAI/2026',
        status: 'Published' as AssignmentLetterStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Audit Kepatuhan Manajemen Risiko & Pengadaan Barang/Jasa 2026',
        leader: 'Budi Santoso',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-08-01',
        finishPeriod: '2026-08-31',
        workingUnit: 'Procurement',
        auditPurpose: 'Compliance Audit',
        letterDate: '2026-02-01',
        caeSignature: 'System',
        executionPeriod: '2026-08-01 to 2026-08-31',
        membersList: [
          { name: 'Budi Santoso', role: 'Chairperson' },
          { name: 'Zeta Ramadhani', role: 'Supervisor' }
        ],
        purposeList: ['Evaluasi proses HPS dan tender SCM'],
        scopeList: ['Kontrak dan vendor management'],
        ccList: ['Head of Procurement', 'Head of Internal Audit']
      },
      {
        id: 'mock-uuid-dewi-5',
        letterNumber: 'ST-005/SKAI/2026',
        status: 'Published' as AssignmentLetterStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Audit Pengendalian K3LH & Manufaktur Pembangkit 2026',
        leader: 'Dewi Kusumawati',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-09-01',
        finishPeriod: '2026-09-30',
        workingUnit: 'Maintenance',
        auditPurpose: 'HSE Audit',
        letterDate: '2026-02-15',
        caeSignature: 'System',
        executionPeriod: '2026-09-01 to 2026-09-30',
        membersList: [
          { name: 'Dewi Kusumawati', role: 'Chairperson' },
          { name: 'Rina Wulandari', role: 'Member' }
        ],
        purposeList: ['Evaluasi implementasi K3LH dan sertifikasi alat'],
        scopeList: ['Inspeksi K3 dan fasilitas pemadam kebakaran'],
        ccList: ['Chief Safety Officer', 'Head of Internal Audit']
      },
      {
        id: 'mock-uuid-tomy-020',
        letterNumber: '020/ST/01/KSIAD/2023',
        status: 'Published' as AssignmentLetterStatus,
        createdAt: new Date().toISOString(),
        auditTitle: 'Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau',
        leader: 'Tomy Afrilianto',
        category: AuditCategory.ASSURANCE,
        auditYear: '2023',
        auditTeam: 'SKAI',
        startPeriod: '2023-01-01',
        finishPeriod: '2023-08-31',
        workingUnit: 'Operasi & Pemeliharaan',
        auditPurpose: 'Operational Audit',
        letterDate: '2023-01-05',
        caeSignature: 'System',
        executionPeriod: 'Januari 2023 s.d Agustus 2023',
        membersList: [
          { name: 'Tomy Afrilianto', role: 'Chairperson' },
          { name: 'Robert Sunarijanto', role: 'Supervisor' }
        ],
        purposeList: ['Evaluasi ketersediaan pembangkit, K3LH, dan manajemen risiko'],
        scopeList: ['Operasional UPDK Kepulauan Riau'],
        ccList: ['Head of SKAI', 'Board of Directors']
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
      { accessorKey: 'auditTitle', header: 'Audit Title / Object' },
      { accessorKey: 'workingUnit', header: 'Work Unit' },
      { accessorKey: 'executionPeriod', header: 'Execution Period' },
      { accessorKey: 'auditTeam', header: 'Audit Team' },
      { accessorKey: 'status', header: 'Status' },
      { accessorKey: 'actions', header: 'Actions' }
    ],
    options: {
      auditTeam: ['SKAI', 'DAI', 'CAE'],
      workingUnit: ['Production', 'Marketing', 'Finance', 'IT', 'Operations', 'Procurement', 'Maintenance'],
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
          auditTitle: 'Audit Pengendalian Keuangan & Pengeluaran Kas Q1 2026',
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
            { name: 'Rina Wulandari', role: 'Member' }
          ],
          purposeList: ['Menilai efektivitas pengendalian internal pada divisi keuangan'],
          scopeList: ['Proses pencatatan dan pelaporan keuangan'],
          ccList: ['President Director']
        },
        {
          id: 'mock-uuid-andi-2',
          letterNumber: 'ST-002/SKAI/2026',
          status: 'Published' as AssignmentLetterStatus,
          createdAt: new Date().toISOString(),
          auditTitle: 'Audit Keamanan Sistem Informasi & Infrastruktur ERP 2026',
          leader: 'Andi Firmansyah',
          category: AuditCategory.ASSURANCE,
          auditYear: '2026',
          auditTeam: 'SKAI',
          startPeriod: '2026-04-01',
          finishPeriod: '2026-04-30',
          workingUnit: 'IT',
          auditPurpose: 'IT Security Audit',
          letterDate: '2026-01-10',
          caeSignature: 'System',
          executionPeriod: '2026-04-01 to 2026-04-30',
          membersList: [
            { name: 'Andi Firmansyah', role: 'Chairperson' },
            { name: 'Budi Santoso', role: 'Supervisor' }
          ],
          purposeList: ['Evaluasi tata kelola akses pengguna dan keamanan database ERP'],
          scopeList: ['Hak akses user ERP dan database pusat'],
          ccList: ['Chief Information Officer']
        },
        {
          id: 'mock-uuid-rina-3',
          letterNumber: 'ST-003/SKAI/2026',
          status: 'Published' as AssignmentLetterStatus,
          createdAt: new Date().toISOString(),
          auditTitle: 'Audit Operasional Branch, Gudang & Logistik 2026',
          leader: 'Rina Wulandari',
          category: AuditCategory.ASSURANCE,
          auditYear: '2026',
          auditTeam: 'SKAI',
          startPeriod: '2026-07-01',
          finishPeriod: '2026-07-31',
          workingUnit: 'Operations',
          auditPurpose: 'Operational Audit',
          letterDate: '2026-01-25',
          caeSignature: 'System',
          executionPeriod: '2026-07-01 to 2026-07-31',
          membersList: [
            { name: 'Rina Wulandari', role: 'Chairperson' },
            { name: 'Budi Santoso', role: 'Supervisor' }
          ],
          purposeList: ['Memastikan pengelolaan persediaan dan stok gudang sesuai SOP'],
          scopeList: ['Opname stok persediaan gudang pusat'],
          ccList: ['Chief Operating Officer']
        },
        {
          id: 'mock-uuid-budi-4',
          letterNumber: 'ST-004/SKAI/2026',
          status: 'Published' as AssignmentLetterStatus,
          createdAt: new Date().toISOString(),
          auditTitle: 'Audit Kepatuhan Manajemen Risiko & Pengadaan Barang/Jasa 2026',
          leader: 'Budi Santoso',
          category: AuditCategory.ASSURANCE,
          auditYear: '2026',
          auditTeam: 'SKAI',
          startPeriod: '2026-08-01',
          finishPeriod: '2026-08-31',
          workingUnit: 'Procurement',
          auditPurpose: 'Compliance Audit',
          letterDate: '2026-02-01',
          caeSignature: 'System',
          executionPeriod: '2026-08-01 to 2026-08-31',
          membersList: [{ name: 'Budi Santoso', role: 'Chairperson' }],
          purposeList: ['Evaluasi proses HPS dan tender SCM'],
          scopeList: ['Kontrak dan vendor management'],
          ccList: ['Head of Procurement']
        },
        {
          id: 'mock-uuid-dewi-5',
          letterNumber: 'ST-005/SKAI/2026',
          status: 'Published' as AssignmentLetterStatus,
          createdAt: new Date().toISOString(),
          auditTitle: 'Audit Pengendalian K3LH & Manufaktur Pembangkit 2026',
          leader: 'Dewi Kusumawati',
          category: AuditCategory.ASSURANCE,
          auditYear: '2026',
          auditTeam: 'SKAI',
          startPeriod: '2026-09-01',
          finishPeriod: '2026-09-30',
          workingUnit: 'Maintenance',
          auditPurpose: 'HSE Audit',
          letterDate: '2026-02-15',
          caeSignature: 'System',
          executionPeriod: '2026-09-01 to 2026-09-30',
          membersList: [{ name: 'Dewi Kusumawati', role: 'Chairperson' }],
          purposeList: ['Evaluasi K3LH'],
          scopeList: ['Inspeksi fasilitas K3'],
          ccList: ['Chief Safety Officer']
        },
        {
          id: 'mock-uuid-tomy-020',
          letterNumber: '020/ST/01/KSIAD/2023',
          status: 'Published' as AssignmentLetterStatus,
          createdAt: new Date().toISOString(),
          auditTitle: 'Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau',
          leader: 'Tomy Afrilianto',
          category: AuditCategory.ASSURANCE,
          auditYear: '2023',
          auditTeam: 'SKAI',
          startPeriod: '2023-01-01',
          finishPeriod: '2023-08-31',
          workingUnit: 'Operasi & Pemeliharaan',
          auditPurpose: 'Operational Audit',
          letterDate: '2023-01-05',
          caeSignature: 'System',
          executionPeriod: 'Januari 2023 s.d Agustus 2023',
          membersList: [{ name: 'Tomy Afrilianto', role: 'Chairperson' }],
          purposeList: ['Evaluasi ketersediaan pembangkit'],
          scopeList: ['Operasional UPDK Kepulauan Riau'],
          ccList: ['Head of SKAI']
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
      this.editingId = null
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

    openEditModal(letter: AssignmentLetter) {
      this.editingId = letter.id

      Object.assign(this.form, {
        auditTitle: letter.auditTitle,
        leader: letter.leader,
        category: letter.category,
        auditYear: letter.auditYear,
        auditTeam: letter.auditTeam,
        startPeriod: letter.startPeriod,
        finishPeriod: letter.finishPeriod,
        workingUnit: letter.workingUnit,
        auditPurpose: letter.auditPurpose,
        letterDate: letter.letterDate
          ? String(letter.letterDate).slice(0, 10)
          : '',
        caeSignature: letter.caeSignature || '',
        membersList: letter.membersList.map(member => ({ ...member })),
        purposeList: [...letter.purposeList],
        scopeList: [...letter.scopeList],
        ccList: [...letter.ccList]
      })

      this.isModalOpen = true
    },

    closeModal() {
      this.isModalOpen = false
      this.editingId = null
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
        const executionPeriod =
          `${this.form.startPeriod} to ${this.form.finishPeriod}`

        if (this.editingId) {
          const existingLetter = this.assignmentLetterList.find(
            letter => letter.id === this.editingId
          )

          if (!existingLetter) {
            throw new Error('Assignment letter not found.')
          }

          const payload = {
            ...this.form,
            letterNumber: existingLetter.letterNumber,
            status: existingLetter.status,
            executionPeriod
          }

          await $fetch(
            `${baseUrl}/assignment-letters/${this.editingId}`,
            {
              method: 'PUT',
              body: payload
            }
          )
        } else {
          const letterNumber = this.generateNomorSurat(
            this.form.auditTeam,
            this.form.auditYear
          )

          const payload = {
            ...this.form,
            letterNumber,
            status: 'Draft',
            executionPeriod
          }

          await $fetch(`${baseUrl}/assignment-letters`, {
            method: 'POST',
            body: payload
          })
        }

        this.closeModal()
        await this.fetchAssignmentLetters()
      } catch (error: any) {
        console.error(
          this.editingId
            ? 'Failed to update assignment letter:'
            : 'Failed to create assignment letter:',
          error
        )

        alert(
          this.editingId
            ? 'Failed to update assignment letter.'
            : 'Failed to save assignment letter.'
        )
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