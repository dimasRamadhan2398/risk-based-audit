// Auth Types
export interface User {
  id: string
  username: string
  email: string
  fullName: string
  phone?: string
  department?: string
  /** Linked employee (users.employee_id = employees.employee_code). Filled once /auth/me returns it (spec B1). */
  employeeId?: string | null
  /** master-service department UUID of the linked employee (spec B1/B2). Filled by fetchEmployeeLink(). */
  departmentId?: string | null
  position?: string
  /** Profile picture as a data URL. Held in memory only, never written to the auth-user cookie. */
  avatarUrl?: string
  roles: string[]
  mustChangePassword?: boolean
  createdAt?: string
  updatedAt?: string
}

/** Role names as seeded by the backend (auth-service/pkg/database/seeders/seeder.go). */
export enum UserRole {
  ADMIN = 'ADMIN',
  AUDITOR = 'AUDITOR',
  EXECUTIVE = 'EXECUTIVE',
  AUDITEE = 'AUDITEE',
  VIEWER = 'VIEWER',
  DEPARTMENT_HEAD = 'DEPARTMENT_HEAD'
}

export interface LoginCredentials {
  username: string
  password: string
  rememberMe?: boolean
  deviceFingerprint?: string
  deviceName?: string
  deviceType?: string
}

export interface RegisterData {
  username: string
  email: string
  password: string
  fullName: string
}

export interface DeviceInfo {
  deviceFingerprint: string
  deviceName: string
  deviceType: string
}

export interface LoginResponse {
  token?: string
  expiresAt?: number
  user?: User
  mfaRequired: boolean
  mfaToken?: string
  isNewDevice: boolean
}

export interface MFAVerifyPayload {
  mfaToken: string
  code: string
  trustDevice: boolean
  deviceFingerprint?: string
  deviceName?: string
  deviceType?: string
}
