const { chromium } = require('/Users/Izaname/Documents/AuditSphere/risk-based-audit-main-2/frontend/node_modules/@playwright/test');

(async () => {
  const browser = await chromium.launch({
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });

  // 1. Login
  console.log('Navigating to login...');
  await page.goto('http://localhost:3000/auth/login', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1000);

  // Fill login
  await page.fill('#username', 'superadmin');
  await page.fill('#password', 'password123');
  await page.click('button[type="submit"]');

  await page.waitForTimeout(2000);
  console.log('Current URL after login:', page.url());

  // Check sidebar navigation text
  const sidebarText = await page.innerText('body');
  console.log('Contains "Audit Planning & Fieldwork"?', sidebarText.includes('Audit Planning & Fieldwork') || sidebarText.includes('Perencanaan Audit & Fieldwork'));
  console.log('Contains "Settings" in sidebar?', sidebarText.includes('Settings') || sidebarText.includes('Pengaturan'));

  // Take screenshot of dashboard with sidebar
  await page.screenshot({ path: 'scratch/staging_dashboard.png', fullPage: false });
  console.log('Saved scratch/staging_dashboard.png');

  await browser.close();
})();
