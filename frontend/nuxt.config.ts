import yaml from "@rollup/plugin-yaml";
import { fileURLToPath } from "node:url";

// https://nuxt.com/docs/api/configuration/nuxt-config
const isProd = process.env.NODE_ENV === 'production';
const defaultBaseUrl = isProd ? 'https://api.auditsphere.app/api/v1' : '/api/v1';
const defaultAnalyticsUrl = isProd ? 'https://api.auditsphere.app/api/analytics' : '/api/analytics';
const defaultPythonAiUrl = isProd ? 'https://api.auditsphere.app/api/python-ai' : 'http://localhost:8000';

export default defineNuxtConfig({
  devtools: {
    enabled: false,
  },
  telemetry: false,
  modules: [
    "@pinia/nuxt",
    "@nuxt/ui",
    "@nuxt/eslint",
    "@vueuse/nuxt",
    "@nuxt/image",
    "nuxt-charts",
    "nuxt-chatgpt",
  ],
  fonts: {
    providers: {
      google: false,
      bunny: false,
      fontshare: false,
      fontsource: false,
      adobe: false
    }
  },
  ssr: false,
  routeRules: {
    "/**": {
      ssr: false,
    },
    "/api/v1/**": {
      proxy: process.env.API_BASE_URL_SERVER || "http://localhost:8080/api/v1/**"
    },
    "/api/analytics/**": {
      proxy: process.env.ANALYTICS_API_BASE_URL_SERVER || "http://localhost:8080/api/analytics/**"
    }
  },
  alias: {
    "@": fileURLToPath(new URL("./", import.meta.url)),
    "@components": fileURLToPath(new URL("./components", import.meta.url)),
    "@composables": fileURLToPath(new URL("./composables", import.meta.url)),
    "@layouts": fileURLToPath(new URL("./layouts", import.meta.url)),
    "@pages": fileURLToPath(new URL("./pages", import.meta.url)),
    "@stores": fileURLToPath(new URL("./stores", import.meta.url)),
    "@utils": fileURLToPath(new URL("./utils", import.meta.url)),
    "@types": fileURLToPath(new URL("./types", import.meta.url)),
    "@assets": fileURLToPath(new URL("./assets", import.meta.url)),
  },
  typescript: {
    strict: true,
    typeCheck: false, // Disable during dev for better performance, use 'npm run type-check' instead
  },
  css: ["~/assets/css/main.css"],
  app: {
    head: {
      title: "Risk-Based Internal Audit System",
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1" },
        {
          name: "description",
          content: "Enterprise Risk Management and Internal Audit System",
        },
      ],
      link: [
        { rel: "icon", type: "image/png", href: "/favicon.png" },
        // Google Fonts
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300;400;500;600;700&display=swap",
        },
      ],
      script: [
        // Phosphor Icons
        { src: "https://unpkg.com/@phosphor-icons/web", defer: true },
        // CATATAN: Tailwind & Vue 3 CDN TIDAK boleh dimasukkan di sini
      ],
    },
  },
  eslint: {
    config: {
      stylistic: {
        commaDangle: "never",
        braceStyle: "1tbs",
      },
    },
  },
  runtimeConfig: {
    // Private keys (only available server-side)
    apiSecret: "",

    // Public keys (exposed to client)
    public: {
      apiBase: process.env.API_BASE_URL || defaultBaseUrl,
      analyticsApiBase: process.env.ANALYTICS_API_BASE_URL || defaultAnalyticsUrl,
      pythonAiBaseUrl: process.env.PYTHON_AI_BASE_URL || defaultPythonAiUrl,
      authServiceBaseUrl: process.env.NUXT_PUBLIC_AUTH_SERVICE_BASE_URL || defaultBaseUrl,
      auditServiceBaseUrl: process.env.NUXT_PUBLIC_AUDIT_SERVICE_BASE_URL || defaultBaseUrl,
      riskServiceBaseUrl: process.env.NUXT_PUBLIC_RISK_SERVICE_BASE_URL || defaultBaseUrl,
      masterServiceBaseUrl: process.env.NUXT_PUBLIC_MASTER_SERVICE_BASE_URL || defaultBaseUrl,
    },
  },

  compatibilityDate: "2026-01-23",
  future: {
    compatibilityVersion: 4,
  },
});