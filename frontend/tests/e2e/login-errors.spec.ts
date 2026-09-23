import { test, expect } from '@playwright/test'

test.describe('Login Page Error Scenarios', () => {
  test.beforeEach(async ({ context, page }) => {
    // Navigate to the login page before each test scenario
    await page.goto('/auth/login')
  })

  test('F-01-E1: Should show "Service Not Found (404)" error message', async ({ page }) => {
    // Mock API 404 response
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({
          error: {
            message: 'endpoint_not_found'
          }
        })
      })
    })

    // Fill in valid credentials formats to bypass client-side validation
    await page.fill('#username', 'testuser')
    await page.fill('#password', 'password123')

    // Click submit
    await page.click('#login-submit-btn')

    // Expect the translation of 404 error (either English or Indonesian)
    const errorAlert = page.locator('.bg-error-700')
    await expect(errorAlert).toBeVisible()
    await expect(errorAlert).toContainText(/(Authentication service not found|Layanan autentikasi tidak ditemukan)/)
  })

  test('F-01-E2: Should show "Server Unreachable (502/503/504)" error message', async ({ page }) => {
    // Mock API 502 response
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.fulfill({
        status: 502,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Bad Gateway'
        })
      })
    })

    await page.fill('#username', 'testuser')
    await page.fill('#password', 'password123')
    await page.click('#login-submit-btn')

    const errorAlert = page.locator('.bg-error-700')
    await expect(errorAlert).toBeVisible()
    await expect(errorAlert).toContainText(/(Server is currently unreachable|Server sedang tidak dapat diakses|Bad Gateway)/)
  })

  test('F-01-E3: Should show "Internal Server Error (500)" error message', async ({ page }) => {
    // Mock API 500 response
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Internal Server Error'
        })
      })
    })

    await page.fill('#username', 'testuser')
    await page.fill('#password', 'password123')
    await page.click('#login-submit-btn')

    const errorAlert = page.locator('.bg-error-700')
    await expect(errorAlert).toBeVisible()
    await expect(errorAlert).toContainText(/(An internal server error occurred|Terjadi kesalahan internal pada server|Internal Server Error)/)
  })

  test('F-01-E4: Should show "Invalid Credentials (401/403)" error message', async ({ page }) => {
    // Mock API 401 response
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          error: {
            message: 'invalid credentials'
          }
        })
      })
    })

    await page.fill('#username', 'testuser')
    await page.fill('#password', 'wrongpassword')
    await page.click('#login-submit-btn')

    const errorAlert = page.locator('.bg-error-700')
    await expect(errorAlert).toBeVisible()
    await expect(errorAlert).toContainText(/(Incorrect username or password|Username atau password salah)/)
  })

  test('F-01-E5: Should show "Too Many Requests (429)" error message', async ({ page }) => {
    // Mock API 429 response
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.fulfill({
        status: 429,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Too many requests'
        })
      })
    })

    await page.fill('#username', 'testuser')
    await page.fill('#password', 'password123')
    await page.click('#login-submit-btn')

    const errorAlert = page.locator('.bg-error-700')
    await expect(errorAlert).toBeVisible()
    await expect(errorAlert).toContainText(/(Too many login attempts|Terlalu banyak percobaan login|Too many requests)/)
  })

  test('F-01-E6: Should show "Network/Fetch Error" when server connection fails', async ({ page }) => {
    // Mock request failure (Network disconnected / Connection refused)
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.abort('failed')
    })

    await page.fill('#username', 'testuser')
    await page.fill('#password', 'password123')
    await page.click('#login-submit-btn')

    const errorAlert = page.locator('.bg-error-700')
    await expect(errorAlert).toBeVisible()
    await expect(errorAlert).toContainText(/(Failed to connect to the server|Gagal menghubungi server|no response|Load failed)/)
  })
})
