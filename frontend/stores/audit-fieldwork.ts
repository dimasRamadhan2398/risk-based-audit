import { defineStore } from 'pinia'
import { reactive, ref, computed } from 'vue'
import { useAssignmentLetterStore } from './assignment-letter'

export interface InterviewItem {
  id: number
  assignmentLetterId: string
  interviewee: string
  intervieweePosition: string
  interviewer: string
  interviewerPosition: string
  date: string
  topic: string
  file: File | null
}

export interface ObservationItem {
  id: number
  assignmentLetterId: string
  activity: string
  location: string
  date: string
  observer: string
  file: File | null
}

export interface DocumentItem {
  id: number
  assignmentLetterId: string
  documentName: string
  description: string
  requiredDate: string
  file: File | null
}

export interface SampleItem {
  id: number
  assignmentLetterId: string
  documentName: string
  documentNumber: string
  date: string
  description: string
}

export interface TestControlItem {
  id: number
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
  // Get Assignment Letter store for reference
  const assignmentLetterStore = useAssignmentLetterStore()

  // Selected Assignment Letter
  const selectedAssignmentLetter = ref<string>('')

  // Get only Published Assignment Letters for dropdown
  const publishedAssignmentLetters = computed(() => {
    return assignmentLetterStore.assignmentLetterList
      .filter((st: any) => st.status === 'Published')
      .map((st: any) => st.letterNumber)
  })

  // Check if user has selected an Assignment Letter
  const hasSelectedAssignmentLetter = computed(() => !!selectedAssignmentLetter.value && selectedAssignmentLetter.value !== '')

  // Get the full assignment letter object for selected letter
  const selectedAssignmentLetterData = computed(() => {
    return assignmentLetterStore.assignmentLetterList.find(
      (st: any) => st.letterNumber === selectedAssignmentLetter.value
    )
  })

  // Get members list from selected assignment letter for interview dropdowns
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

  // Tabs Configuration
  const tabs = [
    { label: 'Interview', slot: 'tab01', icon: 'i-heroicons-microphone' },
    { label: 'Observation', slot: 'tab02', icon: 'i-heroicons-eye' },
    { label: 'Document Collection', slot: 'tab03', icon: 'i-heroicons-document-duplicate' },
    { label: 'Sample Data', slot: 'tab04', icon: 'i-heroicons-table-cells' },
    { label: 'Test Controls', slot: 'tab05', icon: 'i-heroicons-shield-check' },
    { label: 'Working Papers', slot: 'tab06', icon: 'i-heroicons-document-text' }
  ]

  // Options
  const options = {
    positions: ['Manager', 'Supervisor', 'Staff', 'Director', 'Head of Department', 'Other'],
    controlTypes: ['Preventive', 'Detective', 'Corrective', 'Manual', 'Automated'],
    testResults: ['Effective', 'Ineffective', 'Partially Effective', 'Not Tested'],
    auditTopics: ['Internal Control', 'Compliance', 'Risk Management', 'Process Efficiency', 'Other']
  }

  // Data storage keyed by Assignment Letter ID
  const fieldworkData = ref<Record<string, {
    interviews: InterviewItem[]
    observations: ObservationItem[]
    documents: DocumentItem[]
    samples: SampleItem[]
    testControls: TestControlItem[]
  }>>({})

  // Helper to ensure data structure exists for selected Assignment Letter
  const ensureDataExists = () => {
    if (!selectedAssignmentLetter.value) return
    if (!fieldworkData.value[selectedAssignmentLetter.value]) {
      fieldworkData.value[selectedAssignmentLetter.value] = {
        interviews: [],
        observations: [],
        documents: [],
        samples: [],
        testControls: []
      }
    }
  }

  // Computed properties for current Assignment Letter data
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

  const openInterviewModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetInterviewForm()
    showInterviewModal.value = true
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

  const saveInterview = () => {
    if (!selectedAssignmentLetter.value) return
    ensureDataExists()
    fieldworkData.value[selectedAssignmentLetter.value]!.interviews.push({
      id: Date.now(),
      assignmentLetterId: selectedAssignmentLetter.value,
      ...interviewForm
    })
    showInterviewModal.value = false
    resetInterviewForm()
    alert('Interview data saved successfully!')
  }

  const deleteInterview = (index: number) => {
    if (!selectedAssignmentLetter.value) return
    fieldworkData.value[selectedAssignmentLetter.value]!.interviews.splice(index, 1)
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

  const openObservationModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetObservationForm()
    showObservationModal.value = true
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

  const saveObservation = () => {
    if (!selectedAssignmentLetter.value) return
    ensureDataExists()
    fieldworkData.value[selectedAssignmentLetter.value]!.observations.push({
      id: Date.now(),
      assignmentLetterId: selectedAssignmentLetter.value,
      ...observationForm
    })
    showObservationModal.value = false
    resetObservationForm()
    alert('Observation data saved successfully!')
  }

  const deleteObservation = (index: number) => {
    if (!selectedAssignmentLetter.value) return
    fieldworkData.value[selectedAssignmentLetter.value]!.observations.splice(index, 1)
  }

  // Document Form
  const documentForm = reactive({
    documentName: '',
    description: '',
    requiredDate: '',
    file: null as File | null
  })
  const showDocumentModal = ref(false)

  const openDocumentModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetDocumentForm()
    showDocumentModal.value = true
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

  const saveDocument = () => {
    if (!selectedAssignmentLetter.value) return
    ensureDataExists()
    fieldworkData.value[selectedAssignmentLetter.value]!.documents.push({
      id: Date.now(),
      assignmentLetterId: selectedAssignmentLetter.value,
      ...documentForm
    })
    showDocumentModal.value = false
    resetDocumentForm()
    alert('Document data saved successfully!')
  }

  const deleteDocument = (index: number) => {
    if (!selectedAssignmentLetter.value) return
    fieldworkData.value[selectedAssignmentLetter.value]!.documents.splice(index, 1)
  }

  // Sample Form
  const sampleForm = reactive({
    documentName: '',
    documentNumber: '',
    date: '',
    description: ''
  })
  const showSampleModal = ref(false)

  const openSampleModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetSampleForm()
    showSampleModal.value = true
  }

  const resetSampleForm = () => {
    sampleForm.documentName = ''
    sampleForm.documentNumber = ''
    sampleForm.date = ''
    sampleForm.description = ''
  }

  const saveSample = () => {
    if (!selectedAssignmentLetter.value) return
    ensureDataExists()
    fieldworkData.value[selectedAssignmentLetter.value]!.samples.push({
      id: Date.now(),
      assignmentLetterId: selectedAssignmentLetter.value,
      ...sampleForm
    })
    showSampleModal.value = false
    resetSampleForm()
    alert('Sample data saved successfully!')
  }

  const deleteSample = (index: number) => {
    if (!selectedAssignmentLetter.value) return
    fieldworkData.value[selectedAssignmentLetter.value]!.samples.splice(index, 1)
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

  const openTestControlModal = () => {
    if (!selectedAssignmentLetter.value) {
      alert('Please select an Assignment Letter first!')
      return
    }
    resetTestControlForm()
    showTestControlModal.value = true
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

  const saveTestControl = () => {
    if (!selectedAssignmentLetter.value) return
    ensureDataExists()
    fieldworkData.value[selectedAssignmentLetter.value]!.testControls.push({
      id: Date.now(),
      assignmentLetterId: selectedAssignmentLetter.value,
      ...testControlForm
    })
    showTestControlModal.value = false
    resetTestControlForm()
    alert('Test Control data saved successfully!')
  }

  const deleteTestControl = (index: number) => {
    if (!selectedAssignmentLetter.value) return
    fieldworkData.value[selectedAssignmentLetter.value]!.testControls.splice(index, 1)
  }

  // Computed for statistics (for current Assignment Letter)
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
    // Data
    interviews,
    observations,
    documents,
    samples,
    testControls,
    // Interview
    interviewForm,
    showInterviewModal,
    openInterviewModal,
    handleInterviewFileChange,
    saveInterview,
    deleteInterview,
    // Observation
    observationForm,
    showObservationModal,
    openObservationModal,
    handleObservationFileChange,
    saveObservation,
    deleteObservation,
    // Document
    documentForm,
    showDocumentModal,
    openDocumentModal,
    handleDocumentFileChange,
    saveDocument,
    deleteDocument,
    // Sample
    sampleForm,
    showSampleModal,
    openSampleModal,
    saveSample,
    deleteSample,
    // Test Control
    testControlForm,
    showTestControlModal,
    openTestControlModal,
    saveTestControl,
    deleteTestControl,
    // Statistics
    interviewCount,
    observationCount,
    documentCount,
    sampleCount,
    testControlCount,
    effectiveControls,
    ineffectiveControls
  }
})
