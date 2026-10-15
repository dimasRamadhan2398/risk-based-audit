import { defineStore } from 'pinia'
import type { User, LoginCredentials, MFAVerifyPayload } from '~/types/auth'
import { extractErrorMessage } from '~/utils/error'
import { getAuthServiceBaseUrl } from '~/composables/useApiUrl'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  mfaRequired: boolean
  mfaToken: string | null
  isNewDevice: boolean
  needsConfidentialityAgreement: boolean
  _initialized: boolean
  _profileHydrated: boolean
}

/** Cookies are limited to ~4 KB; the avatar data URL must stay out of them */
const serializeUserForCookie = (user: User | null) => {
  const { avatarUrl: _avatarUrl, ...rest } = (user ?? {}) as User & { avatar_url?: string }
  delete (rest as { avatar_url?: string }).avatar_url
  return JSON.stringify(rest)
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
    _profileHydrated: false,
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
          `${getAuthServiceBaseUrl()}/auth/login`,
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
        const msg = extractErrorMessage(error, 'Login failed')
        const err = new Error(msg) as any
        err.status = error?.status ?? error?.statusCode ?? error?.response?.status
        err.data = error?.data
        throw err
      }
    },

    /** F-03: Verify MFA OTP and complete login (POST /api/v1/auth/verify-mfa-login) */
    async verifyMFALogin(payload: MFAVerifyPayload) {
      const config = useRuntimeConfig()

      try {
        const response = await $fetch<any>(
          `${getAuthServiceBaseUrl()}/auth/verify-mfa-login`,
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
        const msg = extractErrorMessage(error, 'MFA verification failed')
        throw new Error(msg)
      }
    },

    /** F-05: Check if user has accepted the system confidentiality agreement */
    async _checkConfidentialityAgreement() {
      if (!this.token) return
      const config = useRuntimeConfig()
      try {
        const response = await $fetch<any>(
          `${getAuthServiceBaseUrl()}/confidentiality/status`,
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
          `${getAuthServiceBaseUrl()}/confidentiality/accept`,
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
          `${getAuthServiceBaseUrl()}/users/${this.user.id}`,
          {
            headers: { Authorization: `Bearer ${this.token}` },
          },
        )
        const profile = response.data ?? response
        // Hydrates the avatar (kept out of the cookie) after a reload
        if (this.user && profile && profile.avatar_url !== undefined) {
          this.user.avatarUrl = profile.avatar_url || ''
        }
        return profile
      }
      catch (error) {
        console.error('Failed to fetch user profile:', error)
        return null
      }
    },

    /** Load the avatar once per session after the cookie restore (login already returns it) */
    async hydrateProfile() {
      if (this._profileHydrated || !this.token || !this.user?.id) return
      this._profileHydrated = true
      await this.fetchUserProfile()
    },

    /**
     * Update user profile (PUT /api/v1/users/:id).
     * `avatarUrl`: undefined = unchanged (field omitted), '' = remove, data URL = replace.
     */
    async updateProfile(profile: { fullName: string, phone: string, department: string, position?: string, avatarUrl?: string }) {
      if (!this.token || !this.user?.id) return
      try {
        const response = await $fetch<any>(
          `${getAuthServiceBaseUrl()}/users/${this.user.id}`,
          {
            method: 'PUT',
            headers: { Authorization: `Bearer ${this.token}` },
            body: {
              full_name: profile.fullName,
              phone: profile.phone,
              department: profile.department,
              position: profile.position,
              ...(profile.avatarUrl !== undefined ? { avatar_url: profile.avatarUrl } : {}),
            },
          },
        )

        // Apply the server's view of the user; fall back to what we sent
        const saved = response?.data ?? response ?? {}
        if (this.user) {
          this.user.fullName = saved.full_name ?? profile.fullName
          this.user.phone = saved.phone ?? profile.phone
          this.user.department = saved.department ?? profile.department
          this.user.position = saved.position ?? profile.position
          if (saved.avatar_url !== undefined) {
            this.user.avatarUrl = saved.avatar_url || ''
          }
          else if (profile.avatarUrl !== undefined) {
            this.user.avatarUrl = profile.avatarUrl
          }

          // Re-cookie updated user (without the avatar)
          const userCookie = useCookie('auth-user')
          userCookie.value = serializeUserForCookie(this.user)
        }
      }
      catch (error: any) {
        // Keep status/data so callers can map it with getUserErrorMessage
        const err = new Error(extractErrorMessage(error, 'Failed to update profile')) as any
        err.status = error?.status ?? error?.statusCode ?? error?.response?.status
        err.data = error?.data
        throw err
      }
    },

    /** Change own password (POST /api/v1/auth/change-password) */
    async changePassword(oldPassword: string, newPassword: string) {
      try {
        await $fetch(`${getAuthServiceBaseUrl()}/auth/change-password`, {
          method: 'POST',
          headers: { Authorization: `Bearer ${this.token}` },
          body: { old_password: oldPassword, new_password: newPassword },
        })
      }
      catch (error: any) {
        throw new Error(extractErrorMessage(error, 'Failed to change password'))
      }

      // An admin-issued temporary password has now been replaced
      if (this.user?.mustChangePassword) {
        this.user.mustChangePassword = false
        const userCookie = useCookie('auth-user')
        userCookie.value = serializeUserForCookie(this.user)
      }
    },

    /** Persist session data to store and cookies */
    async _persistSession(data: any, rememberMe?: boolean) {
      const maxAge = rememberMe ? 60 * 60 * 24 * 30 : 60 * 60 * 24

      const user: User = {
        ...data.user,
        id: data.user?.id ?? '',
        email: data.user?.email ?? '',
        fullName: data.user?.full_name ?? data.user?.fullName ?? '',
        ...(data.user?.username !== undefined ? { username: data.user.username } : {}),
        ...(data.user?.phone !== undefined ? { phone: data.user.phone } : {}),
        ...(data.user?.department !== undefined ? { department: data.user.department } : {}),
        ...(data.user?.position !== undefined ? { position: data.user.position } : {}),
        ...(data.user?.roles !== undefined ? { roles: data.user.roles } : {}),
        mustChangePassword: data.user?.must_change_password ?? false,
        avatarUrl: data.user?.avatar_url ?? data.user?.avatarUrl ?? '',
      }
      delete (user as { avatar_url?: string }).avatar_url

      this.user = user
      this._profileHydrated = true
      this.token = data.token
      this.isAuthenticated = true
      this.isNewDevice = data.is_new_device ?? false

      const tokenCookie = useCookie('auth-token', { maxAge })
      tokenCookie.value = data.token

      const userCookie = useCookie('auth-user', { maxAge })
      userCookie.value = serializeUserForCookie(user)
    },

    /** Logout — F-06: triggers audit trail on backend */
    async logout() {
      const config = useRuntimeConfig()
      try {
        if (this.token) {
          await $fetch(`${getAuthServiceBaseUrl()}/auth/logout`, {
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
      this._profileHydrated = false

      const tokenCookie = useCookie('auth-token')
      tokenCookie.value = null
      const userCookie = useCookie('auth-user')
      userCookie.value = null
    },

    /** Restore session from cookies on app init */
    async fetchUser() {
      const tokenCookie = useCookie<string | null>('auth-token')
      const userCookie = useCookie<any>('auth-user')

      console.log('[AuthStore] fetchUser() cookies read: token =', tokenCookie.value ? 'present' : 'missing', 'user =', typeof userCookie.value, userCookie.value)

      if (!tokenCookie.value || !userCookie.value) {
        console.log('[AuthStore] fetchUser() failed: one or both cookies are missing.')
        this._initialized = true
        return
      }

      try {
        let user = userCookie.value
        if (typeof user === 'string') {
          user = JSON.parse(user)
        }
        this.user = user
        this.token = tokenCookie.value
        this.isAuthenticated = true
        this._initialized = true
        console.log('[AuthStore] Session successfully restored for user:', this.user?.username)
      }
      catch (err) {
        console.error('[AuthStore] Failed to restore session from cookie:', err)
        this._initialized = true
        this._clearState()
      }
    },

    async fetchTrustedDevices() {
      if (!this.token) return []
      const config = useRuntimeConfig()
      try {
        const response = await $fetch<any>(
          `${getAuthServiceBaseUrl()}/devices`,
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
          `${getAuthServiceBaseUrl()}/devices/${deviceId}`,
          {
            method: 'DELETE',
            headers: { Authorization: `Bearer ${this.token}` },
          },
        )
      }
      catch (error: any) {
        throw new Error(extractErrorMessage(error, 'Failed to remove device'))
      }
    },

    async forgotPassword(email: string) {
      const config = useRuntimeConfig()
      try {
        await $fetch(`${getAuthServiceBaseUrl()}/auth/forgot-password`, {
          method: 'POST',
          body: { email },
        })
      }
      catch (error: any) {
        throw new Error(extractErrorMessage(error, 'Failed to send reset email'))
      }
    },

    async resetPassword(token: string, password: string) {
      const config = useRuntimeConfig()
      try {
        await $fetch(`${getAuthServiceBaseUrl()}/auth/reset-password`, {
          method: 'POST',
          body: { token, password },
        })
      }
      catch (error: any) {
        throw new Error(extractErrorMessage(error, 'Failed to reset password'))
      }
    },
  },
})
