import { expect as baseExpect, test, type BrowserContext, type Page } from '@playwright/test'

const expect = baseExpect.configure({ timeout: 15_000 })
test.setTimeout(60_000)

const user = {
  id: 'user-e2e',
  username: 'executive.e2e',
  email: 'executive@example.test',
  fullName: 'Executive E2E',
  roles: ['admin', 'chief_audit_executive']
}

const summaryFixture = {
  id: '11111111-1111-1111-1111-111111111111',
  quarter: 1,
  periodeBulan: 'Januari',
  tahun: 2026,
  nomorDokumen: 'LHA-TEST-001',
  dokumenPath: 'executive-summary-test.pdf',
  status: 'Draft',
  executiveNote: '',
  narrative: 'Ringkasan pengujian checklist minor.',
  jumlahLaporan: 1,
  risikoTinggi: 1,
  risikoSedang: 1,
  risikoRendah: 0,
  jumlahRekomendasi: 2,
  followUpTable: JSON.stringify([
    { status: 'Closed', jumlah: 3, persentase: 50, keterangan: 'Selesai' },
    { status: 'In Progress', jumlah: 2, persentase: 33.3, keterangan: 'Berjalan' },
    { status: 'Overdue', jumlah: 1, persentase: 16.7, keterangan: 'Terlambat' }
  ]),
  topFindings: '[]',
  matriksKompilasi: '[]',
  akarMasalah: 'Akar masalah pengujian.',
  kesimpulan: 'Kesimpulan pengujian.',
  signatureTempat: 'Jakarta',
  signatureTanggal: '2026-09-24',
  signatureNamaKepala: 'Executive E2E',
  signatureNIK: 'E2E-001'
}

const annualPlanFixture = {
  id: 'annual-plan-e2e',
  code: 'PKAT-2026-ASR-007',
  version: 'v1.0',
  status: 'Work In Progress',
  year: '2026',
  selectedMonths: [0, 1],
  auditorCount: 2,
  daysPerAuditor: 5,
  supervisorId: 'S01',
  notes: 'Annual plan E2E',
  isActive: true,
  activities: [{
    name: 'Audit E2E',
    category: 'Assurance',
    department: 'IT',
    riskName: 'Risiko E2E',
    riskLevel: 'High'
  }]
}

async function authenticate(context: BrowserContext) {
  await context.addCookies([
    { name: 'auth-token', value: 'e2e-token', domain: 'localhost', path: '/' },
    { name: 'auth-user', value: encodeURIComponent(JSON.stringify(user)), domain: 'localhost', path: '/' }
  ])
}

async function mockApis(page: Page) {
  let summary = { ...summaryFixture }
  const writes: Array<Record<string, unknown>> = []

  await page.route('**/api/v1/**', async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    const method = request.method()

    if (url.pathname === '/api/v1/executive-summaries' && method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { items: [summary] } })
      })
      return
    }

    if (url.pathname.startsWith('/api/v1/executive-summaries/') && method === 'PUT') {
      const body = request.postDataJSON() as Record<string, unknown>
      writes.push(body)
      summary = { ...summary, ...body } as typeof summary
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: summary })
      })
      return
    }

    if (url.pathname === '/api/v1/annual-audit-plans' && method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { items: [annualPlanFixture] } })
      })
      return
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { items: [] } })
    })
  })

  return writes
}

async function openFirstExecutiveSummary(page: Page) {
  const heading = page.getByRole('heading', { name: summaryFixture.nomorDokumen, exact: true })
  await expect(heading).toBeVisible()
  const card = heading.locator('xpath=ancestor::div[contains(@class, "group")][1]')
  await card.getByRole('button').first().click()
}

test.beforeEach(async ({ context }) => {
  await authenticate(context)
})

test('activity code is generated and disabled in add and edit Annual Audit forms', async ({ page }) => {
  await mockApis(page)
  await page.goto('/annual-audit')
  await expect(page.getByText('PKAT-2026-ASR-007', { exact: true })).toBeVisible()

  await page.getByRole('button', { name: /New Audit Plan|Rencana Audit Baru/ }).click()
  const addCode = page.locator('input[placeholder="Terisi otomatis"]')
  await expect(addCode).toBeDisabled()
  await expect(addCode).toHaveValue(/^PKAT-2026-ASR-\d{3}$/)

  await page.reload()
  await expect(page.getByText('PKAT-2026-ASR-007', { exact: true })).toBeVisible()
  await page.locator('button[title="Edit"]').first().click()
  const editCode = page.locator('input[placeholder="Terisi otomatis"]')
  await expect(editCode).toBeDisabled()
  await expect(editCode).toHaveValue('PKAT-2026-ASR-007')
})

test('individual summary list uses handling and status labels without workflow buttons or period', async ({ page }) => {
  await mockApis(page)
  await page.goto('/executive-summary')

  await expect(page.getByText('Jumlah Penanganan / Handling', { exact: true }).locator('..')).toContainText('6')
  await expect(page.getByText('Status: Draft', { exact: true }).first()).toBeVisible()
  await expect(page.getByText('Ditolak (Rejected)', { exact: true })).toHaveCount(0)
  await expect(page.getByText('Periode Audit:', { exact: false })).toHaveCount(0)
  await expect(page.getByRole('button', { name: /Approve|Setujui/ })).toHaveCount(0)
})

test('individual view contains notes and the only Approve action, with removed sections absent', async ({ page }) => {
  const writes = await mockApis(page)
  await page.goto('/executive-summary')
  await openFirstExecutiveSummary(page)

  await expect(page.getByText('Detail Executive Summary Individual')).toBeVisible()
  await expect(page.getByText('Noted dari Executive untuk Auditor', { exact: true })).toBeVisible()
  await expect(page.getByText('Noted', { exact: true }).first()).toBeVisible()
  await expect(page.getByRole('button', { name: /Setujui.*Approve/ })).toBeVisible()
  await expect(page.getByText(/Section II: Statistik Kompilasi/)).toHaveCount(0)
  await expect(page.getByText(/Tren & Grafik/)).toHaveCount(0)
  await expect(page.getByText('Tahun Laporan', { exact: true })).toHaveCount(0)
  await expect(page.getByText('Periode Bulan', { exact: true })).toHaveCount(0)

  await page.locator('textarea[placeholder="Tuliskan catatan untuk Auditor..."]').fill('Mohon tindak lanjuti temuan prioritas.')
  await page.getByRole('button', { name: 'Simpan Catatan' }).click()
  await expect.poll(() => writes.some(write => write.executiveNote === 'Mohon tindak lanjuti temuan prioritas.')).toBe(true)

  await page.getByRole('button', { name: /Setujui.*Approve/ }).click()
  await expect.poll(() => writes.some(write => write.status === 'Approved')).toBe(true)
})

test('individual create and edit modes do not expose notes or Approve', async ({ page }) => {
  await mockApis(page)
  await page.goto('/executive-summary')
  await page.getByRole('button', { name: 'Buat Executive Summary Baru' }).click()

  await expect(page.getByText('Noted dari Executive untuk Auditor', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('button', { name: /Approve|Setujui/ })).toHaveCount(0)

  await page.getByRole('button', { name: 'Tutup' }).first().click()
  const heading = page.getByRole('heading', { name: summaryFixture.nomorDokumen, exact: true })
  const card = heading.locator('xpath=ancestor::div[contains(@class, "group")][1]')
  await card.getByRole('button').nth(1).click()
  await expect(page.getByText('Edit Executive Summary Individual')).toBeVisible()
  await expect(page.getByText('Noted dari Executive untuk Auditor', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('button', { name: /Approve|Setujui/ })).toHaveCount(0)
})

test('compilation uses quarters and removes charts and matrix while keeping notes in View', async ({ page }) => {
  await mockApis(page)
  await page.goto('/executive-summary-compilation')

  await expect(page.getByText('Jumlah Penanganan / Handling', { exact: true }).locator('..')).toContainText('6')
  await expect(page.getByText('Status: Draft', { exact: true }).first()).toBeVisible()
  await expect(page.getByText(/Periode:/).locator('..')).toContainText('Kuartal 1 2026')
  await expect(page.getByRole('button', { name: /Approve|Setujui/ })).toHaveCount(0)

  await openFirstExecutiveSummary(page)
  await expect(page.getByText('Noted dari Executive untuk Auditor', { exact: true })).toBeVisible()
  await expect(page.getByText('Noted', { exact: true }).first()).toBeVisible()
  await expect(page.getByRole('button', { name: /Setujui.*Approve/ })).toBeVisible()
  await expect(page.getByText(/Tren & Grafik/)).toHaveCount(0)
  await expect(page.getByText(/Matriks Induk Kompilasi Temuan/)).toHaveCount(0)

  await page.getByRole('button', { name: 'Tutup' }).click()
  await page.getByRole('button', { name: /Buat|Create/ }).first().click()
  await expect(page.getByText('Periode Kuartal', { exact: true })).toBeVisible()
  await expect(page.getByText('Periode Bulan', { exact: true })).toHaveCount(0)
  await expect(page.getByText('Kuartal I', { exact: true }).last()).toBeVisible()
  await expect(page.getByText('Noted dari Executive untuk Auditor', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('button', { name: /Approve|Setujui/ })).toHaveCount(0)

  await page.getByRole('button', { name: 'Tutup' }).click()
  const heading = page.getByRole('heading', { name: summaryFixture.nomorDokumen, exact: true })
  const card = heading.locator('xpath=ancestor::div[contains(@class, "group")][1]')
  await card.getByRole('button').nth(1).click()
  await expect(page.getByText(/Edit Compiled Report|Edit Executive Summary/).first()).toBeVisible()
  await expect(page.getByText('Noted dari Executive untuk Auditor', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('button', { name: /Approve|Setujui/ })).toHaveCount(0)
})
