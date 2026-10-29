const fs = require('fs');
const path = require('path');
const { marked } = require('/Users/Izaname/Documents/AuditSphere/risk-based-audit-main-2/frontend/node_modules/marked');
const { chromium } = require('/Users/Izaname/Documents/AuditSphere/risk-based-audit-main-2/frontend/node_modules/@playwright/test');

const DOCS_DIR = path.resolve('docs/user-manual');
const IMAGES_DIR = path.resolve('docs/user-manual/images');

function getImageBase64(imgRelPath) {
  const cleanPath = imgRelPath.replace(/^images\//, '');
  const fullPath = path.join(IMAGES_DIR, cleanPath);
  if (!fs.existsSync(fullPath)) {
    console.warn('Image not found: ' + fullPath);
    return null;
  }
  const ext = path.extname(fullPath).toLowerCase();
  const mime = ext === '.png' ? 'image/png' : 'image/jpeg';
  return 'data:' + mime + ';base64,' + fs.readFileSync(fullPath).toString('base64');
}

function mdToHtml(md, title) {
  let p = md.replace(/<!--\s*pagebreak\s*-->/gi, '<div style="page-break-after: always; break-after: page;"></div>');
  p = p.replace(/!\[(.*?)\]\((images\/.*?)\)/g, (m, alt, src) => {
    const b64 = getImageBase64(src);
    if (!b64) return '<p>[' + alt + ']</p>';
    return '<div style="margin: 14pt 0; text-align: center; page-break-inside: avoid; break-inside: avoid;"><img src="' + b64 + '" style="max-width: 100%; border: 1px solid #cbd5e1; border-radius: 6px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1);" /><div style="font-size: 8.5pt; color: #64748b; font-weight: 600; margin-top: 6pt;">' + alt + '</div></div>';
  });

  const parsedBody = marked.parse(p);

  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>${title}</title>
<style>
@page {
  size: A4;
  margin: 18mm 16mm;
}
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  font-size: 9.5pt;
  line-height: 1.55;
  color: #1e293b;
  margin: 0;
  padding: 0;
}
em, i {
  font-style: normal !important;
  font-weight: 600;
  color: inherit;
}
h1 {
  font-size: 18pt;
  color: #0f172a;
  border-bottom: 2px solid #ea580c;
  padding-bottom: 6pt;
  margin-top: 0;
  margin-bottom: 12pt;
}
h2 {
  font-size: 13pt;
  color: #0f172a;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 4pt;
  margin-top: 14pt;
  margin-bottom: 8pt;
  page-break-after: avoid;
  break-after: avoid;
}
h3 {
  font-size: 11pt;
  color: #1e293b;
  margin-top: 10pt;
  margin-bottom: 6pt;
  page-break-after: avoid;
  break-after: avoid;
}
h4, h5 {
  font-size: 10pt;
  color: #334155;
  margin-top: 8pt;
  margin-bottom: 4pt;
  page-break-after: avoid;
  break-after: avoid;
}
p {
  margin-top: 0;
  margin-bottom: 8pt;
}
ul, ol {
  margin-top: 0;
  margin-bottom: 8pt;
  padding-left: 20px;
}
li {
  margin-bottom: 4pt;
}
table {
  width: 100%;
  border-collapse: collapse;
  margin: 12pt 0;
  font-size: 8.5pt;
  page-break-inside: avoid;
  break-inside: avoid;
}
th, td {
  border: 1px solid #cbd5e1;
  padding: 5pt 7pt;
  text-align: left;
}
th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #0f172a;
}
tr:nth-child(even) {
  background-color: #fcfdfe;
}
code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 8.5pt;
  background-color: #f1f5f9;
  padding: 1.5pt 3pt;
  border-radius: 4pt;
  color: #0f172a;
}
hr {
  border: none;
  border-top: 1px solid #e2e8f0;
  margin: 14pt 0;
}
blockquote {
  border-left: 4px solid #ea580c;
  margin: 10pt 0;
  padding: 6pt 12pt;
  background-color: #fff7ed;
  color: #9a3412;
  border-radius: 0 4pt 4pt 0;
}
</style>
</head>
<body>
${parsedBody}
</body>
</html>`;
}

async function renderDoc(browser, mdFile, htmlFile, pdfFile, title) {
  console.log(`Processing: ${mdFile}...`);
  const mdPath = path.join(DOCS_DIR, mdFile);
  const htmlPath = path.join(DOCS_DIR, htmlFile);
  const pdfPath = path.join(DOCS_DIR, pdfFile);

  const md = fs.readFileSync(mdPath, 'utf8');
  const html = mdToHtml(md, title);
  fs.writeFileSync(htmlPath, html, 'utf8');

  const page = await browser.newPage();
  await page.goto('file://' + htmlPath, { waitUntil: 'networkidle' });
  await page.pdf({
    path: pdfPath,
    format: 'A4',
    printBackground: true,
    displayHeaderFooter: true,
    headerTemplate: `<div style="font-size: 8pt; color: #94a3b8; width: 100%; padding: 0 16mm; display: flex; justify-content: space-between; font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif;">
      <span style="font-weight: 700; color: #ea580c;">AUDITSPHERE ENTERPRISE</span>
      <span>${title}</span>
    </div>`,
    footerTemplate: `<div style="font-size: 8pt; color: #94a3b8; width: 100%; padding: 0 16mm; display: flex; justify-content: space-between; font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif;">
      <span>Hak Cipta © 2026 AuditSphere Platform</span>
      <span>Halaman <span class="pageNumber"></span> dari <span class="totalPages"></span></span>
    </div>`,
    margin: { top: '18mm', bottom: '18mm', left: '16mm', right: '16mm' }
  });
  await page.close();
  console.log(`SUCCESS: Rendered ${pdfFile}`);
}

(async () => {
  const browser = await chromium.launch({
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  try {
    await renderDoc(browser, 'BAB_2_AUDIT_CHARTER.md', 'BAB_2_AUDIT_CHARTER.html', 'BAB_2_AUDIT_CHARTER.pdf', 'BAB 2: AUDIT CHARTER & GUIDELINES');
    await renderDoc(browser, 'BAB_3_RISK_PROFILE_DAN_AUDIT_UNIVERSE.md', 'BAB_3_RISK_PROFILE_DAN_AUDIT_UNIVERSE.html', 'BAB_3_RISK_PROFILE_DAN_AUDIT_UNIVERSE.pdf', 'BAB 3: RISK PROFILE & AUDIT UNIVERSE');
    await renderDoc(browser, 'BAB_4_PERENCANAAN_DAN_PELAKSANAAN_AUDIT.md', 'BAB_4_PERENCANAAN_DAN_PELAKSANAAN_AUDIT.html', 'BAB_4_PERENCANAAN_DAN_PELAKSANAAN_AUDIT.pdf', 'BAB 4: PERENCANAAN AUDIT & FIELDWORK');
    await renderDoc(browser, 'BAB_5_PELAPORAN_HASIL_AUDIT_DAN_TINDAK_LANJUT.md', 'BAB_5_PELAPORAN_HASIL_AUDIT_DAN_TINDAK_LANJUT.html', 'BAB_5_PELAPORAN_HASIL_AUDIT_DAN_TINDAK_LANJUT.pdf', 'BAB 5: PELAPORAN HASIL AUDIT DAN ATR');
    await renderDoc(browser, 'BUKU_PANDUAN_PENGGUNA_AUDITSPHERE_RINGKAS.md', 'BUKU_PANDUAN_PENGGUNA_AUDITSPHERE_RINGKAS.html', 'BUKU_PANDUAN_PENGGUNA_AUDITSPHERE_RINGKAS.pdf', 'PANDUAN PENGGUNA RINGKAS');
    await renderDoc(browser, 'BUKU_PANDUAN_PENGGUNA_AUDITSPHERE_LENGKAP.md', 'BUKU_PANDUAN_PENGGUNA_AUDITSPHERE_LENGKAP.html', 'BUKU_PANDUAN_PENGGUNA_AUDITSPHERE_LENGKAP.pdf', 'BUKU PANDUAN PENGGUNA LENGKAP AUDITSPHERE');
  } finally {
    await browser.close();
  }
})();
