import { defineStore } from 'pinia'
import type { User, LoginCredentials, MFAVerifyPayload } from '~/types/auth'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  mfaRequired: boolean
  mfaToken: string | null
  isNewDevice: boolean
  needsConfidentialityAgreement: boolean
  _initialized: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    token: null,
    isAuthenticated: false,
    mfaRequired: false,
    mfaToken: null,
    isNewDevice: false,
    needsConfidentialityAgreement: false,
    _initialized: false,
  }),

  getters: {
    getUser: (state) => state.user,
    isLoggedIn: (state) => state.isAuthenticated,
    userRoles: (state) => state.user?.roles ?? [],
  },

  actions: {
    /** Core login against auth-service (POST /api/v1/auth/login) */
    async login(credentials: LoginCredentials) {
      const config = useRuntimeConfig()

      try {
        const response = await $fetch<any>(
          `${config.public.authServiceBaseUrl}/auth/login`,
          {
            method: 'POST',
            body: {
              username: credentials.username,
              password: credentials.password,
              device_fingerprint: credentials.deviceFingerprint,
              device_name: credentials.deviceName,
              device_type: credentials.deviceType,
            },
          },
        )

        const data = response.data ?? response

        // MFA required — store temp token and redirect
        if (data.mfa_required) {
          this.mfaRequired = true
          this.mfaToken = data.mfa_token
          this.isNewDevice = data.is_new_device ?? false
          return { mfaRequired: true }
        }

        // Full login success
        await this._persistSession(data, credentials.rememberMe)

        // F-05: Check if user needs to accept confidentiality agreement
        await this._checkConfidentialityAgreement()

        return data
      }
      catch (error: any) {
        const msg = error?.data?.error?.message || error?.data?.message || error?.message || 'Login failed'
        throw new Error(msg)
      }
    },

    /** F-03: Verify MFA OTP and complete login (POST /api/v1/auth/verify-mfa-login) */
    async verifyMFALogin(payload: MFAVerifyPayload) {
      const config = useRuntimeConfig()

      try {
        const response = await $fetch<any>(
          `${config.public.authServiceBaseUrl}/auth/verify-mfa-login`,
          {
            method: 'POST',
            body: {
              mfa_token: payload.mfaToken,
              code: payload.code,
              trust_device: payload.trustDevice,
              device_fingerprint: payload.deviceFingerprint,
              device_name: payload.deviceName,
              device_type: payload.deviceType,
            },
          },
        )

        const data = response.data ?? response

        this.mfaRequired = false
        this.mfaToken = null

        await this._persistSession(data, true)

        // F-05: Check if user needs to accept confidentiality agreement
        await this._checkConfidentialityAgreement()

        return data
      }
      catch (error: any) {
        const msg = error?.data?.error?.message || error?.data?.message || error?.message || 'MFA verification failed'
        throw new Error(msg)
      }
    },

    /** F-05: Check if user has accepted the system confidentiality agreement */
    async _checkConfidentialityAgreement() {
      if (!this.token) return
      const config = useRuntimeConfig()
      try {
        const response = await $fetch<any>(
          `${config.public.authServiceBaseUrl}/confidentiality/status`,
          {
            headers: { Authorization: `Bearer ${this.token}` },
          },
        )
        const data = response.data ?? response
        this.needsConfidentialityAgreement = !data.has_accepted
      }
      catch {
        // If endpoint doesn't exist yet, show agreement on first login
        this.needsConfidentialityAgreement = true
      }
    },

    /** F-05: Accept the confidentiality agreement */
    async acceptConfidentialityAgreement() {
      const config = useRuntimeConfig()
      try {
        await $fetch(
          `${config.public.authServiceBaseUrl}/confidentiality/accept`,
          {
            method: 'POST',
            headers: { Authorization: `Bearer ${this.token}` },
            body: {
              agreement_type: 'SYSTEM',
              title: 'Pakta Integritas Sistem Audit Internal Berbasis Risiko',
              content: 'Saya menyatakan bahwa saya memahami dan menyetujui ketentuan kerahasiaan sistem ini.',
              version: '1.0',
            },
          },
        )
      }
      catch {
        // Gracefully handle if endpoint not yet wired
      }
      finally {
        this.needsConfidentialityAgreement = false
      }
    },

    /** Fetch complete detailed user profile */
    async fetchUserProfile() {
      if (!this.token || !this.user?.id) return null
      const config = useRuntimeConfig()
      try {
        const response = await $fetch<any>(
          `${config.public.authServiceBaseUrl}/users/${this.user.id}`,
          {
            headers: { Authorization: `Bearer ${this.token}` },
          },
        )
        return response.data ?? response
      }
      catch (error) {
        console.error('Failed to fetch user profile:', error)
        return null
      }
    },

    /** Update user profile (PUT /api/v1/users/:id) */
    async updateProfile(profile: { fullName: string, phone: string, department: string }) {
      if (!this.token || !this.user?.id) return
      const config = useRuntimeConfig()
      try {
        await $fetch(
          `${config.public.authServiceBaseUrl}/users/${this.user.id}`,
          {
            method: 'PUT',
            headers: { Authorization: `Bearer ${this.token}` },
            body: {
              full_name: profile.fullName,
              phone: profile.phone,
              department: profile.department,
            },
          },
        )
        // Update local store state
        if (this.user) {
          this.user.fullName = profile.fullName
          this.user.phone = profile.phone
          this.user.department = profile.department

          // Re-cookie updated user
          const userCookie = useCookie('auth-user')
          userCookie.value = JSON.stringify(this.user)
        }
      }
      catch (error: any) {
        throw new Error(error?.data?.error?.message || error?.data?.message || 'Failed to update profile')
      }
    },

    /** Persist session data to store and cookies */
    async _persistSession(data: any, rememberMe?: boolean) {
      const maxAge = rememberMe ? 60 * 60 * 24 * 30 : 60 * 60 * 24

      const user: User = {
        id: data.user?.id ?? '',
        username: data.user?.username ?? '',
        email: data.user?.email ?? '',
        fullName: data.user?.full_name ?? data.user?.fullName ?? '',
        phone: data.user?.phone,
        department: data.user?.department,
        roles: data.user?.roles ?? [],
      }

      this.user = user
      this.token = data.token
      this.isAuthenticated = true
      this.isNewDevice = data.is_new_device ?? false

      const tokenCookie = useCookie('auth-token', { maxAge })
      tokenCookie.value = data.token

      const userCookie = useCookie('auth-user', { maxAge })
      userCookie.value = JSON.stringify(user)
    },

    /** Logout — F-06: triggers audit trail on backend */
    async logout() {
      const config = useRuntimeConfig()
      try {
        if (this.token) {
          await $fetch(`${config.public.authServiceBaseUrl}/auth/logout`, {
            method: 'POST',
            headers: { Authorization: `Bearer ${this.token}` },
          })
        }
      }
      catch {
        // Always clear local state
      }
      finally {
        this._clearState()
        await navigateTo('/auth/login')
      }
    },

    _clearState() {
      this.user = null
      this.token = null
      this.isAuthenticated = false
      this.mfaRequired = false
      this.mfaToken = null
      this.isNewDevice = false
      this.needsConfidentialityAgreement = false

      const tokenCookie = useCookie('auth-token')
      tokenCookie.value = null
      const userCookie = useCookie('auth-user')
      userCookie.value = null
    },

    /** Restore session from cookies on app init */
    async fetchUser() {
      const tokenCookie = useCookie('auth-token')
      const userCookie = useCookie('auth-user')

      if (!tokenCookie.value || !userCookie.value) {
        this._initialized = true
        return
      }

      try {
        const user = JSON.parse(userCookie.value)
        this.user = user
        this.token = tokenCookie.value
        this.isAuthenticated = true
        this._initialized = true
      }
      catch {
        this._initialized = true
        this._clearState()
      }
    },

    async fetchTrustedDevices() {
      if (!this.token) return []
      const config = useRuntimeConfig()
      try {
        const response = await $fetch<any>(
          `${config.public.authServiceBaseUrl}/devices`,
          {
            headers: { Authorization: `Bearer ${this.token}` },
          },
        )
        return response.data ?? response
      }
      catch (error) {
        console.error('Failed to fetch trusted devices:', error)
        return []
      }
    },

    async unenrollDevice(deviceId: string) {
      if (!this.token) return
      const config = useRuntimeConfig()
      try {
        await $fetch(
          `${config.public.authServiceBaseUrl}/devices/${deviceId}`,
          {
            method: 'DELETE',
            headers: { Authorization: `Bearer ${this.token}` },
          },
        )
      }
      catch (error: any) {
        throw new Error(error?.data?.error?.message || error?.data?.message || 'Failed to remove device')
      }
    },

    async forgotPassword(email: string) {
      const config = useRuntimeConfig()
      try {
        await $fetch(`${config.public.authServiceBaseUrl}/auth/forgot-password`, {
          method: 'POST',
          body: { email },
        })
      }
      catch (error: any) {
        throw new Error(error?.data?.error?.message || error?.data?.message || 'Failed to send reset email')
      }
    },

    async resetPassword(token: string, password: string) {
      const config = useRuntimeConfig()
      try {
        await $fetch(`${config.public.authServiceBaseUrl}/auth/reset-password`, {
          method: 'POST',
          body: { token, password },
        })
      }
      catch (error: any) {
        throw new Error(error?.data?.error?.message || error?.data?.message || 'Failed to reset password')
      }
    },
  },
})
