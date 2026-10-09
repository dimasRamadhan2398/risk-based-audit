const { chromium } = require('/Users/Izaname/Documents/AuditSphere/risk-based-audit-main-2/frontend/node_modules/@playwright/test');
const path = require('path');
const fs = require('fs');

const IMAGES_DIR = path.resolve('docs/user-manual/images');

(async () => {
  const browser = await chromium.launch({
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });

  // 1. Login
  console.log('Logging in...');
  await page.goto('http://localhost:3000/auth/login');
  await page.fill('#username', 'superadmin');
  await page.fill('#password', 'password123');
  await page.click('button[type="submit"]');
  await page.waitForTimeout(2000);

  // Set language to ID
  await page.evaluate(() => {
    localStorage.setItem('i18n_locale', 'id');
  });

  // Inject CSS to suppress ONLY toast notifications
  await page.addStyleTag({
    content: 'div[role="region"][aria-label*="Notification"] { display: none !important; pointer-events: none !important; }'
  });

  // ==========================================
  // BAB 4: AUDIT FIELDWORK (POIN 22)
  // ==========================================
  console.log('\n--- BAB 4: AUDIT FIELDWORK ---');
  await page.goto('http://localhost:3000/audit-fieldwork', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);

  // Re-inject suppression
  await page.addStyleTag({
    content: 'div[role="region"][aria-label*="Notification"] { display: none !important; pointer-events: none !important; }'
  });

  // Select Assignment Letter ST-001/SKAI/2026 via UI
  const fwSelectBtn = await page.$('button:has-text("Pilih Surat Tugas")') || 
                      await page.$('button:has-text("Select Assignment Letter")') ||
                      await page.$('.size-5.i-heroicons-document-text');
  if (fwSelectBtn) {
    await fwSelectBtn.click({ force: true });
    await page.waitForTimeout(500);
    const firstOption = await page.$('[role="option"]:has-text("ST-001/SKAI/2026")') || await page.$('[role="option"]');
    if (firstOption) {
      await firstOption.click({ force: true });
      await page.waitForTimeout(2000);
    }
  }

  const fwTabsInfo = [
    { label: 'Interview', tabImg: '04_fieldwork_01_interview.png', modalBtn: 'Add Interview', modalImg: '04_fieldwork_modal_01_interview.png' },
    { label: 'Observation', tabImg: '04_fieldwork_02_observation.png', modalBtn: 'Add Observation', modalImg: '04_fieldwork_modal_02_observation.png' },
    { label: 'Document Collection', tabImg: '04_fieldwork_03_document.png', modalBtn: 'Add Document', modalImg: '04_fieldwork_modal_03_document.png' },
    { label: 'Sample Data', tabImg: '04_fieldwork_04_sample.png', modalBtn: 'Add Sample', modalImg: '04_fieldwork_modal_04_sample.png' },
    { label: 'Test Controls', tabImg: '04_fieldwork_05_test_control.png', modalBtn: 'Add Test Control', modalImg: '04_fieldwork_modal_05_test_control.png' }
  ];

  for (let i = 0; i < fwTabsInfo.length; i++) {
    const tabInfo = fwTabsInfo[i];
    // Click tab i
    const tabBtn = await page.$(`button[role="tab"]:has-text("${tabInfo.label}")`);
    if (tabBtn) {
      await tabBtn.click({ force: true });
      await page.waitForTimeout(1000);
    }

    // Take screenshot of tab
    await page.screenshot({ path: path.join(IMAGES_DIR, tabInfo.tabImg) });
    console.log(`Saved ${tabInfo.tabImg}`);

    // Click Add Button
    const addBtn = await page.$(`button:has-text("${tabInfo.modalBtn}")`);
    if (addBtn) {
      await addBtn.click({ force: true });
      await page.waitForTimeout(1000);

      // Screenshot modal
      const modal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
      if (modal) {
        await modal.screenshot({ path: path.join(IMAGES_DIR, tabInfo.modalImg) });
        console.log(`Saved ${tabInfo.modalImg}`);
      }

      // Close modal
      const closeBtn = await page.$('[role="dialog"] button:has-text("Batal")') || 
                       await page.$('[role="dialog"] button:has-text("Cancel")') || 
                       await page.$('[role="dialog"] button[aria-label="Close"]') ||
                       await page.$('.i-heroicons-x-mark-20-solid');
      if (closeBtn) await closeBtn.click({ force: true });
      await page.waitForTimeout(500);
    }
  }

  // ==========================================
  // BAB 5: RESULTS REPORT (LHA) (POIN 23)
  // ==========================================
  console.log('\n--- BAB 5: RESULTS REPORT (LHA) ---');
  await page.goto('http://localhost:3000/audit-result-report', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);

  await page.addStyleTag({
    content: 'div[role="region"][aria-label*="Notification"] { display: none !important; pointer-events: none !important; }'
  });

  // Select Assignment Letter ST-001/SKAI/2026 via UI
  const lhaSelectBtn = await page.$('button:has-text("Pilih Surat Tugas")') || 
                       await page.$('button:has-text("Select Assignment Letter")') ||
                       await page.$('.size-5.i-heroicons-document-text');
  if (lhaSelectBtn) {
    await lhaSelectBtn.click({ force: true });
    await page.waitForTimeout(500);
    const opt = await page.$('[role="option"]:has-text("ST-001/SKAI/2026")') || await page.$('[role="option"]');
    if (opt) {
      await opt.click({ force: true });
      await page.waitForTimeout(2000);
    }
  }

  // Screenshot page with selected SPT
  await page.screenshot({ path: path.join(IMAGES_DIR, '05_lha_selected_spt.png') });
  console.log('Saved 05_lha_selected_spt.png');

  // Click Edit button
  const editBtn = await page.$('.i-lucide-edit');
  if (editBtn) {
    await editBtn.click({ force: true });
    await page.waitForTimeout(1000);

    const modal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
    if (modal) {
      await modal.screenshot({ path: path.join(IMAGES_DIR, '05_lha_modal_edit.png') });
      console.log('Saved 05_lha_modal_edit.png');
    }

    // Close modal
    const closeBtn = await page.$('[role="dialog"] button:has-text("Batal")') || 
                     await page.$('[role="dialog"] button:has-text("Cancel")') || 
                     await page.$('[role="dialog"] button[aria-label="Close"]') ||
                     await page.$('.i-heroicons-x-mark-20-solid');
    if (closeBtn) await closeBtn.click({ force: true });
    await page.waitForTimeout(500);
  }

  // ==========================================
  // BAB 5: ACTION TAKEN REPORT (ATR) (POIN 24)
  // ==========================================
  console.log('\n--- BAB 5: ACTION TAKEN REPORT ---');
  await page.goto('http://localhost:3000/action-taken-report', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  await page.addStyleTag({
    content: 'div[role="region"][aria-label*="Notification"] { display: none !important; pointer-events: none !important; }'
  });

  // Click View (eye icon)
  const viewBtn = await page.$('.i-lucide-eye');
  if (viewBtn) {
    await viewBtn.click({ force: true });
    await page.waitForTimeout(1000);

    const modal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
    if (modal) {
      await modal.screenshot({ path: path.join(IMAGES_DIR, '05_atr_modal_view_detail.png') });
      console.log('Saved 05_atr_modal_view_detail.png');
    }

    // Close modal
    const closeBtn = await page.$('[role="dialog"] button:has-text("Close")') || 
                     await page.$('[role="dialog"] button:has-text("Tutup")') ||
                     await page.$('[role="dialog"] button[aria-label="Close"]');
    if (closeBtn) await closeBtn.click({ force: true });
    await page.waitForTimeout(500);
  }

  // ==========================================
  // BAB 5: CLIENT SATISFACTION SURVEY (CSAT) (POIN 25)
  // ==========================================
  console.log('\n--- BAB 5: CLIENT SATISFACTION SURVEY ---');
  await page.goto('http://localhost:3000/audit-result-report/satisfaction-survey', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  await page.addStyleTag({
    content: 'div[role="region"][aria-label*="Notification"] { display: none !important; pointer-events: none !important; }'
  });

  // 1. Summary Cards
  const csatCards = await page.$('.grid.grid-cols-1.md\\:grid-cols-3.gap-6.mb-6') || await page.$('.grid.grid-cols-1.md\\:grid-cols-3');
  if (csatCards) {
    await csatCards.screenshot({ path: path.join(IMAGES_DIR, '05_csat_01_cards_summary.png') });
    console.log('Saved 05_csat_01_cards_summary.png');
  }

  // 2. Metrics Guide & Explanation Banner
  // Check if explanation is already shown or click toggle button
  let csatGuide = await page.$('.mb-8.border.border-primary-200');
  if (!csatGuide) {
    const guideToggle = await page.$('button:has-text("Penjelasan")') || 
                        await page.$('button:has-text("Explanation")') ||
                        await page.$('.i-heroicons-information-circle');
    if (guideToggle) {
      await guideToggle.click({ force: true });
      await page.waitForTimeout(800);
      csatGuide = await page.$('.mb-8.border.border-primary-200');
    }
  }
  if (csatGuide) {
    await csatGuide.screenshot({ path: path.join(IMAGES_DIR, '05_csat_02_metrics_guide.png') });
    console.log('Saved 05_csat_02_metrics_guide.png');
  }

  // 3. Published Reports & Survey Status Table
  const csatTable = await page.$('.shadow-sm.border.border-\\[var\\(--border-main\\)\\]') || await page.$('.shadow-sm');
  if (csatTable) {
    await csatTable.screenshot({ path: path.join(IMAGES_DIR, '05_csat_03_published_table.png') });
    console.log('Saved 05_csat_03_published_table.png');
  }

  // 4. Modal Fill Survey
  const fillSurveyBtn = await page.$('button:has-text("Isi Survei")') || 
                        await page.$('button:has-text("Fill Survey")') ||
                        await page.$('.i-heroicons-pencil-square');
  if (fillSurveyBtn) {
    await fillSurveyBtn.click({ force: true });
    await page.waitForTimeout(1000);

    const surveyModal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
    if (surveyModal) {
      await surveyModal.screenshot({ path: path.join(IMAGES_DIR, '05_csat_04_modal_fill_survey.png') });
      console.log('Saved 05_csat_04_modal_fill_survey.png');
    }
  }

  console.log('\n--- ALL GROUP 4 & 5 SCREENSHOTS CAPTURED SUCCESSFULLY! ---');
  await browser.close();
})();
