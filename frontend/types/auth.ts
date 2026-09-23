// Auth Types
export interface User {
  id: string
  username: string
  email: string
  fullName: string
  phone?: string
  department?: string
  position?: string
  roles: string[]
  createdAt?: string
  updatedAt?: string
}

export enum UserRole {
  ADMIN = 'admin',
  AUDITOR = 'auditor',
  DEPARTMENT_HEAD = 'department_head',
  AUDITEE = 'auditee',
  VIEWER = 'viewer',
  AUDIT_STAFF = 'audit_staff',
  AUDIT_MANAGER = 'audit_manager',
  CHIEF_AUDIT_EXECUTIVE = 'chief_audit_executive',
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

