import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: ['./tests/setup.ts'],
    // tests/e2e holds Playwright specs (npm run test:e2e); vitest cannot load them.
    exclude: ['**/node_modules/**', '**/.nuxt/**', '**/.output/**', 'tests/e2e/**', '.claude/**'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'tests/',
        '*.config.ts',
        '.nuxt/',
        '.output/'
      ]
    }
  },
  resolve: {
    alias: {
      '~': resolve(__dirname, './'),
      '@': resolve(__dirname, './'),
      '#imports': resolve(__dirname, './tests/mocks/imports.ts'),
      '#components': resolve(__dirname, './tests/mocks/components.ts')
    }
  }
})
