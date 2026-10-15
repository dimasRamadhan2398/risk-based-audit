/**
 * Helper to open uploaded documents in a new browser tab with custom title matching document name.
 */

export function escapeHtml(text: string): string {
  return String(text || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

export function detectMimeType(fileName?: string, responseType?: string): string {
  if (responseType && responseType !== 'application/octet-stream') {
    return responseType;
  }
  if (!fileName) return 'application/pdf';

  const lower = fileName.toLowerCase();
  if (lower.endsWith('.pdf')) return 'application/pdf';
  if (lower.endsWith('.png')) return 'image/png';
  if (lower.endsWith('.jpg') || lower.endsWith('.jpeg')) return 'image/jpeg';
  if (lower.endsWith('.txt')) return 'text/plain';
  if (lower.endsWith('.html') || lower.endsWith('.htm')) return 'text/html';
  return responseType || 'application/pdf';
}

/**
 * Synchronously opens a new blank browser tab and displays a loading spinner.
 * Must be invoked directly in the user click gesture before any awaits to avoid popup blockers.
 */
export function openDocumentTab(documentName: string): Window | null {
  if (typeof window === 'undefined') return null;

  try {
    const tab = window.open('', '_blank');
    if (tab) {
      try {
        tab.document.title = `Memuat ${documentName}...`;
        tab.document.body.style.margin = '0';
        tab.document.body.style.backgroundColor = '#1e293b';
        tab.document.body.style.color = '#f8fafc';
        tab.document.body.style.fontFamily = 'system-ui, -apple-system, sans-serif';
        tab.document.body.style.display = 'flex';
        tab.document.body.style.alignItems = 'center';
        tab.document.body.style.justifyContent = 'center';
        tab.document.body.style.height = '100vh';
        tab.document.body.innerHTML = `
          <div style="text-align:center; padding: 20px;">
            <div style="display:inline-block; width:36px; height:36px; border:3px solid rgba(255,255,255,0.15); border-top-color:#38bdf8; border-radius:50%; animation:spin 1s linear infinite;"></div>
            <p style="margin-top:16px; font-size:15px; font-weight:500;">Memuat ${escapeHtml(documentName)}...</p>
            <style>@keyframes spin { to { transform: rotate(360deg); } }</style>
          </div>
        `;
      } catch (_e) {
        // Ignore cross-origin access errors if any
      }
    }
    return tab;
  } catch (_e) {
    return null;
  }
}

/**
 * Injects the fetched document blob into the opened tab within a dedicated HTML wrapper.
 * Sets the tab title to documentName.
 */
export function renderDocumentInTab(
  tab: Window | null,
  blobData: any,
  documentName: string,
  fileName?: string
): void {
  const mimeType = detectMimeType(fileName || documentName, blobData?.type);
  const file = new File([blobData], documentName, { type: mimeType });
  const url = window.URL.createObjectURL(file);

  if (tab) {
    try {
      tab.document.open();
      tab.document.write(`
        <!DOCTYPE html>
        <html lang="id">
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1.0">
          <title>${escapeHtml(documentName)}</title>
          <link rel="icon" href="/favicon.ico">
          <style>
            html, body {
              margin: 0;
              padding: 0;
              width: 100%;
              height: 100%;
              overflow: hidden;
              background-color: #525659;
            }
            iframe {
              width: 100%;
              height: 100%;
              border: none;
              display: block;
            }
          </style>
        </head>
        <body>
          <iframe src="${url}"></iframe>
        </body>
        </html>
      `);
      tab.document.close();

      tab.document.title = documentName;
      tab.addEventListener('beforeunload', () => {
        window.URL.revokeObjectURL(url);
      });
      return;
    } catch (_err) {
      // Fallback if writing to tab failed
    }
  }

  window.open(url, '_blank');
}

/**
 * High-level helper: Opens tab, fetches document via fetchFn, and displays it with tab title = documentName.
 */
export async function viewDocumentInNewTab(
  fetchFn: () => Promise<any>,
  documentName: string,
  fileName?: string
): Promise<void> {
  const finalDocName = documentName || fileName || 'document.pdf';
  const tab = openDocumentTab(finalDocName);

  try {
    const data = await fetchFn();
    renderDocumentInTab(tab, data, finalDocName, fileName);
  } catch (err) {
    if (tab) {
      try {
        tab.close();
      } catch (_) {}
    }
    throw err;
  }
}
