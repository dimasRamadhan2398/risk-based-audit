// Master Options Store - shared dropdown sources for master data forms
// Centralises the reference lists (company, department, job role, location,
// employee) so forms can bind readable names instead of raw UUIDs.
import type { Company, Department, Employee, JobRole, Location } from '~/types/master'

export interface SelectOption {
  label: string
  value: string
}

export const useMasterOptionsStore = defineStore('master-options', () => {
  // ============= State =============
  const companies = ref<Company[]>([])
  const departments = ref<Department[]>([])
  const jobRoles = ref<JobRole[]>([])
  const locations = ref<Location[]>([])
  const employees = ref<Employee[]>([])

  const loading = ref(false)
  const loaded = ref(false)

  // ============= Getters =============

  /**
   * The API exposes company code/name as either `company_code`/`company_name`
   * or the shorter `code`/`name`, depending on the service that serves it.
   */
  const companyOptions = computed<SelectOption[]>(() =>
    companies.value.map((c: any) => ({
      label: c.company_name || c.name || c.company_code || c.code || c.id,
      value: c.id
    }))
  )

  const departmentOptions = computed<SelectOption[]>(() =>
    departments.value.map((d) => ({
      label: d.department_code ? `${d.department_name} (${d.department_code})` : d.department_name,
      value: d.id
    }))
  )

  const jobRoleOptions = computed<SelectOption[]>(() =>
    jobRoles.value.map((r) => ({
      label: r.job_position_type ? `${r.job_role_name} — ${r.job_position_type}` : r.job_role_name,
      value: r.id
    }))
  )

  const locationOptions = computed<SelectOption[]>(() =>
    locations.value.map((l) => ({
      label: l.city ? `${l.name} — ${l.city}` : l.name,
      value: l.id
    }))
  )

  const employeeOptions = computed<SelectOption[]>(() =>
    employees.value.map((e) => ({
      label: e.employee_code ? `${e.full_name} (${e.employee_code})` : e.full_name,
      value: e.id
    }))
  )

  // ============= Actions =============

  /**
   * Load every reference list in parallel. Safe to call on each modal open —
   * it only refetches when the cache is empty or `force` is set.
   */
  const fetchAll = async (force = false) => {
    if (loaded.value && !force) return
    if (loading.value) return

    loading.value = true
    try {
      const companyApi = useCompanyApi()
      const departmentApi = useDepartmentApi()
      const jobRoleApi = useJobRoleApi()
      const locationApi = useLocationApi()
      const employeeApi = useEmployeeApi()

      const [companyList, departmentList, jobRoleList, locationList, employeeList] =
        await Promise.all([
          companyApi.getAllCompanies().catch(() => []),
          departmentApi.getAllDepartments().catch(() => []),
          jobRoleApi.getAllJobRoles().catch(() => []),
          locationApi.getAllLocations().catch(() => []),
          employeeApi.getAllEmployees().catch(() => [])
        ])

      companies.value = companyList
      departments.value = departmentList
      jobRoles.value = jobRoleList
      locations.value = locationList
      employees.value = employeeList
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  /**
   * Resolve a stored UUID back to its human readable label, so an existing
   * value still renders even when the reference list has not loaded yet.
   */
  const labelFor = (options: SelectOption[], id?: string | null): string => {
    if (!id) return ''
    return options.find((o) => o.value === id)?.label || id
  }

  return {
    // State
    companies,
    departments,
    jobRoles,
    locations,
    employees,
    loading,
    loaded,

    // Getters
    companyOptions,
    departmentOptions,
    jobRoleOptions,
    locationOptions,
    employeeOptions,

    // Actions
    fetchAll,
    labelFor
  }
})
