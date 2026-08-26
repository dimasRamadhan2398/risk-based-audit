import { defineStore } from 'pinia'
import { reactive, ref, computed, watch } from 'vue'
import { useAssignmentLetterStore } from './assignment-letter'

export interface InterviewItem {
  id: any
  assignmentLetterId: string
  interviewee: string
  intervieweePosition: string
  interviewer: string
  interviewerPosition: string
  date: string
  topic: string
  file: File | null
  fileName?: string
}

export interface ObservationItem {
  id: any
  assignmentLetterId: string
  activity: string
  location: string
  date: string
  observer: string
  file: File | null
  fileName?: string
}

export interface DocumentItem {
  id: any
  assignmentLetterId: string
  documentName: string
  description: string
  requiredDate: string
  file: File | null
  fileName?: string
}

export interface SampleItem {
  id: any
  assignmentLetterId: string
  documentName: string
  documentNumber: string
  date: string
  description: string
}

export interface TestControlItem {
  id: any
  assignmentLetterId: string
  controlName: string
  controlDescription: string
  controlType: string
  testProcedure: string
  testResult: string
  finding: string
  recommendation: string
  mitigationPlan: string
  pic: string
  dueDate: string
}

export const useAuditFieldworkStore = defineStore('audit-fieldwork', () => {
  const assignmentLetterStore = useAssignmentLetterStore()
  const selectedAssignmentLetter = ref<string>('')
  const loading = ref(false)
  const errorMsg = ref('')

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const publishedAssignmentLetters = computed(() => {
    return assignmentLetterStore.assignmentLetterList
      .filter((st: any) => st.status === 'Published')
      .map((st: any) => st.letterNumber)
  })

  const hasSelectedAssignmentLetter = computed(() => !!selectedAssignmentLetter.value && selectedAssignmentLetter.value !== '')

  const selectedAssignmentLetterData = computed(() => {
    return assignmentLetterStore.assignmentLetterList.find(
      (st: any) => st.letterNumber === selectedAssignmentLetter.value
    )
  })

  const memberOptions = computed(() => {
    const data = selectedAssignmentLetterData.value
    if (!data?.membersList) return []
    return data.membersList
      .filter((m: any) => m.name && m.name.trim() !== '')
      .map((m: any) => ({
        label: m.name,
        value: m.name,
        role: m.role
      }))
  })

  const tabs = [
    { label: 'Interview', slot: 'tab01', icon: 'i-heroicons-microphone' },
    { label: 'Observation', slot: 'tab02', icon: 'i-heroicons-eye' },
    { label: 'Document Collection', slot: 'tab03', icon: 'i-heroicons-document-duplicate' },
    { label: 'Sample Data', slot: 'tab04', icon: 'i-heroicons-table-cells' },
    { label: 'Test Controls', slot: 'tab05', icon: 'i-heroicons-shield-check' },
    { label: 'Working Papers', slot: 'tab06', icon: 'i-heroicons-document-text' }
  ]

  const options = {
    positions: ['Manager', 'Supervisor', 'Staff', 'Director', 'Head of Department', 'Other'],
    controlTypes: ['Preventive', 'Detective', 'Corrective', 'Manual', 'Automated'],
    testResults: ['Effective', 'Ineffective', 'Partially Effective', 'Not Tested'],
    auditTopics: ['Internal Control', 'Compliance', 'Risk Management', 'Process Efficiency', 'Other']
  }

  const mockFieldwork: Record<string, {
    interviews: InterviewItem[]
    observations: ObservationItem[]
    documents: DocumentItem[]
    samples: SampleItem[]
    testControls: TestControlItem[]
  }> = {
    'ST-001/SKAI/2026': {
      interviews: [
        {
          id: '1',
          assignmentLetterId: 'ST-001/SKAI/2026',
          interviewee: 'Ahmad Yani',
          intervieweePosition: 'Head of Finance',
          interviewer: 'Zeta Ramadhani',
          interviewerPosition: 'Chairperson',
          date: '2026-03-05',
          topic: 'Internal Control over Financial Reporting (ICOFR) implementation, segregation of duties in payment approvals, and monthly bank reconciliation process.',
          file: null,
          fileName: 'Meeting_Minutes_ICOFR.pdf'
        },
        {
          id: '2',
          assignmentLetterId: 'ST-001/SKAI/2026',
          interviewee: 'Rudi Hermawan',
          intervieweePosition: 'IT Manager',
          interviewer: 'Andi Firmansyah',
          interviewerPosition: 'Member',
          date: '2026-03-08',
          topic: 'Access controls to the ERP system, user privilege review procedures, and backup recovery tests.',
          file: null,
          fileName: 'IT_Access_Controls_Interview.pdf'
        }
      ],
      observations: [
        {
          id: '1',
          assignmentLetterId: 'ST-001/SKAI/2026',
          activity: 'Observe cash count process in main vault and verify safety box lock code authorization.',
          location: 'Headquarters - Vault Room',
          date: '2026-03-10',
          observer: 'Rina Wulandari',
          file: null,
          fileName: 'Cash_Count_Observation.pdf'
        },
        {
          id: '2',
          assignmentLetterId: 'ST-001/SKAI/2026',
          activity: 'Walkthrough observation of transaction recording inside ERP and checking automated system logs.',
          location: 'IT Server Room & Finance Desk',
          date: '2026-03-12',
          observer: 'Andi Firmansyah',
          file: null
        }
      ],
      documents: [
        {
          id: '1',
          assignmentLetterId: 'ST-001/SKAI/2026',
          documentName: 'General Ledger 2025',
          description: 'Full year accounting ledger containing all transactions for balance sheet validation.',
          requiredDate: '2026-03-02',
          file: null,
          fileName: 'General_Ledger_2025_Final.xlsx'
        },
        {
          id: '2',
          assignmentLetterId: 'ST-001/SKAI/2026',
          documentName: 'ERP System Access Logs',
          description: 'Active user list and permission matrix for segregation of duties audit.',
          requiredDate: '2026-03-05',
          file: null,
          fileName: 'ERP_Access_Logs_Q4_2025.csv'
        }
      ],
      samples: [
        {
          id: '1',
          assignmentLetterId: 'ST-001/SKAI/2026',
          documentName: 'Procurement Invoice',
          documentNumber: 'INV-2025-0988',
          date: '2025-11-12',
          description: 'Sample transaction for validation of purchase order matching and invoice approval.'
        },
        {
          id: '2',
          assignmentLetterId: 'ST-001/SKAI/2026',
          documentName: 'Bank Statement Reconciliation',
          documentNumber: 'BR-2025-12',
          date: '2025-12-31',
          description: 'Reconciliation report for main bank account verifying outstanding checks and deposits.'
        }
      ],
      testControls: [
        {
          id: '1',
          assignmentLetterId: 'ST-001/SKAI/2026',
          controlName: 'Payment Authorization Limit',
          controlDescription: 'All payments above Rp 50,000,000 must be approved by the Finance Director.',
          controlType: 'Preventive',
          testProcedure: 'Select a sample of payments > Rp 50m and verify presence of Director signature or digital approval.',
          testResult: 'Effective',
          finding: 'No deviations found. All sample vouchers had the required dual approval.',
          recommendation: 'Maintain current process and ensure limits are updated in system configuration.',
          mitigationPlan: 'Routine system configuration audit.',
          pic: 'Ahmad Yani (Head of Finance)',
          dueDate: '2026-06-30'
        },
        {
          id: '2',
          assignmentLetterId: 'ST-001/SKAI/2026',
          controlName: 'IT Backup Daily Execution',
          controlDescription: 'ERP database backups must run automatically every night at 23:00.',
          controlType: 'Automated',
          testProcedure: 'Review backup logs for the month of December 2025 and verify successful backup statuses.',
          testResult: 'Ineffective',
          finding: 'On 3 days (Dec 12, 13, and 18), backups failed due to disk space issues. No alert was sent.',
          recommendation: 'Implement automated notification alerting IT team when backup status is failed.',
          mitigationPlan: 'Setup automated SMTP notification inside backup script.',
          pic: 'Rudi Hermawan (IT Manager)',
          dueDate: '2026-04-15'
        }
      ]
    },
    'ST-002/SKAI/2026': {
      interviews: [
        {
          id: '101',
          assignmentLetterId: 'ST-002/SKAI/2026',
          interviewee: 'Bambang Susilo',
          intervieweePosition: 'Head of IT',
          interviewer: 'Andi Firmansyah',
          interviewerPosition: 'Chairperson',
          date: '2026-04-05',
          topic: 'Keamanan Database ERP & Multi-Factor Authentication enforcement.',
          file: null,
          fileName: 'Interview_IT_Security_ST002.pdf'
        }
      ],
      observations: [
        {
          id: '101',
          assignmentLetterId: 'ST-002/SKAI/2026',
          activity: 'Pengujian Disaster Recovery & Backup Restore Data ERP.',
          location: 'Data Center Utama',
          date: '2026-04-12',
          observer: 'Andi Firmansyah',
          file: null
        }
      ],
      documents: [
        {
          id: '101',
          assignmentLetterId: 'ST-002/SKAI/2026',
          documentName: 'ERP Security Matrix Log',
          description: 'Log autentikasi dan daftar hak akses pengguna ERP Q1 2026.',
          requiredDate: '2026-04-02',
          file: null,
          fileName: 'ERP_Security_Matrix.xlsx'
        }
      ],
      samples: [
        {
          id: '101',
          assignmentLetterId: 'ST-002/SKAI/2026',
          documentName: 'Log Akses User Admin',
          documentNumber: 'LOG-ERP-2026-04',
          date: '2026-04-10',
          description: 'Sample log akses superadmin database ERP.'
        }
      ],
      testControls: [
        {
          id: '101',
          assignmentLetterId: 'ST-002/SKAI/2026',
          controlName: 'MFA Enforcement',
          controlDescription: 'MFA wajib untuk superadmin ERP.',
          controlType: 'Automated',
          testProcedure: 'Verify TOTP configuration for admin accounts.',
          testResult: 'Ineffective',
          finding: 'Single factor login active for 2 admin accounts.',
          recommendation: 'Enforce mandatory MFA.',
          mitigationPlan: 'Activate TOTP policy.',
          pic: 'Bambang Susilo (Head of IT)',
          dueDate: '2026-05-31'
        }
      ]
    },
    'ST-003/SKAI/2026': {
      interviews: [
        {
          id: '201',
          assignmentLetterId: 'ST-003/SKAI/2026',
          interviewee: 'Hendra Wijaya',
          intervieweePosition: 'Warehouse Manager',
          interviewer: 'Rina Wulandari',
          interviewerPosition: 'Chairperson',
          date: '2026-07-08',
          topic: 'Manajemen Stok Persediaan Gudang & Prosedur Opname Fisik.',
          file: null,
          fileName: 'Interview_Warehouse_ST003.pdf'
        }
      ],
      observations: [
        {
          id: '201',
          assignmentLetterId: 'ST-003/SKAI/2026',
          activity: 'Inspeksi Fisik dan Pengukuran Suhu Gudang Bahan Kimia.',
          location: 'Gudang Logistik Branch A',
          date: '2026-07-15',
          observer: 'Rina Wulandari',
          file: null
        }
      ],
      documents: [
        {
          id: '201',
          assignmentLetterId: 'ST-003/SKAI/2026',
          documentName: 'Kartu Stok & Laporan Opname',
          description: 'Berita acara stok opname barang semester 1 2026.',
          requiredDate: '2026-07-03',
          file: null,
          fileName: 'Stok_Opname_S1_2026.pdf'
        }
      ],
      samples: [
        {
          id: '201',
          assignmentLetterId: 'ST-003/SKAI/2026',
          documentName: 'Work Order Pengeluaran Barang',
          documentNumber: 'WO-LOG-2026-89',
          date: '2026-07-20',
          description: 'Sample otorisasi pengeluaran barang persediaan.'
        }
      ],
      testControls: [
        {
          id: '201',
          assignmentLetterId: 'ST-003/SKAI/2026',
          controlName: 'Warehouse Temperature Monitor',
          controlDescription: 'Suhu gudang terpantau 24/7.',
          controlType: 'Preventive',
          testProcedure: 'Inspect temp sensor logbook.',
          testResult: 'Ineffective',
          finding: 'Manual logging delayed by 2 days.',
          recommendation: 'Install IoT digital sensor.',
          mitigationPlan: 'Procure IoT sensors.',
          pic: 'Hendra Wijaya (Warehouse Manager)',
          dueDate: '2026-08-31'
        }
      ]
    }
  }

  const fieldworkData = ref<Record<string, {
    interviews: InterviewItem[]
    observations: ObservationItem[]
    documents: DocumentItem[]
    samples: SampleItem[]
    testControls: TestControlItem[]
  }>>({})

  const ensureDataExists = () => {
    if (!selectedAssignmentLetter.value) return
    if (!fieldworkData.value[selectedAssignmentLetter.value]) {
      if (mockFieldwork[selectedAssignmentLetter.value]) {
        fieldworkData.value[selectedAssignmentLetter.value] = JSON.parse(
          JSON.stringify(mockFieldwork[selectedAssignmentLetter.value])
        )
      } else {
        fieldworkData.value[selectedAssignmentLetter.value] = {
          interviews: [],
          observations: [],
          documents: [],
          samples: [],
          testControls: []
        }
      }
    }
  }

  const interviews = computed(() => {
    if (!selectedAssignmentLetter.value) return []
    return fieldworkData.value[selectedAssignmentLetter.value]?.interviews || []
  })

  const observations = computed(() => {
    if (!selectedAssignmentLetter.value) return []
    return fieldworkData.value[selectedAssignmentLetter.value]?.observations || []
  })

  const documents = computed(() => {
    if (!selectedAssignmentLetter.value) return []
    return fieldworkData.value[selectedAssignmentLetter.value]?.documents || []
  })

  const samples = computed(() => {
    if (!selectedAssignmentLetter.value) return []
    return fieldworkData.value[selectedAssignmentLetter.value]?.samples || []
  })

  const testControls = computed(() => {
    if (!selectedAssignmentLetter.value) return []
    return fieldworkData.value[selectedAssignmentLetter.value]?.testControls || []
  })

  const fetchAllFieldworkData = async (assignmentLetterId: string) => {
    if (!assignmentLetterId) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const [interviewsRes, observationsRes, documentsRes, samplesRes, testControlsRes]: any = await Promise.all([
        $fetch(`${baseUrl}/fieldwork/interviews?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/observations?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/documents?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/samples?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/test-controls?assignmentLetterId=${assignmentLetterId}`)
      ])

      const interviewsList = interviewsRes?.data?.items || interviewsRes?.items || (Array.isArray(interviewsRes) ? interviewsRes : [])
      const observationsList = observationsRes?.data?.items || observationsRes?.items || (Array.isArray(observationsRes) ? observationsRes : [])
      const documentsList = documentsRes?.data?.items || documentsRes?.items || (Array.isArray(documentsRes) ? documentsRes : [])
      const samplesList = samplesRes?.data?.items || samplesRes?.items || (Array.isArray(samplesRes) ? samplesRes : [])
      const testControlsList = testControlsRes?.data?.items || testControlsRes?.items || (Array.isArray(testControlsRes) ? testControlsRes : [])

      fieldworkData.value[assignmentLetterId] = {
        interviews: Array.isArray(interviewsList) && interviewsList.length > 0 ? interviewsList : (mockFieldwork[assignmentLetterId]?.interviews || []),
        observations: Array.isArray(observationsList) && observationsList.length > 0 ? observationsList : (mockFieldwork[assignmentLetterId]?.observations || []),
        documents: Array.isArray(documentsList) && documentsList.length > 0 ? documentsList : (mockFieldwork[assignmentLetterId]?.documents || []),
        samples: Array.isArray(samplesList) && samplesList.length > 0 ? samplesList : (mockFieldwork[assignmentLetterId]?.samples || []),
        testControls: Array.isArray(testControlsList) && testControlsList.length > 0 ? testControlsList : (mockFieldwork[assignmentLetterId]?.testControls || [])
      }
    } catch (error: any) {
      console.error('Failed to fetch fieldwork data:', error)
      errorMsg.value = 'Failed to load fieldwork data.'
      if (mockFieldwork[assignmentLetterId]) {
        fieldworkData.value[assignmentLetterId] = JSON.parse(
          JSON.stringify(mockFieldwork[assignmentLetterId])
        )
      }
    } finally {
      loading.value = false
    }
  }

  const ensureFieldworkDataHolder = (assignmentLetterId: string) => {
    if (!fieldworkData.value[assignmentLetterId]) {
      fieldworkData.value[assignmentLetterId] = {
        interviews: mockFieldwork[assignmentLetterId]?.interviews || [],
        observations: mockFieldwork[assignmentLetterId]?.observations || [],
        documents: mockFieldwork[assignmentLetterId]?.documents || [],
        samples: mockFieldwork[assignmentLetterId]?.samples || [],
        testControls: mockFieldwork[assignmentLetterId]?.testControls || []
      }
    }
  }

  const fetchInterviews = async (assignmentLetterId: string) => {
    if (!assignmentLetterId) return
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res: any = await $fetch(`${baseUrl}/fieldwork/interviews?assignmentLetterId=${assignmentLetterId}`)
      const list = res?.data?.items || res?.items || (Array.isArray(res) ? res : null)
      ensureFieldworkDataHolder(assignmentLetterId)
      if (Array.isArray(list) && list.length > 0) {
        fieldworkData.value[assignmentLetterId].interviews = list
      }
    } catch (error) {
      console.error('Failed to fetch interviews:', error)
    }
  }

  const fetchObservations = async (assignmentLetterId: string) => {
    if (!assignmentLetterId) return
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res: any = await $fetch(`${baseUrl}/fieldwork/observations?assignmentLetterId=${assignmentLetterId}`)
      const list = res?.data?.items || res?.items || (Array.isArray(res) ? res : null)
      ensureFieldworkDataHolder(assignmentLetterId)
      if (Array.isArray(list) && list.length > 0) {
        fieldworkData.value[assignmentLetterId].observations = list
      }
    } catch (error) {
      console.error('Failed to fetch observations:', error)
    }
  }

  const fetchDocuments = async (assignmentLetterId: string) => {
    if (!assignmentLetterId) return
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res: any = await $fetch(`${baseUrl}/fieldwork/documents?assignmentLetterId=${assignmentLetterId}`)
      const list = res?.data?.items || res?.items || (Array.isArray(res) ? res : null)
      ensureFieldworkDataHolder(assignmentLetterId)
      if (Array.isArray(list) && list.length > 0) {
        fieldworkData.value[assignmentLetterId].documents = list
      }
    } catch (error) {
      console.error('Failed to fetch documents:', error)
    }
  }

  const fetchSamples = async (assignmentLetterId: string) => {
    if (!assignmentLetterId) return
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res: any = await $fetch(`${baseUrl}/fieldwork/samples?assignmentLetterId=${assignmentLetterId}`)
      const list = res?.data?.items || res?.items || (Array.isArray(res) ? res : null)
      ensureFieldworkDataHolder(assignmentLetterId)
      if (Array.isArray(list) && list.length > 0) {
        fieldworkData.value[assignmentLetterId].samples = list
      }
    } catch (error) {
      console.error('Failed to fetch samples:', error)
    }
  }

  const fetchTestControls = async (assignmentLetterId: string) => {
    if (!assignmentLetterId) return
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res: any = await $fetch(`${baseUrl}/fieldwork/test-controls?assignmentLetterId=${assignmentLetterId}`)
      const list = res?.data?.items || res?.items || (Array.isArray(res) ? res : null)
      ensureFieldworkDataHolder(assignmentLetterId)
      if (Array.isArray(list) && list.length > 0) {
        fieldworkData.value[assignmentLetterId].testControls = list
      }
    } catch (error) {
      console.error('Failed to fetch test controls:', error)
    }
  }

  watch(selectedAssignmentLetter, (newVal) => {
    if (newVal) {
      fetchAllFieldworkData(newVal)
    }
  })

  // Interview Form
  const interviewForm = reactive({
    interviewee: '',
    intervieweePosition: '',
    interviewer: '',
    interviewerPosition: '',
    date: '',
    topic: '',
    file: null as File | null
  })
  const showInterviewModal = ref(false)
  const isEditingInterview = ref(false)
  const isReadOnlyInterview = ref(false)
  const editingInterviewId = ref<any>(null)

  const openInterviewModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetInterviewForm()
    isEditingInterview.value = false
    isReadOnlyInterview.value = false
    editingInterviewId.value = null
    showInterviewModal.value = true
  }

  const editInterview = (item: InterviewItem) => {
    isEditingInterview.value = true
    isReadOnlyInterview.value = false
    editingInterviewId.value = item.id
    interviewForm.interviewee = item.interviewee
    interviewForm.intervieweePosition = item.intervieweePosition
    interviewForm.interviewer = item.interviewer
    interviewForm.interviewerPosition = item.interviewerPosition
    interviewForm.date = item.date
    interviewForm.topic = item.topic
    interviewForm.file = item.file
    showInterviewModal.value = true
  }

  const viewInterview = (item: InterviewItem) => {
    editInterview(item)
    isEditingInterview.value = false
    isReadOnlyInterview.value = true
  }

  const resetInterviewForm = () => {
    interviewForm.interviewee = ''
    interviewForm.intervieweePosition = ''
    interviewForm.interviewer = ''
    interviewForm.interviewerPosition = ''
    interviewForm.date = ''
    interviewForm.topic = ''
    interviewForm.file = null
  }

  const handleInterviewFileChange = (e: Event) => {
    const target = e.target as HTMLInputElement
    const file = target.files?.[0]
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        alert('File too large! Maximum 10MB.')
        return
      }
      interviewForm.file = file
    }
  }

  const saveInterview = async () => {
    if (!selectedAssignmentLetter.value || isReadOnlyInterview.value) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = {
        ...interviewForm,
        assignmentLetterId: selectedAssignmentLetter.value,
        fileName: interviewForm.file ? interviewForm.file.name : ''
      }
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].interviews

      let res: any = null
      if (isEditingInterview.value && editingInterviewId.value) {
        res = await $fetch(`${baseUrl}/fieldwork/interviews/${editingInterviewId.value}`, {
          method: 'PUT',
          body: payload
        })
        const updatedItem = res?.data || res
        const idx = currentList.findIndex((item: any) => item.id === editingInterviewId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...payload,
            id: updatedItem?.id || editingInterviewId.value,
            file: interviewForm.file
          }
        }
      } else {
        res = await $fetch(`${baseUrl}/fieldwork/interviews`, {
          method: 'POST',
          body: payload
        })
        const createdItem = res?.data || res
        const newItem: InterviewItem = {
          id: createdItem?.id || Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          interviewee: interviewForm.interviewee,
          intervieweePosition: interviewForm.intervieweePosition,
          interviewer: interviewForm.interviewer,
          interviewerPosition: interviewForm.interviewerPosition,
          date: interviewForm.date,
          topic: interviewForm.topic,
          file: interviewForm.file,
          fileName: interviewForm.file ? interviewForm.file.name : ''
        }
        currentList.push(newItem)
      }
      showInterviewModal.value = false
      resetInterviewForm()
      await fetchInterviews(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API save error, applying local state update:', error)
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].interviews
      if (isEditingInterview.value && editingInterviewId.value) {
        const idx = currentList.findIndex((item: any) => item.id === editingInterviewId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...interviewForm,
            fileName: interviewForm.file ? interviewForm.file.name : currentList[idx].fileName
          }
        }
      } else {
        const newItem: InterviewItem = {
          id: Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          interviewee: interviewForm.interviewee,
          intervieweePosition: interviewForm.intervieweePosition,
          interviewer: interviewForm.interviewer,
          interviewerPosition: interviewForm.interviewerPosition,
          date: interviewForm.date,
          topic: interviewForm.topic,
          file: interviewForm.file,
          fileName: interviewForm.file ? interviewForm.file.name : ''
        }
        currentList.push(newItem)
      }
      showInterviewModal.value = false
      resetInterviewForm()
    } finally {
      loading.value = false
    }
  }

  const deleteInterview = async (index: number) => {
    if (!selectedAssignmentLetter.value) return
    const item = interviews.value[index]
    if (!item || !confirm('Are you sure you want to delete this interview?')) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/fieldwork/interviews/${item.id}`, {
        method: 'DELETE'
      })
      if (fieldworkData.value[selectedAssignmentLetter.value]?.interviews) {
        fieldworkData.value[selectedAssignmentLetter.value].interviews.splice(index, 1)
      }
      await fetchInterviews(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API delete error, applying local delete:', error)
      if (fieldworkData.value[selectedAssignmentLetter.value]?.interviews) {
        fieldworkData.value[selectedAssignmentLetter.value].interviews.splice(index, 1)
      }
    } finally {
      loading.value = false
    }
  }

  // Observation Form
  const observationForm = reactive({
    activity: '',
    location: '',
    date: '',
    observer: '',
    file: null as File | null
  })
  const showObservationModal = ref(false)
  const isEditingObservation = ref(false)
  const isReadOnlyObservation = ref(false)
  const editingObservationId = ref<any>(null)

  const openObservationModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetObservationForm()
    isEditingObservation.value = false
    isReadOnlyObservation.value = false
    editingObservationId.value = null
    showObservationModal.value = true
  }

  const editObservation = (item: ObservationItem) => {
    isEditingObservation.value = true
    isReadOnlyObservation.value = false
    editingObservationId.value = item.id
    observationForm.activity = item.activity
    observationForm.location = item.location
    observationForm.date = item.date
    observationForm.observer = item.observer
    observationForm.file = item.file
    showObservationModal.value = true
  }

  const viewObservation = (item: ObservationItem) => {
    editObservation(item)
    isEditingObservation.value = false
    isReadOnlyObservation.value = true
  }

  const resetObservationForm = () => {
    observationForm.activity = ''
    observationForm.location = ''
    observationForm.date = ''
    observationForm.observer = ''
    observationForm.file = null
  }

  const handleObservationFileChange = (e: Event) => {
    const target = e.target as HTMLInputElement
    const file = target.files?.[0]
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        alert('File too large! Maximum 10MB.')
        return
      }
      observationForm.file = file
    }
  }

  const saveObservation = async () => {
    if (!selectedAssignmentLetter.value || isReadOnlyObservation.value) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = {
        ...observationForm,
        assignmentLetterId: selectedAssignmentLetter.value,
        fileName: observationForm.file ? observationForm.file.name : ''
      }
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].observations

      let res: any = null
      if (isEditingObservation.value && editingObservationId.value) {
        res = await $fetch(`${baseUrl}/fieldwork/observations/${editingObservationId.value}`, {
          method: 'PUT',
          body: payload
        })
        const updatedItem = res?.data || res
        const idx = currentList.findIndex((item: any) => item.id === editingObservationId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...payload,
            id: updatedItem?.id || editingObservationId.value,
            file: observationForm.file
          }
        }
      } else {
        res = await $fetch(`${baseUrl}/fieldwork/observations`, {
          method: 'POST',
          body: payload
        })
        const createdItem = res?.data || res
        const newItem: ObservationItem = {
          id: createdItem?.id || Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          activity: observationForm.activity,
          location: observationForm.location,
          date: observationForm.date,
          observer: observationForm.observer,
          file: observationForm.file,
          fileName: observationForm.file ? observationForm.file.name : ''
        }
        currentList.push(newItem)
      }
      showObservationModal.value = false
      resetObservationForm()
      await fetchObservations(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API save error, applying local state update:', error)
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].observations
      if (isEditingObservation.value && editingObservationId.value) {
        const idx = currentList.findIndex((item: any) => item.id === editingObservationId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...observationForm,
            fileName: observationForm.file ? observationForm.file.name : currentList[idx].fileName
          }
        }
      } else {
        const newItem: ObservationItem = {
          id: Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          activity: observationForm.activity,
          location: observationForm.location,
          date: observationForm.date,
          observer: observationForm.observer,
          file: observationForm.file,
          fileName: observationForm.file ? observationForm.file.name : ''
        }
        currentList.push(newItem)
      }
      showObservationModal.value = false
      resetObservationForm()
    } finally {
      loading.value = false
    }
  }

  const deleteObservation = async (index: number) => {
    if (!selectedAssignmentLetter.value) return
    const item = observations.value[index]
    if (!item || !confirm('Are you sure you want to delete this observation?')) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/fieldwork/observations/${item.id}`, {
        method: 'DELETE'
      })
      if (fieldworkData.value[selectedAssignmentLetter.value]?.observations) {
        fieldworkData.value[selectedAssignmentLetter.value].observations.splice(index, 1)
      }
      await fetchObservations(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API delete error, applying local delete:', error)
      if (fieldworkData.value[selectedAssignmentLetter.value]?.observations) {
        fieldworkData.value[selectedAssignmentLetter.value].observations.splice(index, 1)
      }
    } finally {
      loading.value = false
    }
  }

  // Document Form
  const documentForm = reactive({
    documentName: '',
    description: '',
    requiredDate: '',
    file: null as File | null
  })
  const showDocumentModal = ref(false)
  const isEditingDocument = ref(false)
  const isReadOnlyDocument = ref(false)
  const editingDocumentId = ref<any>(null)

  const openDocumentModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetDocumentForm()
    isEditingDocument.value = false
    isReadOnlyDocument.value = false
    editingDocumentId.value = null
    showDocumentModal.value = true
  }

  const editDocument = (item: DocumentItem) => {
    isEditingDocument.value = true
    isReadOnlyDocument.value = false
    editingDocumentId.value = item.id
    documentForm.documentName = item.documentName
    documentForm.description = item.description
    documentForm.requiredDate = item.requiredDate
    documentForm.file = item.file
    showDocumentModal.value = true
  }

  const viewDocument = (item: DocumentItem) => {
    editDocument(item)
    isEditingDocument.value = false
    isReadOnlyDocument.value = true
  }

  const resetDocumentForm = () => {
    documentForm.documentName = ''
    documentForm.description = ''
    documentForm.requiredDate = ''
    documentForm.file = null
  }

  const handleDocumentFileChange = (e: Event) => {
    const target = e.target as HTMLInputElement
    const file = target.files?.[0]
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        alert('File too large! Maximum 10MB.')
        return
      }
      documentForm.file = file
    }
  }

  const saveDocument = async () => {
    if (!selectedAssignmentLetter.value || isReadOnlyDocument.value) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = {
        ...documentForm,
        assignmentLetterId: selectedAssignmentLetter.value,
        fileName: documentForm.file ? documentForm.file.name : ''
      }
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].documents

      let res: any = null
      if (isEditingDocument.value && editingDocumentId.value) {
        res = await $fetch(`${baseUrl}/fieldwork/documents/${editingDocumentId.value}`, {
          method: 'PUT',
          body: payload
        })
        const updatedItem = res?.data || res
        const idx = currentList.findIndex((item: any) => item.id === editingDocumentId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...payload,
            id: updatedItem?.id || editingDocumentId.value,
            file: documentForm.file
          }
        }
      } else {
        res = await $fetch(`${baseUrl}/fieldwork/documents`, {
          method: 'POST',
          body: payload
        })
        const createdItem = res?.data || res
        const newItem: DocumentItem = {
          id: createdItem?.id || Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          documentName: documentForm.documentName,
          description: documentForm.description,
          requiredDate: documentForm.requiredDate,
          file: documentForm.file,
          fileName: documentForm.file ? documentForm.file.name : ''
        }
        currentList.push(newItem)
      }
      showDocumentModal.value = false
      resetDocumentForm()
      await fetchDocuments(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API save error, applying local state update:', error)
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].documents
      if (isEditingDocument.value && editingDocumentId.value) {
        const idx = currentList.findIndex((item: any) => item.id === editingDocumentId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...documentForm,
            fileName: documentForm.file ? documentForm.file.name : currentList[idx].fileName
          }
        }
      } else {
        const newItem: DocumentItem = {
          id: Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          documentName: documentForm.documentName,
          description: documentForm.description,
          requiredDate: documentForm.requiredDate,
          file: documentForm.file,
          fileName: documentForm.file ? documentForm.file.name : ''
        }
        currentList.push(newItem)
      }
      showDocumentModal.value = false
      resetDocumentForm()
    } finally {
      loading.value = false
    }
  }

  const deleteDocument = async (index: number) => {
    if (!selectedAssignmentLetter.value) return
    const item = documents.value[index]
    if (!item || !confirm('Are you sure you want to delete this document?')) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/fieldwork/documents/${item.id}`, {
        method: 'DELETE'
      })
      if (fieldworkData.value[selectedAssignmentLetter.value]?.documents) {
        fieldworkData.value[selectedAssignmentLetter.value].documents.splice(index, 1)
      }
      await fetchDocuments(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API delete error, applying local delete:', error)
      if (fieldworkData.value[selectedAssignmentLetter.value]?.documents) {
        fieldworkData.value[selectedAssignmentLetter.value].documents.splice(index, 1)
      }
    } finally {
      loading.value = false
    }
  }

  // Sample Form
  const sampleForm = reactive({
    documentName: '',
    documentNumber: '',
    date: '',
    description: ''
  })
  const showSampleModal = ref(false)
  const isEditingSample = ref(false)
  const isReadOnlySample = ref(false)
  const editingSampleId = ref<any>(null)

  const openSampleModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetSampleForm()
    isEditingSample.value = false
    isReadOnlySample.value = false
    editingSampleId.value = null
    showSampleModal.value = true
  }

  const editSample = (item: SampleItem) => {
    isEditingSample.value = true
    isReadOnlySample.value = false
    editingSampleId.value = item.id
    sampleForm.documentName = item.documentName
    sampleForm.documentNumber = item.documentNumber
    sampleForm.date = item.date
    sampleForm.description = item.description
    showSampleModal.value = true
  }

  const viewSample = (item: SampleItem) => {
    editSample(item)
    isEditingSample.value = false
    isReadOnlySample.value = true
  }

  const resetSampleForm = () => {
    sampleForm.documentName = ''
    sampleForm.documentNumber = ''
    sampleForm.date = ''
    sampleForm.description = ''
  }

  const saveSample = async () => {
    if (!selectedAssignmentLetter.value || isReadOnlySample.value) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = {
        ...sampleForm,
        assignmentLetterId: selectedAssignmentLetter.value
      }
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].samples

      let res: any = null
      if (isEditingSample.value && editingSampleId.value) {
        res = await $fetch(`${baseUrl}/fieldwork/samples/${editingSampleId.value}`, {
          method: 'PUT',
          body: payload
        })
        const updatedItem = res?.data || res
        const idx = currentList.findIndex((item: any) => item.id === editingSampleId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...payload,
            id: updatedItem?.id || editingSampleId.value
          }
        }
      } else {
        res = await $fetch(`${baseUrl}/fieldwork/samples`, {
          method: 'POST',
          body: payload
        })
        const createdItem = res?.data || res
        const newItem: SampleItem = {
          id: createdItem?.id || Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          documentName: sampleForm.documentName,
          documentNumber: sampleForm.documentNumber,
          date: sampleForm.date,
          description: sampleForm.description
        }
        currentList.push(newItem)
      }
      showSampleModal.value = false
      resetSampleForm()
      await fetchSamples(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API save error, applying local state update:', error)
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].samples
      if (isEditingSample.value && editingSampleId.value) {
        const idx = currentList.findIndex((item: any) => item.id === editingSampleId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...sampleForm
          }
        }
      } else {
        const newItem: SampleItem = {
          id: Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          documentName: sampleForm.documentName,
          documentNumber: sampleForm.documentNumber,
          date: sampleForm.date,
          description: sampleForm.description
        }
        currentList.push(newItem)
      }
      showSampleModal.value = false
      resetSampleForm()
    } finally {
      loading.value = false
    }
  }

  const deleteSample = async (index: number) => {
    if (!selectedAssignmentLetter.value) return
    const item = samples.value[index]
    if (!item || !confirm('Are you sure you want to delete this sample?')) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/fieldwork/samples/${item.id}`, {
        method: 'DELETE'
      })
      if (fieldworkData.value[selectedAssignmentLetter.value]?.samples) {
        fieldworkData.value[selectedAssignmentLetter.value].samples.splice(index, 1)
      }
      await fetchSamples(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API delete error, applying local delete:', error)
      if (fieldworkData.value[selectedAssignmentLetter.value]?.samples) {
        fieldworkData.value[selectedAssignmentLetter.value].samples.splice(index, 1)
      }
    } finally {
      loading.value = false
    }
  }

  // Test Control Form
  const testControlForm = reactive({
    controlName: '',
    controlDescription: '',
    controlType: '',
    testProcedure: '',
    testResult: '',
    finding: '',
    recommendation: '',
    mitigationPlan: '',
    pic: '',
    dueDate: ''
  })
  const showTestControlModal = ref(false)
  const isEditingTestControl = ref(false)
  const isReadOnlyTestControl = ref(false)
  const editingTestControlId = ref<any>(null)

  const openTestControlModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetTestControlForm()
    isEditingTestControl.value = false
    isReadOnlyTestControl.value = false
    editingTestControlId.value = null
    showTestControlModal.value = true
  }

  const editTestControl = (item: TestControlItem) => {
    isEditingTestControl.value = true
    isReadOnlyTestControl.value = false
    editingTestControlId.value = item.id
    testControlForm.controlName = item.controlName
    testControlForm.controlDescription = item.controlDescription
    testControlForm.controlType = item.controlType
    testControlForm.testProcedure = item.testProcedure
    testControlForm.testResult = item.testResult
    testControlForm.finding = item.finding
    testControlForm.recommendation = item.recommendation
    testControlForm.mitigationPlan = item.mitigationPlan
    testControlForm.pic = item.pic
    testControlForm.dueDate = item.dueDate
    showTestControlModal.value = true
  }

  const viewTestControl = (item: TestControlItem) => {
    editTestControl(item)
    isEditingTestControl.value = false
    isReadOnlyTestControl.value = true
  }

  const resetTestControlForm = () => {
    testControlForm.controlName = ''
    testControlForm.controlDescription = ''
    testControlForm.controlType = ''
    testControlForm.testProcedure = ''
    testControlForm.testResult = ''
    testControlForm.finding = ''
    testControlForm.recommendation = ''
    testControlForm.mitigationPlan = ''
    testControlForm.pic = ''
    testControlForm.dueDate = ''
  }

  const saveTestControl = async () => {
    if (!selectedAssignmentLetter.value || isReadOnlyTestControl.value) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = {
        ...testControlForm,
        assignmentLetterId: selectedAssignmentLetter.value
      }
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].testControls

      let res: any = null
      if (isEditingTestControl.value && editingTestControlId.value) {
        res = await $fetch(`${baseUrl}/fieldwork/test-controls/${editingTestControlId.value}`, {
          method: 'PUT',
          body: payload
        })
        const updatedItem = res?.data || res
        const idx = currentList.findIndex((item: any) => item.id === editingTestControlId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...payload,
            id: updatedItem?.id || editingTestControlId.value
          }
        }
      } else {
        res = await $fetch(`${baseUrl}/fieldwork/test-controls`, {
          method: 'POST',
          body: payload
        })
        const createdItem = res?.data || res
        const newItem: TestControlItem = {
          id: createdItem?.id || Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          controlName: testControlForm.controlName,
          controlDescription: testControlForm.controlDescription,
          controlType: testControlForm.controlType,
          testProcedure: testControlForm.testProcedure,
          testResult: testControlForm.testResult,
          finding: testControlForm.finding,
          recommendation: testControlForm.recommendation,
          mitigationPlan: testControlForm.mitigationPlan,
          pic: testControlForm.pic,
          dueDate: testControlForm.dueDate
        }
        currentList.push(newItem)
      }
      showTestControlModal.value = false
      resetTestControlForm()
      await fetchTestControls(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API save error, applying local state update:', error)
      ensureFieldworkDataHolder(selectedAssignmentLetter.value)
      const currentList = fieldworkData.value[selectedAssignmentLetter.value].testControls
      if (isEditingTestControl.value && editingTestControlId.value) {
        const idx = currentList.findIndex((item: any) => item.id === editingTestControlId.value)
        if (idx !== -1) {
          currentList[idx] = {
            ...currentList[idx],
            ...testControlForm
          }
        }
      } else {
        const newItem: TestControlItem = {
          id: Date.now().toString(),
          assignmentLetterId: selectedAssignmentLetter.value,
          controlName: testControlForm.controlName,
          controlDescription: testControlForm.controlDescription,
          controlType: testControlForm.controlType,
          testProcedure: testControlForm.testProcedure,
          testResult: testControlForm.testResult,
          finding: testControlForm.finding,
          recommendation: testControlForm.recommendation,
          mitigationPlan: testControlForm.mitigationPlan,
          pic: testControlForm.pic,
          dueDate: testControlForm.dueDate
        }
        currentList.push(newItem)
      }
      showTestControlModal.value = false
      resetTestControlForm()
    } finally {
      loading.value = false
    }
  }

  const deleteTestControl = async (index: number) => {
    if (!selectedAssignmentLetter.value) return
    const item = testControls.value[index]
    if (!item || !confirm('Are you sure you want to delete this test control?')) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/fieldwork/test-controls/${item.id}`, {
        method: 'DELETE'
      })
      if (fieldworkData.value[selectedAssignmentLetter.value]?.testControls) {
        fieldworkData.value[selectedAssignmentLetter.value].testControls.splice(index, 1)
      }
      await fetchTestControls(selectedAssignmentLetter.value)
    } catch (error) {
      console.error('API delete error, applying local delete:', error)
      if (fieldworkData.value[selectedAssignmentLetter.value]?.testControls) {
        fieldworkData.value[selectedAssignmentLetter.value].testControls.splice(index, 1)
      }
    } finally {
      loading.value = false
    }
  }

  const interviewCount = computed(() => interviews.value.length)
  const observationCount = computed(() => observations.value.length)
  const documentCount = computed(() => documents.value.length)
  const sampleCount = computed(() => samples.value.length)
  const testControlCount = computed(() => testControls.value.length)

  const effectiveControls = computed(() =>
    testControls.value.filter(tc => tc.testResult === 'Effective').length
  )

  const ineffectiveControls = computed(() =>
    testControls.value.filter(tc => tc.testResult === 'Ineffective').length
  )

  return {
    tabs,
    options,
    selectedAssignmentLetter,
    publishedAssignmentLetters,
    hasSelectedAssignmentLetter,
    memberOptions,
    interviews,
    observations,
    documents,
    samples,
    testControls,
    interviewForm,
    showInterviewModal,
    isEditingInterview,
    isReadOnlyInterview,
    openInterviewModal,
    editInterview,
    viewInterview,
    handleInterviewFileChange,
    saveInterview,
    deleteInterview,
    observationForm,
    showObservationModal,
    isEditingObservation,
    isReadOnlyObservation,
    openObservationModal,
    editObservation,
    viewObservation,
    handleObservationFileChange,
    saveObservation,
    deleteObservation,
    documentForm,
    showDocumentModal,
    isEditingDocument,
    isReadOnlyDocument,
    openDocumentModal,
    editDocument,
    viewDocument,
    handleDocumentFileChange,
    saveDocument,
    deleteDocument,
    sampleForm,
    showSampleModal,
    isEditingSample,
    isReadOnlySample,
    openSampleModal,
    editSample,
    viewSample,
    saveSample,
    deleteSample,
    testControlForm,
    showTestControlModal,
    isEditingTestControl,
    isReadOnlyTestControl,
    openTestControlModal,
    editTestControl,
    viewTestControl,
    saveTestControl,
    deleteTestControl,
    interviewCount,
    observationCount,
    documentCount,
    sampleCount,
    testControlCount,
    effectiveControls,
    ineffectiveControls,
    fetchAllFieldworkData,
    loading,
    errorMsg
  }
})
