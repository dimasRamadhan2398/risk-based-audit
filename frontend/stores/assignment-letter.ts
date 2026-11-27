import type { TableColumn } from '@nuxt/ui';
import { defineStore } from 'pinia'
import { type AssignmentLetter, type AssignmentLetterForm, type AssignmentLetterStatus, AuditCategory } from '~/types/audit'
import { useToastNotification } from '~/components/shared/ToastNotification.vue';
import { extractErrorMessage } from '~/utils/error';
import { getAuditServiceBaseUrl as resolveAuditUrl } from '~/composables/useApiUrl';
import { useI18n } from '~/composables/useI18n';
import { useRbac } from '~/composables/useRbac';
import { getFiscalYearStrings } from '~/composables/useFiscalYear';
import { useEmployeeApi } from '~/composables/useEmployeeApi';
import { useDepartmentApi } from '~/composables/useDepartmentApi';
import type { Employee, Department } from '~/types/master';

export interface AssignmentLetterState {
  isModalOpen: boolean;
  editingId: string | null;
  assignmentLetterList: AssignmentLetter[];
  form: AssignmentLetterForm;
  columns: (TableColumn<AssignmentLetter> & { class?: string })[];
  options: {
    auditTeam: { value: string, label: string }[];
    role: string[];
    yearOptions: string[];
  };
  loading: boolean;
  errorMsg: string;
  employees: Employee[];
  employeesLoading: boolean;
  employeesLoaded: boolean;
  employeesError: string;
  departments: Department[];
  departmentsLoading: boolean;
  departmentsLoaded: boolean;
  departmentsError: string;
}

export const useAssignmentLetterStore = defineStore('assignment-letter', {
  state: (): AssignmentLetterState => ({
    isModalOpen: false,
    editingId: null,
    assignmentLetterList: [],
    loading: false,
    errorMsg: '',
    employees: [],
    employeesLoading: false,
    employeesLoaded: false,
    employeesError: '',
    departments: [],
    departmentsLoading: false,
    departmentsLoaded: false,
    departmentsError: '',
    form: {
      auditTitle: '',
      leader: '',
      category: AuditCategory.ASSURANCE,
      auditYear: new Date().getFullYear().toString(),
      auditTeam: 'SKAI',
      startPeriod: '',
      finishPeriod: '',
      workingUnit: '',
      workingUnitDepartmentId: '',
      companyId: '',
      companyName: '',
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
    // Headers are resolved in AssignmentLetterTable.vue (assignmentLetter.table.columns.*).
    columns: [
      { accessorKey: 'letterNumber', class: 'w-40 whitespace-nowrap font-medium' },
      { accessorKey: 'auditTitle', class: 'min-w-[200px] max-w-xs' },
      { accessorKey: 'workingUnit', class: 'w-36' },
      { accessorKey: 'executionPeriod', class: 'w-48 whitespace-nowrap' },
      { accessorKey: 'auditTeam', class: 'w-40' },
      { accessorKey: 'status', class: 'w-28 whitespace-nowrap text-center' },
      { accessorKey: 'actions', class: 'w-24 whitespace-nowrap text-center' }
    ],
    options: {
      auditTeam: [
        { value: 'SKAI', label: 'Satuan Kerja Audit Internal' },
        { value: 'SPI', label: 'Satuan Pengawas Intern' },
        { value: 'CAE', label: 'Chief Audit Executive' }
      ],
      role: ['Person in Charge', 'Supervisor', 'Chairperson', 'Member'],
      yearOptions: getFiscalYearStrings()
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
    },

    employeeOptions: (state): string[] => {
      const list = state.employees.map(e => e.employee_code ? `${e.full_name} (${e.employee_code})` : e.full_name)
      if (state.form.membersList && Array.isArray(state.form.membersList)) {
        for (const member of state.form.membersList) {
          if (member.name && !list.includes(member.name)) {
            list.unshift(member.name)
          }
        }
      }
      return list.sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
    },

    getAvailableMemberOptions: (state) => (index: number): string[] => {
      const excludedKeys = new Set<string>()
      const excludedText = new Set<string>()

      if (Array.isArray(state.form.membersList)) {
        state.form.membersList.forEach((m, idx) => {
          if (idx !== index && m.name && m.name.trim()) {
            const trimmed = m.name.trim()
            excludedText.add(trimmed.toLowerCase())
            const match = state.employees.find(
              e => e.id.toLowerCase() === trimmed.toLowerCase() ||
                (e.employee_code && e.employee_code.toLowerCase() === trimmed.toLowerCase()) ||
                (e.full_name && e.full_name.toLowerCase() === trimmed.toLowerCase()) ||
                (e.employee_code && `${e.full_name} (${e.employee_code})`.toLowerCase() === trimmed.toLowerCase())
            )
            if (match) {
              excludedKeys.add(match.id.toLowerCase())
              if (match.employee_code) excludedText.add(match.employee_code.toLowerCase())
              excludedText.add(match.full_name.toLowerCase())
              if (match.employee_code) {
                excludedText.add(`${match.full_name} (${match.employee_code})`.toLowerCase())
              }
            }
          }
        })
      }

      // Base options from master employee list
      const list = state.employees.map(e => e.employee_code ? `${e.full_name} (${e.employee_code})` : e.full_name)
      // Current row's own value, if present and not in list (e.g. custom/legacy entry)
      const currentName = state.form.membersList?.[index]?.name?.trim()
      if (currentName && !list.includes(currentName)) {
        list.unshift(currentName)
      }

      const available = list.filter(opt => {
        const optLower = opt.toLowerCase()
        if (excludedText.has(optLower)) return false

        const match = state.employees.find(
          e => (e.employee_code && `${e.full_name} (${e.employee_code})`.toLowerCase() === optLower) ||
            e.full_name.toLowerCase() === optLower
        )
        if (match && excludedKeys.has(match.id.toLowerCase())) {
          return false
        }
        return true
      })

      return available.sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
    },

    isDuplicateMember: (state) => (name: string, index: number): boolean => {
      if (!name || !name.trim()) return false
      const trimmed = name.trim().toLowerCase()
      const match = state.employees.find(
        e => e.id.toLowerCase() === trimmed ||
          (e.employee_code && e.employee_code.toLowerCase() === trimmed) ||
          (e.full_name && e.full_name.toLowerCase() === trimmed) ||
          (e.employee_code && `${e.full_name} (${e.employee_code})`.toLowerCase() === trimmed)
      )
      const targetKey = match ? match.id.toLowerCase() : trimmed

      return state.form.membersList.some((m, idx) => {
        if (idx === index || !m.name || !m.name.trim()) return false
        const otherTrimmed = m.name.trim().toLowerCase()
        const otherMatch = state.employees.find(
          e => e.id.toLowerCase() === otherTrimmed ||
            (e.employee_code && e.employee_code.toLowerCase() === otherTrimmed) ||
            (e.full_name && e.full_name.toLowerCase() === otherTrimmed) ||
            (e.employee_code && `${e.full_name} (${e.employee_code})`.toLowerCase() === otherTrimmed)
        )
        const otherKey = otherMatch ? otherMatch.id.toLowerCase() : otherTrimmed
        return otherKey === targetKey
      })
    },

    hasDuplicateMembers: (state): boolean => {
      if (!Array.isArray(state.form.membersList) || state.form.membersList.length <= 1) {
        return false
      }
      return state.form.membersList.some((m, idx) => {
        if (!m.name || !m.name.trim()) return false
        const trimmed = m.name.trim().toLowerCase()
        const match = state.employees.find(
          e => e.id.toLowerCase() === trimmed ||
            (e.employee_code && e.employee_code.toLowerCase() === trimmed) ||
            (e.full_name && e.full_name.toLowerCase() === trimmed) ||
            (e.employee_code && `${e.full_name} (${e.employee_code})`.toLowerCase() === trimmed)
        )
        const targetKey = match ? match.id.toLowerCase() : trimmed

        return state.form.membersList.some((other, otherIdx) => {
          if (otherIdx <= idx || !other.name || !other.name.trim()) return false
          const otherTrimmed = other.name.trim().toLowerCase()
          const otherMatch = state.employees.find(
            e => e.id.toLowerCase() === otherTrimmed ||
              (e.employee_code && e.employee_code.toLowerCase() === otherTrimmed) ||
              (e.full_name && e.full_name.toLowerCase() === otherTrimmed) ||
              (e.employee_code && `${e.full_name} (${e.employee_code})`.toLowerCase() === otherTrimmed)
          )
          const otherKey = otherMatch ? otherMatch.id.toLowerCase() : otherTrimmed
          return otherKey === targetKey
        })
      })
    },

    /** Unit Kerja choices: active master departments only (value = department id). */
    workUnitItems: (state): { label: string, value: string }[] =>
      state.departments
        .map(d => ({
          label: d.department_code ? `${d.department_name} (${d.department_code})` : d.department_name,
          value: d.id
        }))
        .sort((a, b) => a.label.localeCompare(b.label, 'id', { sensitivity: 'base' })),

    /** A letter whose Unit Kerja text is not linked to a master department (saved before the link existed). */
    workingUnitUnlinked: (state): boolean => !!state.form.workingUnit && !state.form.workingUnitDepartmentId
  },
  actions: {
    getAuditServiceBaseUrl() {
      return resolveAuditUrl();
    },

    resolveEmployeeOption(val?: string): string {
      if (!val) return ''
      const match = this.employees.find(
        e => e.id === val || e.employee_code === val || e.full_name === val || `${e.full_name} (${e.employee_code})` === val
      )
      if (match) {
        return match.employee_code ? `${match.full_name} (${match.employee_code})` : match.full_name
      }
      return val
    },

    formatEmployeeName(value?: string): string {
      if (!value) return '-'
      const found = this.employees.find(
        e => e.id === value || e.employee_code === value || e.full_name === value
      )
      if (found) {
        return found.employee_code ? `${found.full_name} (${found.employee_code})` : found.full_name
      }
      return value
    },

    formatLeaderDisplay(letter: AssignmentLetter): string {
      if (letter.leader) {
        return this.formatEmployeeName(letter.leader)
      }
      if (Array.isArray(letter.membersList) && letter.membersList.length > 0) {
        const pic = letter.membersList.find(m => m.role === 'Person in Charge') || letter.membersList.find(m => m.role === 'Chairperson')
        if (pic?.name) {
          return this.formatEmployeeName(pic.name)
        }
      }
      return '-'
    },

    /** Id of the one active master department val names (id, code, name or "Name (CODE)"), else ''. */
    matchDepartmentId(val?: string | null): string {
      const v = (val || '').trim()
      if (!v) return ''
      const matches = this.departments.filter(
        d => d.id === v
          || d.department_code === v
          || d.department_name === v
          || `${d.department_name} (${d.department_code})` === v
      )
      return matches.length === 1 ? matches[0]!.id : ''
    },

    /** Pick Unit Kerja from the master: stores the id and its display text. */
    selectWorkingUnit(id?: string | null) {
      const dept = this.departments.find(d => d.id === id)
      this.form.workingUnitDepartmentId = dept ? dept.id : ''
      this.form.workingUnit = dept
        ? (dept.department_code ? `${dept.department_name} (${dept.department_code})` : dept.department_name)
        : ''
    },

    resolveDepartmentOption(val?: string): string {
      if (!val) return ''
      const match = this.departments.find(
        d => d.id === val ||
          d.department_code === val ||
          d.department_name === val ||
          `${d.department_name} (${d.department_code})` === val
      )
      if (match) {
        return match.department_code ? `${match.department_name} (${match.department_code})` : match.department_name
      }
      return val
    },

    formatDepartmentName(value?: string): string {
      if (!value) return '-'
      const found = this.departments.find(
        d => d.id === value ||
          d.department_code === value ||
          d.department_name === value ||
          `${d.department_name} (${d.department_code})` === value
      )
      if (found) {
        return found.department_code ? `${found.department_name} (${found.department_code})` : found.department_name
      }
      return value
    },

    async fetchDepartments(force = false) {
      if (this.departmentsLoaded && !force) return
      if (this.departmentsLoading) return
      this.departmentsLoading = true
      this.departmentsError = ''
      try {
        const api = useDepartmentApi()
        const list = await api.getAllDepartments()
        this.departments = (Array.isArray(list) ? list : [])
          .filter(d => d?.department_name && d.is_active !== false)
          .sort((a, b) => a.department_name.localeCompare(b.department_name, 'id', { sensitivity: 'base' }))
        this.departmentsLoaded = true

        if (this.form.workingUnit) {
          this.form.workingUnit = this.resolveDepartmentOption(this.form.workingUnit)
          if (!this.form.workingUnitDepartmentId) {
            this.form.workingUnitDepartmentId = this.matchDepartmentId(this.form.workingUnit)
          }
        }
      } catch (error: any) {
        console.error('Failed to fetch departments for assignment letter:', error)
        this.departmentsError = extractErrorMessage(error, useI18n().t('toasts.common.loadDepartmentsFailed'))
        if (!this.departmentsLoaded) {
          this.departments = []
        }
      } finally {
        this.departmentsLoading = false
      }
    },

    async fetchEmployees(force = false) {
      if (this.employeesLoaded && !force) return
      if (this.employeesLoading) return
      this.employeesLoading = true
      this.employeesError = ''
      try {
        const api = useEmployeeApi()
        const list = await api.getAllEmployees()
        this.employees = (Array.isArray(list) ? list : [])
          .filter(e => e?.full_name && e.is_active !== false)
          .sort((a, b) => a.full_name.localeCompare(b.full_name, 'id', { sensitivity: 'base' }))
        this.employeesLoaded = true

        if (this.form.membersList && Array.isArray(this.form.membersList)) {
          this.form.membersList.forEach(m => {
            if (m.name) {
              m.name = this.resolveEmployeeOption(m.name)
            }
          })
        }
      } catch (error: any) {
        console.error('Failed to fetch employees for assignment letter:', error)
        this.employeesError = extractErrorMessage(error, useI18n().t('toasts.common.loadEmployeesFailed'))
        if (!this.employeesLoaded) {
          this.employees = []
        }
      } finally {
        this.employeesLoading = false
      }
    },

    async fetchAssignmentLetters() {
      if (!this.employeesLoaded && !this.employeesLoading) {
        this.fetchEmployees().catch(() => {})
      }
      if (!this.departmentsLoaded && !this.departmentsLoading) {
        this.fetchDepartments().catch(() => {})
      }
      this.loading = true
      this.errorMsg = ''

      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        const response: any = await $fetch(`${baseUrl}/assignment-letters?page_size=100`, {
          method: 'GET'
        })
        let items: AssignmentLetter[] = []
        if (response && response.data && Array.isArray(response.data.items)) {
          items = [...response.data.items]
          const totalPages = Number(response.data.pagination?.total_pages || 1)
          for (let page = 2; page <= Math.min(totalPages, 20); page++) {
            const nextResp: any = await $fetch(`${baseUrl}/assignment-letters?page=${page}&page_size=100`, {
              method: 'GET'
            })
            if (nextResp?.data && Array.isArray(nextResp.data.items)) {
              items.push(...nextResp.data.items)
            }
          }
        } else if (response && Array.isArray(response.items)) {
          items = response.items
        } else if (Array.isArray(response)) {
          items = response
        }

        this.assignmentLetterList = items
      } catch (error: any) {
        console.error('Failed to fetch assignment letters:', error)
        this.errorMsg = extractErrorMessage(error, useI18n().t('toasts.assignmentLetter.loadFailed'))
        this.assignmentLetterList = []
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

    generateNomorSurat(auditTeam?: string, year?: string): string {
      const rawYear = (year || this.form?.auditYear || '').toString().trim()
      let cleanYear = ''
      const yearMatch = rawYear.match(/\b((?:19|20|21)\d{2})\b/)
      if (yearMatch && yearMatch[1]) {
        cleanYear = yearMatch[1]
      } else if (/^\d{4}$/.test(rawYear)) {
        cleanYear = rawYear
      } else {
        cleanYear = new Date().getFullYear().toString()
      }

      const targetTeam = (auditTeam || this.form?.auditTeam || 'SKAI')
        .trim()
        .toUpperCase()
        .replace(/\//g, '-') || 'SKAI'

      // Scan existing letters strictly within the same audit team and year so each team
      // counts independently from 1 and outliers or other teams cannot cause sequence jumps.
      const usedSeqs = new Set<number>()
      for (const item of this.assignmentLetterList) {
        const parts = (item.letterNumber || '').trim().split('/')
        if (parts.length !== 3) continue
        const seqMatch = parts[0]!.trim().match(/^ST-(\d+)$/i)
        const itemTeam = parts[1]!.trim().toUpperCase()
        const itemYear = parts[2]!.trim()
        if (seqMatch && seqMatch[1] && itemTeam === targetTeam && itemYear === cleanYear) {
          const parsed = parseInt(seqMatch[1], 10)
          if (!isNaN(parsed) && parsed > 0) {
            usedSeqs.add(parsed)
          }
        }
      }

      let nextSeq = 1
      while (usedSeqs.has(nextSeq)) {
        nextSeq++
      }
      const paddedCount = nextSeq.toString().padStart(3, '0')
      return `ST-${paddedCount}/${targetTeam}/${cleanYear}`
    },

    openModal() {
      this.editingId = null
      this.errorMsg = ''
      Object.assign(this.form, {
        auditTitle: '',
        leader: '',
        category: AuditCategory.ASSURANCE,
        auditYear: new Date().getFullYear().toString(),
        auditTeam: 'SKAI',
        startPeriod: '',
        finishPeriod: '',
        workingUnit: '',
        workingUnitDepartmentId: '',
        companyId: '',
        companyName: '',
        auditPurpose: '',
        letterDate: '',
        caeSignature: '',
        membersList: [{ name: '', role: 'Chairperson' }],
        purposeList: [''],
        scopeList: [''],
        ccList: ['President Director']
      })
      if (!this.employeesLoaded && !this.employeesLoading) {
        this.fetchEmployees().catch(() => {})
      }
      if (!this.departmentsLoaded && !this.departmentsLoading) {
        this.fetchDepartments().catch(() => {})
      }
      this.isModalOpen = true
    },

    openEditModal(letter: AssignmentLetter) {
      this.editingId = letter.id
      this.errorMsg = ''

      Object.assign(this.form, {
        auditTitle: letter.auditTitle,
        leader: letter.leader,
        category: letter.category,
        auditYear: letter.auditYear,
        auditTeam: letter.auditTeam,
        startPeriod: letter.startPeriod,
        finishPeriod: letter.finishPeriod,
        workingUnit: this.resolveDepartmentOption(letter.workingUnit),
        workingUnitDepartmentId: letter.workingUnitDepartmentId || this.matchDepartmentId(letter.workingUnit),
        companyId: letter.companyId || '',
        companyName: letter.companyName || '',
        auditPurpose: letter.auditPurpose,
        letterDate: letter.letterDate
          ? String(letter.letterDate).slice(0, 10)
          : '',
        caeSignature: letter.caeSignature || '',
        membersList: (letter.membersList || []).map(member => ({
          ...member,
          name: this.resolveEmployeeOption(member.name)
        })),
        purposeList: [...letter.purposeList],
        scopeList: [...letter.scopeList],
        ccList: [...letter.ccList]
      })

      if (!this.employeesLoaded && !this.employeesLoading) {
        this.fetchEmployees().catch(() => {})
      }
      if (!this.departmentsLoaded && !this.departmentsLoading) {
        this.fetchDepartments().catch(() => {})
      }

      this.isModalOpen = true
    },

    closeModal() {
      this.isModalOpen = false
      this.editingId = null
    },

    async handleSubmit() {
      this.errorMsg = ''

      if (!this.form.auditTitle) {
        this.errorMsg = 'Audit Title must be filled in.'
        return
      }

      if (!this.form.category) {
        this.errorMsg = 'Audit Category must be selected.'
        return
      }

      if (!this.form.auditYear) {
        this.errorMsg = 'Audit Year must be selected.'
        return
      }

      if (!this.form.auditTeam) {
        this.errorMsg = 'Audit Team must be selected.'
        return
      }

      // Unit Kerja must be a master department. Only when the master could not be
      // loaded may an existing letter be saved with its stored text (the id is then
      // omitted, so the server keeps what the letter has).
      const keepStoredUnit = !!this.editingId && !this.departmentsLoaded && !!this.form.workingUnit
      if (!this.form.workingUnitDepartmentId && !keepStoredUnit) {
        this.errorMsg = useI18n().t('assignmentLetter.validation.workUnitRequired')
        return
      }

      const isMembersValid = this.form.membersList.every(m => m.name && m.role)
      if (!isMembersValid) {
        this.errorMsg = useI18n().t('toasts.assignmentLetter.membersMissing')
        return
      }

      if (this.hasDuplicateMembers) {
        this.errorMsg = useI18n().t('toasts.assignmentLetter.duplicateMembers')
        useToastNotification().showError(useI18n().t('toasts.assignmentLetter.validationFailed'), this.errorMsg)
        return
      }

      const isCcValid = this.form.ccList.every(cc => cc && cc.trim() !== '')
      if (!isCcValid) {
        this.errorMsg = 'All CC fields must be filled in or removed.'
        return
      }

      if (this.dateError || !this.form.startPeriod || !this.form.finishPeriod) {
        this.errorMsg = this.dateError || 'Please fill in start and end period.'
        return
      }

      if (this.form.membersList.length < 3) {
        const proceed = await useGlobalModalStore().confirmSubmit({ description: "Template suggests at least 3 team members. Continue saving?" })
        if (!proceed) return
      }

      this.loading = true
      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        const executionPeriod =
          `${this.form.startPeriod} to ${this.form.finishPeriod}`

        const picMember = this.form.membersList.find(m => m.role === 'Person in Charge') || this.form.membersList.find(m => m.role === 'Chairperson')
        if (picMember) {
          this.form.leader = picMember.name
        }

        if (this.editingId) {
          const existingLetter = this.assignmentLetterList.find(
            letter => letter.id === this.editingId
          )

          if (!existingLetter) {
            throw new Error('Assignment letter not found.')
          }

          const payload: any = {
            ...this.form,
            letterNumber: existingLetter.letterNumber,
            status: existingLetter.status,
            executionPeriod
          }
          if (payload.letterDate === '') payload.letterDate = null
          if (!payload.companyId) delete payload.companyId
          if (!payload.workingUnitDepartmentId) delete payload.workingUnitDepartmentId

          await $fetch(
            `${baseUrl}/assignment-letters/${this.editingId}`,
            {
              method: 'PUT',
              body: payload
            }
          )
          useToastNotification().success('Assignment letter updated successfully!')
        } else {
          const letterNumber = this.generateNomorSurat(
            this.form.auditTeam,
            this.form.auditYear
          )

          const payload: any = {
            ...this.form,
            letterNumber,
            status: 'Draft',
            executionPeriod
          }
          if (payload.letterDate === '') payload.letterDate = null
          if (!payload.companyId) delete payload.companyId
          if (!payload.workingUnitDepartmentId) delete payload.workingUnitDepartmentId

          await $fetch(`${baseUrl}/assignment-letters`, {
            method: 'POST',
            body: payload
          })
          useToastNotification().success(useI18n().t('toasts.assignmentLetter.created'))
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
        const fallback = this.editingId
          ? useI18n().t('toasts.assignmentLetter.updateFailed')
          : useI18n().t('toasts.assignmentLetter.saveFailed')
        const detail = extractErrorMessage(error, fallback)
        this.errorMsg = detail
        useToastNotification().showError(fallback, detail)
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
        useToastNotification().showSuccess(useI18n().t('toasts.assignmentLetter.created'))
        await this.fetchAssignmentLetters()
      } catch (error: any) {
        console.error(error)
        const detail = extractErrorMessage(error, useI18n().t('toasts.assignmentLetter.createFailed'))
        this.errorMsg = detail
        useToastNotification().showError(useI18n().t('toasts.assignmentLetter.createFailed'), detail)
      } finally {
        this.loading = false
      }
    },

    async deleteSuratTugas(id: string) {
      if (!useRbac().isAdmin.value) {
        useToastNotification().showError(
          useI18n().t('toasts.assignmentLetter.deleteFailed'),
          'Only ADMIN can delete assignment letters.'
        )
        return
      }
      if (!await useGlobalModalStore().confirmDelete({ description: "Are you sure you want to delete this assignment letter?" })) return
      this.loading = true
      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/assignment-letters/${id}`, {
          method: 'DELETE'
        })
        useToastNotification().showSuccess(useI18n().t('toasts.assignmentLetter.deleted'))
        await this.fetchAssignmentLetters()
      } catch (error: any) {
        console.error(error)
        const detail = extractErrorMessage(error, useI18n().t('toasts.assignmentLetter.deleteFailed'))
        this.errorMsg = detail
        useToastNotification().showError(useI18n().t('toasts.assignmentLetter.deleteFailed'), detail)
      } finally {
        this.loading = false
      }
    },

    async publishLetter(letter: AssignmentLetter) {
      const { t } = useI18n()
      const confirmed = await useGlobalModalStore().confirmSubmit({
        // The confirmation modal shows `description` as its header
        description: t('assignmentLetterPublish.title', { number: letter.letterNumber }),
        body: [
          t('assignmentLetterPublish.confirmQuestion'),
          t('assignmentLetterPublish.availability')
        ],
        confirmLabel: t('assignmentLetterPublish.confirmButton')
      })
      if (!confirmed) return
      await this.changeStatus(letter.id, 'Published')
    },

    async changeStatus(id: string, status: AssignmentLetterStatus) {
      this.loading = true
      try {
        const baseUrl = this.getAuditServiceBaseUrl()
        // Send only the status: the update is a partial update, and resending
        // the whole table row would overwrite other fields with display data
        await $fetch(`${baseUrl}/assignment-letters/${id}`, {
          method: 'PUT',
          body: { status }
        })
        useToastNotification().showSuccess(useI18n().t('toasts.assignmentLetter.statusChanged', { status: useI18n().t(`assignmentLetter.table.status.${status}`) }))
        await this.fetchAssignmentLetters()
      } catch (error: any) {
        console.error(error)
        const detail = extractErrorMessage(error, useI18n().t('toasts.assignmentLetter.statusFailed'))
        this.errorMsg = detail
        useToastNotification().showError(useI18n().t('toasts.assignmentLetter.statusFailed'), detail)
      } finally {
        this.loading = false
      }
    }
  }
})