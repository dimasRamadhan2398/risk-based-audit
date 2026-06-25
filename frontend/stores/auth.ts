import { defineStore } from "pinia";
import {
  type User,
  type LoginCredentials,
  type RegisterData,
  UserRole,
} from "~/types/auth";

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  mfaRequired: boolean;
  mfaToken: string | null;
  isNewDevice: boolean;
  _initialized: boolean;
}

export const useAuthStore = defineStore("auth", {
  state: (): AuthState => ({
    user: null,
    token: null,
    isAuthenticated: false,
    mfaRequired: false,
    mfaToken: null,
    isNewDevice: false,
    _initialized: false,
  }),

  getters: {
    getUser: (state) => state.user,
    isLoggedIn: (state) => state.isAuthenticated,
  },

  actions: {
    async loginDummy(credentials: LoginCredentials) {
      const config = useRuntimeConfig();

      try {
        // const response = await $fetch<{ user: User; token: string }>(
        //   `${config.public.apiBase}/auth/login`,
        //   {
        //     method: "POST",
        //     body: credentials,
        //   },
        // );

        this.user = {
          id: "1",
          email: "x6H8K@example.com",
          fullName: "John Doe",
          role: UserRole.ADMIN,
          department: "IT",
          createdAt: "2022-01-01",
          updatedAt: "2022-01-01",
        };
        this.token =
          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30";
        this.isAuthenticated = true;

        // Store token and user in cookies
        const tokenCookie = useCookie("auth-token", {
          maxAge: credentials.rememberMe ? 60 * 60 * 24 * 30 : 60 * 60 * 24, // 30 days or 1 day
        });
        tokenCookie.value = this.token;

        const userCookie = useCookie("auth-user", {
          maxAge: credentials.rememberMe ? 60 * 60 * 24 * 30 : 60 * 60 * 24,
        });
        userCookie.value = JSON.stringify(this.user);
      } catch (error: any) {
        throw new Error(error.data?.message || "Login failed");
      }
    },
    async login(credentials: LoginCredentials) {
      const config = useRuntimeConfig();

      try {
        const response = await $fetch<any>(
          `${config.public.apiBase}/auth/login`,
          {
            method: "POST",
            body: credentials,
          },
        );

        const data = response.data || response;

        if (data.mfa_required) {
          this.mfaRequired = true;
          this.mfaToken = data.mfa_token;
          this.isNewDevice = data.is_new_device;
          return { mfaRequired: true };
        }

        this.user = data.user;
        this.token = data.token;
        this.isAuthenticated = true;
        this.isNewDevice = data.is_new_device;

        // Store token in cookie
        const tokenCookie = useCookie("auth-token", {
          maxAge: credentials.rememberMe ? 60 * 60 * 24 * 30 : 60 * 60 * 24, // 30 days or 1 day
        });
        tokenCookie.value = data.token;

        const userCookie = useCookie("auth-user", {
          maxAge: credentials.rememberMe ? 60 * 60 * 24 * 30 : 60 * 60 * 24,
        });
        userCookie.value = JSON.stringify(data.user);

        return data;
      } catch (error: any) {
        throw new Error(error.data?.message || "Login failed");
      }
    },

    async verifyMFALogin(code: string, trustDevice: boolean = false) {
      const config = useRuntimeConfig();

      // Basic device info (ideally use a library for fingerprinting)
      const deviceInfo = {
        device_fingerprint: btoa(navigator.userAgent), // Simple fingerprint for now
        device_name: navigator.platform,
        device_type: "Web Browser",
      };

      try {
        const response = await $fetch<any>(
          `${config.public.apiBase}/auth/verify-mfa-login`,
          {
            method: "POST",
            body: {
              mfa_token: this.mfaToken,
              code: code,
              trust_device: trustDevice,
              ...deviceInfo
            },
          },
        );

        const data = response.data || response;

        this.user = data.user;
        this.token = data.token;
        this.isAuthenticated = true;
        this.mfaRequired = false;
        this.mfaToken = null;

        const tokenCookie = useCookie("auth-token", {
          maxAge: 60 * 60 * 24 * 30, // 30 days
        });
        tokenCookie.value = data.token;

        const userCookie = useCookie("auth-user", {
          maxAge: 60 * 60 * 24 * 30,
        });
        userCookie.value = JSON.stringify(data.user);

        return data;
      } catch (error: any) {
        throw new Error(error.data?.message || "MFA verification failed");
      }
    },

    async register(data: RegisterData) {
      const config = useRuntimeConfig();

      try {
        const response = await $fetch<{ user: User; token: string }>(
          `${config.public.apiBase}/auth/register`,
          {
            method: "POST",
            body: data,
          },
        );

        this.user = response.user;
        this.token = response.token;
        this.isAuthenticated = true;

        const tokenCookie = useCookie("auth-token");
        tokenCookie.value = response.token;

        const userCookie = useCookie("auth-user");
        userCookie.value = JSON.stringify(response.user);

        return response;
      } catch (error: any) {
        throw new Error(error.data?.message || "Registration failed");
      }
    },

    async logout() {
      this.user = null;
      this.token = null;
      this.isAuthenticated = false;

      const tokenCookie = useCookie("auth-token");
      tokenCookie.value = null;

      const userCookie = useCookie("auth-user");
      userCookie.value = null;

      await navigateTo("/auth/login");
    },

    async fetchUserDummy() {
      this.user = {
        id: "1",
        email: "x6H8K@example.com",
        fullName: "John Doe",
        role: UserRole.ADMIN,
        department: "IT",
        createdAt: "2022-01-01",
        updatedAt: "2022-01-01",
      };
      this.token =
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30";
      this.isAuthenticated = true;
    },
    async fetchUser() {
      const tokenCookie = useCookie("auth-token");
      const userCookie = useCookie("auth-user");

      if (!tokenCookie.value || !userCookie.value) {
        this._initialized = true;
        return;
      }

      try {
        this.user = JSON.parse(userCookie.value);
        this.token = tokenCookie.value;
        this.isAuthenticated = true;
        this._initialized = true;
      } catch (error) {
        this._initialized = true;
        this.logout();
      }
    },

    async forgotPassword(email: string) {
      const config = useRuntimeConfig();

      try {
        await $fetch(`${config.public.apiBase}/auth/forgot-password`, {
          method: "POST",
          body: { email },
        });
      } catch (error: any) {
        throw new Error(error.data?.message || "Failed to send reset email");
      }
    },

    async resetPassword(token: string, password: string) {
      const config = useRuntimeConfig();

      try {
        await $fetch(`${config.public.apiBase}/auth/reset-password`, {
          method: "POST",
          body: { token, password },
        });
      } catch (error: any) {
        throw new Error(error.data?.message || "Failed to reset password");
      }
    },
  },
});
