/**
 * Helpers for files served behind the bearer-token check (/uploads/* and /media/download/*).
 *
 * A plain window.open, <a href> or location.href sends no Authorization header, so these files
 * are fetched with $fetch (the api-auth plugin attaches the token) and then shown or saved from
 * a blob URL.
 */

/** Object URLs stay valid long enough for a newly opened tab to load the document. */
const VIEW_URL_TTL_MS = 60_000

/** Fetches a protected file as a Blob. Rejects on HTTP errors, like any $fetch call. */
export const fetchProtectedFile = (url: string): Promise<Blob> => {
  return $fetch<Blob>(url, { responseType: 'blob' })
}

/**
 * Shows a fetched file in `tab`. Pass a tab that was opened synchronously in the click handler
 * (window.open('', '_blank')) so the popup blocker allows it; without one a new tab is opened.
 */
export const showBlobInTab = (tab: Window | null, blob: Blob): void => {
  const objectUrl = window.URL.createObjectURL(blob)
  if (tab && !tab.closed) {
    tab.location.href = objectUrl
  }
  else {
    window.open(objectUrl, '_blank')
  }
  setTimeout(() => window.URL.revokeObjectURL(objectUrl), VIEW_URL_TTL_MS)
}

/**
 * Fetches a protected file and opens it in `tab`. If the fetch fails the pre-opened tab is closed
 * and the error is rethrown, so the caller can show its error toast.
 */
export const openProtectedFile = async (url: string, tab: Window | null): Promise<void> => {
  let blob: Blob
  try {
    blob = await fetchProtectedFile(url)
  }
  catch (err) {
    tab?.close()
    throw err
  }
  showBlobInTab(tab, blob)
}

/** Saves a fetched file through a temporary link. The object URL is revoked right after the click. */
export const saveBlobAsFile = (blob: Blob, fileName: string): void => {
  const objectUrl = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = objectUrl
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(objectUrl)
}
