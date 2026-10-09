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
  // BAB 2: AUDIT CHARTER (POIN 3)
  // ==========================================
  console.log('\n--- BAB 2: AUDIT CHARTER ---');
  await page.goto('http://localhost:3000/audit-charter', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Gambar 2.1: Fokus ke section Audit Charter Card
  const charterCard = await page.$('.space-y-8 > div:first-child') || await page.$('.space-y-6 .relative.group') || await page.$('.space-y-8');
  if (charterCard) {
    await charterCard.screenshot({ path: path.join(IMAGES_DIR, 'sample_01_audit_charter_card.png') });
    console.log('Saved focused sample_01_audit_charter_card.png');
  }

  // Gambar 2.2: Fokus ke tabel Audit Charter History
  const historyTable = await page.$('table') || await page.$('.border.border-gray-200.dark\\:border-gray-800');
  if (historyTable) {
    await historyTable.screenshot({ path: path.join(IMAGES_DIR, 'sample_04_audit_charter_history.png') });
    console.log('Saved focused sample_04_audit_charter_history.png');
  }

  // ==========================================
  // BAB 3: RISK APPETITE (POIN 4 & 5)
  // ==========================================
  console.log('\n--- BAB 3: RISK APPETITE ---');
  await page.goto('http://localhost:3000/risk-appetite', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Gambar 3.4: Crop ke KPI dashboard widget saja (Poin 5)
  const kpiSection = await page.$('.grid.grid-cols-1.sm\\:grid-cols-2.lg\\:grid-cols-4');
  if (kpiSection) {
    await kpiSection.screenshot({ path: path.join(IMAGES_DIR, 'risk_appetite_v2.png') });
    console.log('Saved focused KPI widget risk_appetite_v2.png');
  }

  // Gambar 3.5: Tab 2 Compliance Validation (Poin 5)
  // Click Tab 2
  const tabs = await page.$$('button[role="tab"]');
  if (tabs.length >= 2) {
    await tabs[1].click();
    await page.waitForTimeout(1000);
    // Crop ke content area tab compliance
    const tabContent = await page.$('.p-4.sm\\:p-6.space-y-6') || await page.$('table');
    if (tabContent) {
      await tabContent.screenshot({ path: path.join(IMAGES_DIR, 'rp_14_appetite_compliance.png') });
      console.log('Saved focused compliance widget rp_14_appetite_compliance.png');
    }
  }

  // Poin 4: Tab 3 Appetite Statement + Modal Add Statement
  if (tabs.length >= 3) {
    await tabs[2].click();
    await page.waitForTimeout(1000);

    // Screenshot Tab 3 Statements Registry
    const tab3Content = await page.$('.p-4.sm\\:p-6.space-y-6') || await page.$('body');
    if (tab3Content) {
      await tab3Content.screenshot({ path: path.join(IMAGES_DIR, '03_04_tab_appetite_statement.png') });
      console.log('Saved 03_04_tab_appetite_statement.png');
    }

    // Click + Add Statement
    const addStmtBtn = await page.$('button:has-text("Add Statement")') || 
                       await page.$('button:has-text("Tambah Pernyataan")') || 
                       await page.$('button:has-text("New Statement")') ||
                       await page.$('button:has-text("Pernyataan Baru")') ||
                       await page.$('.i-heroicons-plus');
    if (addStmtBtn) {
      await addStmtBtn.click();
      await page.waitForTimeout(1000);

      // Screenshot modal
      const modal = await page.$('[role="dialog"]') || await page.$('.fixed.inset-0');
      if (modal) {
        await modal.screenshot({ path: path.join(IMAGES_DIR, '03_05_modal_add_statement.png') });
        console.log('Saved 03_05_modal_add_statement.png');
      }
      // Close modal
      const closeBtn = await page.$('[role="dialog"] button:has-text("Cancel")') || 
                       await page.$('[role="dialog"] button:has-text("Batal")') || 
                       await page.$('[role="dialog"] button[aria-label="Close"]');
      if (closeBtn) await closeBtn.click();
      await page.waitForTimeout(500);
    }
  }

  // ==========================================
  // BAB 3: RISK FACTORS & AUDIT PRIORITY (POIN 6 & 8)
  // ==========================================
  console.log('\n--- BAB 3: AUDIT PRIORITY & RISK FACTORS ---');
  await page.goto('http://localhost:3000/risk-profile/audit-priority', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // 1. Audit Priority Table
  const priorityTable = await page.$('.space-y-8 > div:nth-child(4)') || await page.$('table') || await page.$('.overflow-x-auto');
  if (priorityTable) {
    await priorityTable.screenshot({ path: path.join(IMAGES_DIR, '03_08_audit_priority_table.png') });
    console.log('Saved 03_08_audit_priority_table.png');
  }

  // 2. Corporate Risk Index Level Information
  const levelInfoTable = await page.$('.space-y-8 > div:last-child') || await page.$$('table').then(t => t[t.length - 1]);
  if (levelInfoTable) {
    await levelInfoTable.screenshot({ path: path.join(IMAGES_DIR, '03_09_corporate_risk_index_info.png') });
    console.log('Saved 03_09_corporate_risk_index_info.png');
  }

  // ==========================================
  // BAB 3: AUDIT UNIVERSE (POIN 7)
  // ==========================================
  console.log('\n--- BAB 3: AUDIT UNIVERSE ---');
  await page.goto('http://localhost:3000/risk-profile/audit-universe', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Standard Library
  const standardLibrary = await page.$('.space-y-6 > .shadow-sm') || await page.$('.shadow-sm:has-text("Standard Audit Universe Library")');
  if (standardLibrary) {
    await standardLibrary.screenshot({ path: path.join(IMAGES_DIR, '03_10_standard_audit_universe_library.png') });
    console.log('Saved 03_10_standard_audit_universe_library.png');
  }

  // Corporate Universe
  const corporateUniverse = await page.$('.shadow-sm:has-text("Corporate Audit Universe")') || await page.$('.border.border-slate-150');
  if (corporateUniverse) {
    await corporateUniverse.screenshot({ path: path.join(IMAGES_DIR, '03_11_corporate_audit_universe.png') });
    console.log('Saved 03_11_corporate_audit_universe.png');
  }

  // ==========================================
  // BAB 3: RISK CONTROL MATRIX (POIN 9)
  // ==========================================
  console.log('\n--- BAB 3: RISK CONTROL MATRIX ---');
  await page.goto('http://localhost:3000/risk-profile/risk-control-matrix', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // 1. Pengukuran Tutup Buku Akhir Tahun
  const cardTutupBuku = await page.$('.lg\\:col-span-2');
  if (cardTutupBuku) {
    await cardTutupBuku.screenshot({ path: path.join(IMAGES_DIR, '03_12_rcm_pengukuran_tutup_buku.png') });
    console.log('Saved 03_12_rcm_pengukuran_tutup_buku.png');
  }

  // 2. Rata-Rata COSO 2013
  const cardCoso = await page.$('.bg-white.dark\\:bg-slate-900:has-text("Rata-Rata COSO 2013")');
  if (cardCoso) {
    await cardCoso.screenshot({ path: path.join(IMAGES_DIR, '03_13_rcm_rata_rata_coso_2013.png') });
    console.log('Saved 03_13_rcm_rata_rata_coso_2013.png');
  }

  // 3. Tabel Standar Interpretasi Rating Efektivitas Kontrol Internal (Click accordion to expand first)
  const accordionBtn = await page.$('button:has-text("Tabel Standar Interpretasi Rating Efektivitas Kontrol Internal")');
  if (accordionBtn) {
    await accordionBtn.click();
    await page.waitForTimeout(500);
  }
  const ratingTable = await page.$('.bg-white.dark\\:bg-slate-900:has-text("Tabel Standar Interpretasi Rating Efektivitas Kontrol Internal")');
  if (ratingTable) {
    await ratingTable.screenshot({ path: path.join(IMAGES_DIR, '03_14_rcm_tabel_standar_interpretasi.png') });
    console.log('Saved 03_14_rcm_tabel_standar_interpretasi.png');
  }

  // 4. Daftar Risk Control Matrix Table
  const rcmTable = await page.$('.bg-white.dark\\:bg-slate-900:has-text("Daftar Risk Control Matrix")');
  if (rcmTable) {
    await rcmTable.screenshot({ path: path.join(IMAGES_DIR, '03_15_rcm_daftar_risk_control_matrix.png') });
    console.log('Saved 03_15_rcm_daftar_risk_control_matrix.png');
  }

  console.log('\n--- GROUP 1 & 2 CAPTURE COMPLETE ---');
  await browser.close();
})();
