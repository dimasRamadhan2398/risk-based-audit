// Master Service Types - Employee, Department & Company

// ============= Company Types =============

export type CompanyType = 'HOLDING' | 'SUBSIDIARY' | 'BRANCH'

export interface Company {
  id: string
  company_code: string
  company_name: string
  legal_name?: string
  tax_id?: string
  company_type: CompanyType
  parent_id?: string
  location_id?: string
  phone?: string
  email?: string
  website?: string
  is_active: boolean
  established_at?: string
  created_at: string
  updated_at: string
}

export interface CompanyFormState {
  company_code: string
  company_name: string
  legal_name?: string
  tax_id?: string
  company_type: CompanyType
  parent_id?: string
  location_id?: string
  phone?: string
  email?: string
  website?: string
  is_active: boolean
  established_at?: string
}

export interface CreateCompanyRequest {
  company_code: string
  company_name: string
  legal_name?: string
  tax_id?: string
  company_type: CompanyType
  parent_id?: string
  location_id?: string
  phone?: string
  email?: string
  website?: string
  is_active: boolean
  established_at?: string
}

export interface UpdateCompanyRequest {
  company_name?: string
  legal_name?: string
  tax_id?: string
  company_type?: CompanyType
  parent_id?: string
  location_id?: string
  phone?: string
  email?: string
  website?: string
  is_active?: boolean
  established_at?: string
}

export interface ListCompaniesResponse {
  companies: Company[]
  pagination: PaginationMeta
}

// ============= Employee Types =============

export interface Employee {
  id: string
  employee_code: string
  full_name: string
  email: string
  phone: string
  company_id: string
  department_id: string
  job_role_id: string
  level_grade: number
  work_location_id?: string
  residence_address?: string
  residence_city?: string
  residence_province?: string
  residence_postal_code?: string
  manager_id?: string
  is_active: boolean
  join_date: string
  created_at: string
  updated_at: string
}

export interface EmployeeFormState {
  employee_code: string
  full_name: string
  email: string
  phone: string
  company_id: string
  department_id: string
  job_role_id: string
  level_grade: number
  work_location_id?: string
  residence_address?: string
  residence_city?: string
  residence_province?: string
  residence_postal_code?: string
  manager_id?: string
  is_active: boolean
  join_date: string
}

export interface CreateEmployeeRequest {
  employee_code: string
  full_name: string
  email: string
  phone?: string
  company_id: string
  department_id: string
  job_role_id: string
  level_grade: number
  work_location_id?: string
  residence_address?: string
  residence_city?: string
  residence_province?: string
  residence_postal_code?: string
  manager_id?: string
  is_active: boolean
  join_date: string
}

export interface UpdateEmployeeRequest {
  full_name?: string
  email?: string
  phone?: string
  department_id?: string
  job_role_id?: string
  level_grade?: number
  work_location_id?: string
  residence_address?: string
  residence_city?: string
  residence_province?: string
  residence_postal_code?: string
  manager_id?: string
  is_active?: boolean
}

// ============= Department Types =============

export interface Department {
  id: string
  department_code: string
  department_name: string
  department_description?: string
  pic_id: string
  level: number
  company_id: string
  business_unit_id?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface DepartmentFormState {
  department_code: string
  department_name: string
  department_description?: string
  pic_id: string
  level: number
  company_id: string
  business_unit_id?: string
  is_active: boolean
}

export interface CreateDepartmentRequest {
  department_code: string
  department_name: string
  department_description?: string
  pic_id: string
  level: number
  company_id: string
  business_unit_id?: string
  is_active: boolean
}

export interface UpdateDepartmentRequest {
  department_name?: string
  department_description?: string
  pic_id?: string
  level?: number
  company_id?: string
  business_unit_id?: string
  is_active?: boolean
}

// ============= Pagination Types =============

export interface PaginationMeta {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: PaginationMeta
}

export interface ListEmployeesResponse {
  employees: Employee[]
  pagination: PaginationMeta
}

export interface ListDepartmentsResponse {
  departments: Department[]
  pagination: PaginationMeta
}

// ============= API Response Types =============

export interface ApiResponse<T> {
  success: boolean
  message: string
  data: T
}

export interface ApiError {
  code: string
  message: string
  status: number
}

// ============= Table Column Types =============

import type { TableColumn } from '@nuxt/ui'

export type EmployeeTableColumn = TableColumn<Employee>
export type DepartmentTableColumn = TableColumn<Department>
export type CompanyTableColumn = TableColumn<Company>

// ============= Vision, Mission & Goals Types =============
export type VisionMissionGoalsStatus = 'Draft' | 'In Review' | 'Approved' | 'Published'

export interface VmgGoal {
  id?: string
  vmg_id?: string
  goal_code: string
  goal_name: string
  goal_description?: string
  strategic_objective?: string
  kpi?: string
  target?: string
  unit?: string
  baseline_year?: string
  baseline_value?: string
  created_at?: string
  updated_at?: string
}

export interface VisionMissionGoals {
  id: string
  company_id: string
  company?: Company
  period: string
  effective_date?: string
  vision: string
  mission: string
  version?: string
  status: VisionMissionGoalsStatus
  notes?: string
  created_at?: string
  updated_at?: string
  created_by?: string
  modified_by?: string
  goals?: VmgGoal[]
}
