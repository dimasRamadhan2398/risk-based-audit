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

  // ==========================================
  // BAB 4: STRATEGIC AUDIT PLAN (POIN 11 & 12)
  // ==========================================
  console.log('\n--- BAB 4: STRATEGIC AUDIT PLAN ---');
  await page.goto('http://localhost:3000/strategic-audit-plan', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Poin 11: Modal + Add Vision and Mission
  try {
    await page.evaluate(() => {
      const vmgStore = (window).$nuxt?.$pinia?._s?.get('vision-mission-goals');
      if (vmgStore) vmgStore.openModal();
    });
    await page.waitForTimeout(1000);

    const vmgModal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
    if (vmgModal) {
      await vmgModal.screenshot({ path: path.join(IMAGES_DIR, '04_strategic_modal_add_vision_mission.png') });
      console.log('Saved 04_strategic_modal_add_vision_mission.png');
      // close modal
      const closeBtn = await page.$('[role="dialog"] button:has-text("Batal")') || 
                       await page.$('[role="dialog"] button:has-text("Cancel")') ||
                       await page.$('[role="dialog"] button[aria-label="Close"]');
      if (closeBtn) await closeBtn.click();
      await page.waitForTimeout(500);
    }
  } catch (e) {
    console.error('Error capturing VMG modal:', e.message);
  }

  // Poin 12: Import Strategic Audit Plan
  await page.goto('http://localhost:3000/strategic-audit-plan/upload', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(IMAGES_DIR, '04_strategic_import.png') });
  console.log('Saved 04_strategic_import.png');

  // ==========================================
  // BAB 4: INTERNAL AUDIT PERFORMANCE (POIN 13)
  // ==========================================
  console.log('\n--- BAB 4: INTERNAL AUDIT PERFORMANCE ---');
  await page.goto('http://localhost:3000/kpi-performance/upload', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(IMAGES_DIR, '04_performance_import.png') });
  console.log('Saved 04_performance_import.png');

  // ==========================================
  // BAB 4: ANNUAL AUDIT PLAN (POIN 14)
  // ==========================================
  console.log('\n--- BAB 4: ANNUAL AUDIT PLAN ---');
  await page.goto('http://localhost:3000/annual-audit/upload', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(IMAGES_DIR, '04_annual_import.png') });
  console.log('Saved 04_annual_import.png');

  // ==========================================
  // BAB 4: AUDIT ACTIVITY PLAN (POIN 16, 17, 18)
  // ==========================================
  console.log('\n--- BAB 4: AUDIT ACTIVITY PLAN ---');
  await page.goto('http://localhost:3000/audit-activity-plan', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Poin 16: 5 Stepper Form Screenshots
  try {
    await page.evaluate(() => {
      const actStore = (window).$nuxt?.$pinia?._s?.get('activity-plan');
      if (actStore) {
        actStore.openModal();
        // fill sample data
        actStore.formState.planTitle = 'Rencana Kegiatan Audit Operasional Keuangan & TI 2026';
        actStore.formState.planYear = '2026';
        actStore.formState.planPeriodStart = '2026-02-01';
        actStore.formState.planPeriodEnd = '2026-11-30';
      }
    });
    await page.waitForTimeout(1000);

    const stepFiles = [
      '04_activity_01_basic_info.png',
      '04_activity_02_activities.png',
      '04_activity_03_resources.png',
      '04_activity_04_review.png',
      '04_activity_05_attachment.png'
    ];

    for (let i = 0; i < 5; i++) {
      // Click stepper item
      const stepPills = await page.$$('.flex.items-center.gap-2.sm\\:gap-2\\.5.cursor-pointer');
      if (stepPills[i]) {
        await stepPills[i].click();
        await page.waitForTimeout(800);
      }
      const modal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
      if (modal) {
        await modal.screenshot({ path: path.join(IMAGES_DIR, stepFiles[i]) });
        console.log(`Saved ${stepFiles[i]}`);
      }
    }

    // Close modal
    const closeBtn = await page.$('[role="dialog"] button:has-text("Batal")') || 
                     await page.$('[role="dialog"] button:has-text("Cancel")') ||
                     await page.$('[role="dialog"] button[aria-label="Close"]') ||
                     await page.$('.i-heroicons-x-mark-20-solid');
    if (closeBtn) await closeBtn.click();
    await page.waitForTimeout(500);
  } catch (e) {
    console.error('Error capturing Activity Plan Form:', e.message);
  }

  // Poin 17 & 18: Import Audit Activity Plan
  await page.goto('http://localhost:3000/audit-activity-plan/upload', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(IMAGES_DIR, '04_activity_import.png') });
  console.log('Saved 04_activity_import.png');

  // ==========================================
  // BAB 4: ASSIGNMENT LETTER (POIN 19)
  // ==========================================
  console.log('\n--- BAB 4: ASSIGNMENT LETTER ---');
  await page.goto('http://localhost:3000/assignment-letter/upload', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(IMAGES_DIR, '04_assignment_import.png') });
  console.log('Saved 04_assignment_import.png');

  // ==========================================
  // BAB 4: WORKING PAPER (POIN 20 & 21)
  // ==========================================
  console.log('\n--- BAB 4: WORKING PAPER ---');
  await page.goto('http://localhost:3000/working-paper', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  const wpTabs = ['f01', 'f02', 'f03', 'f04', 'f05'];
  const wpTableNames = [
    '04_wp_table_01_header.png',
    '04_wp_table_02_risk.png',
    '04_wp_table_03_sample.png',
    '04_wp_table_04_cause.png',
    '04_wp_table_05_action.png'
  ];
  const wpCreateNames = [
    '04_wp_create_01_header.png',
    '04_wp_create_02_risk.png',
    '04_wp_create_03_sample.png',
    '04_wp_create_04_cause.png',
    '04_wp_create_05_action.png'
  ];

  for (let i = 0; i < 5; i++) {
    const tabKey = wpTabs[i];
    // Switch to step
    await page.evaluate((step) => {
      // Find stepper buttons
      const buttons = Array.from(document.querySelectorAll('button'));
      const stepBtn = buttons.find(b => b.innerText.includes(['Header', 'Risk Profile', 'Test Sample', 'AOI & RCA', 'Action Plan'][step]));
      if (stepBtn) stepBtn.click();
    }, i);
    await page.waitForTimeout(1000);

    // 1. Table screenshot
    const tableSection = await page.$('.shadow-sm') || await page.$('table') || await page.$('body');
    if (tableSection) {
      await tableSection.screenshot({ path: path.join(IMAGES_DIR, wpTableNames[i]) });
      console.log(`Saved ${wpTableNames[i]}`);
    }

    // 2. Open Modal Create
    await page.evaluate((step) => {
      const wpStore = (window).$nuxt?.$pinia?._s?.get('working-paper');
      if (wpStore) {
        if (step === 0) wpStore.openModalF01();
        if (step === 1) wpStore.openModalF02();
        if (step === 2) wpStore.openModalF03();
        if (step === 3) wpStore.openModalF04();
        if (step === 4) wpStore.openModalF05();
      }
    }, i);
    await page.waitForTimeout(1000);

    const modal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
    if (modal) {
      await modal.screenshot({ path: path.join(IMAGES_DIR, wpCreateNames[i]) });
      console.log(`Saved ${wpCreateNames[i]}`);
    }

    // Close modal
    await page.evaluate((step) => {
      const wpStore = (window).$nuxt?.$pinia?._s?.get('working-paper');
      if (wpStore) {
        if (step === 0) wpStore.closeModalF01();
        if (step === 1) wpStore.closeModalF02();
        if (step === 2) wpStore.closeModalF03();
        if (step === 3) wpStore.closeModalF04();
        if (step === 4) wpStore.closeModalF05();
      }
    }, i);
    await page.waitForTimeout(500);
  }

  // Poin 21: Import Working Paper
  await page.goto('http://localhost:3000/working-paper/upload', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(IMAGES_DIR, '04_wp_import.png') });
  console.log('Saved 04_wp_import.png');

  // ==========================================
  // BAB 4: AUDIT FIELDWORK (POIN 22)
  // ==========================================
  console.log('\n--- BAB 4: AUDIT FIELDWORK ---');
  await page.goto('http://localhost:3000/audit-fieldwork', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Select Assignment Letter ST-001/SKAI/2026
  await page.evaluate(() => {
    const fwStore = (window).$nuxt?.$pinia?._s?.get('audit-fieldwork');
    if (fwStore) {
      fwStore.selectedAssignmentLetter = 'ST-001/SKAI/2026';
      fwStore.ensureDataExists();
    }
  });
  await page.waitForTimeout(1500);

  const fwTabNames = [
    '04_fieldwork_01_interview.png',
    '04_fieldwork_02_observation.png',
    '04_fieldwork_03_document.png',
    '04_fieldwork_04_sample.png',
    '04_fieldwork_05_test_control.png'
  ];
  const fwModalNames = [
    '04_fieldwork_modal_01_interview.png',
    '04_fieldwork_modal_02_observation.png',
    '04_fieldwork_modal_03_document.png',
    '04_fieldwork_modal_04_sample.png',
    '04_fieldwork_modal_05_test_control.png'
  ];

  for (let i = 0; i < 5; i++) {
    // Click tab
    const tabs = await page.$$('button[role="tab"]');
    if (tabs[i]) {
      await tabs[i].click();
      await page.waitForTimeout(1000);
    }

    // Screenshot tab table
    const content = await page.$('.space-y-6') || await page.$('body');
    if (content) {
      await content.screenshot({ path: path.join(IMAGES_DIR, fwTabNames[i]) });
      console.log(`Saved ${fwTabNames[i]}`);
    }

    // Open Add Modal
    const addBtn = await page.$('button:has-text("Add")') || 
                   await page.$('button:has-text("Tambah")') ||
                   await page.$('.i-heroicons-plus');
    if (addBtn) {
      await addBtn.click();
      await page.waitForTimeout(1000);

      const modal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
      if (modal) {
        await modal.screenshot({ path: path.join(IMAGES_DIR, fwModalNames[i]) });
        console.log(`Saved ${fwModalNames[i]}`);
      }

      // Close modal
      const closeBtn = await page.$('[role="dialog"] button:has-text("Batal")') || 
                       await page.$('[role="dialog"] button:has-text("Cancel")') ||
                       await page.$('[role="dialog"] button[aria-label="Close"]') ||
                       await page.$('.i-heroicons-x-mark-20-solid');
      if (closeBtn) await closeBtn.click();
      await page.waitForTimeout(500);
    }
  }

  // ==========================================
  // BAB 5: RESULTS REPORT (LHA) (POIN 23)
  // ==========================================
  console.log('\n--- BAB 5: RESULTS REPORT (LHA) ---');
  await page.goto('http://localhost:3000/audit-result-report', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Select Assignment Letter ST-001/SKAI/2026
  await page.evaluate(() => {
    const lhaStore = (window).$nuxt?.$pinia?._s?.get('audit-result-report');
    if (lhaStore) {
      lhaStore.selectedAssignmentLetter = 'ST-001/SKAI/2026';
    }
  });
  await page.waitForTimeout(1500);

  // Screenshot table with selected assignment letter
  await page.screenshot({ path: path.join(IMAGES_DIR, '05_lha_selected_spt.png') });
  console.log('Saved 05_lha_selected_spt.png');

  // Open Edit Modal
  await page.evaluate(() => {
    const lhaStore = (window).$nuxt?.$pinia?._s?.get('audit-result-report');
    if (lhaStore && lhaStore.filteredReports.length > 0) {
      lhaStore.editReport(lhaStore.filteredReports[0]);
    }
  });
  await page.waitForTimeout(1000);

  const lhaModal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
  if (lhaModal) {
    await lhaModal.screenshot({ path: path.join(IMAGES_DIR, '05_lha_modal_edit.png') });
    console.log('Saved 05_lha_modal_edit.png');
    // Close modal
    const closeBtn = await page.$('[role="dialog"] button:has-text("Batal")') || 
                     await page.$('[role="dialog"] button:has-text("Cancel")') ||
                     await page.$('[role="dialog"] button[aria-label="Close"]') ||
                     await page.$('.i-heroicons-x-mark-20-solid');
    if (closeBtn) await closeBtn.click();
    await page.waitForTimeout(500);
  }

  // ==========================================
  // BAB 5: ACTION TAKEN REPORT (ATR) (POIN 24)
  // ==========================================
  console.log('\n--- BAB 5: ACTION TAKEN REPORT ---');
  await page.goto('http://localhost:3000/action-taken-report', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Open Detail Modal
  await page.evaluate(() => {
    const atrStore = (window).$nuxt?.$pinia?._s?.get('action-taken-report');
    if (atrStore && atrStore.filteredReports.length > 0) {
      atrStore.openDetail(atrStore.filteredReports[0]);
    }
  });
  await page.waitForTimeout(1000);

  const atrModal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
  if (atrModal) {
    await atrModal.screenshot({ path: path.join(IMAGES_DIR, '05_atr_modal_view_detail.png') });
    console.log('Saved 05_atr_modal_view_detail.png');
    // Close modal
    const closeBtn = await page.$('[role="dialog"] button:has-text("Close")') || 
                     await page.$('[role="dialog"] button:has-text("Tutup")') ||
                     await page.$('[role="dialog"] button[aria-label="Close"]');
    if (closeBtn) await closeBtn.click();
    await page.waitForTimeout(500);
  }

  // ==========================================
  // BAB 5: CLIENT SATISFACTION SURVEY (CSAT) (POIN 25)
  // ==========================================
  console.log('\n--- BAB 5: CLIENT SATISFACTION SURVEY ---');
  await page.goto('http://localhost:3000/audit-result-report/satisfaction-survey', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // 1. Summary Cards
  const csatCards = await page.$('.grid.grid-cols-1.md\\:grid-cols-3.gap-6.mb-6') || await page.$('.grid.grid-cols-1.md\\:grid-cols-3');
  if (csatCards) {
    await csatCards.screenshot({ path: path.join(IMAGES_DIR, '05_csat_01_cards_summary.png') });
    console.log('Saved 05_csat_01_cards_summary.png');
  }

  // 2. Metrics Guide & Explanation Banner
  // Ensure explanation is shown
  await page.evaluate(() => {
    // If showExplanationGuide is false, click button
    const btn = Array.from(document.querySelectorAll('button')).find(b => b.innerText.includes('Penjelasan') || b.innerText.includes('Explanation'));
    if (btn) btn.click();
  });
  await page.waitForTimeout(1000);

  const csatGuide = await page.$('.mb-8.border.border-primary-200') || await page.$('.border-primary-200') || await page.$('.rounded-2xl.overflow-hidden');
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
    await fillSurveyBtn.click();
    await page.waitForTimeout(1000);

    const surveyModal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
    if (surveyModal) {
      await surveyModal.screenshot({ path: path.join(IMAGES_DIR, '05_csat_04_modal_fill_survey.png') });
      console.log('Saved 05_csat_04_modal_fill_survey.png');
    }
  }

  console.log('\n--- ALL GROUP 3 & 4 SCREENSHOTS CAPTURED SUCCESSFULLY! ---');
  await browser.close();
})();
