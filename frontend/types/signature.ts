/**
 * Value of the reusable signature box (components/shared/SignatureBox.vue).
 * `image` is a PNG data URL drawn on the canvas, or '' when nothing is signed.
 * `signedAt` is an ISO timestamp set when the first stroke is drawn, '' when cleared.
 */
export interface SignatureBoxValue {
  image: string
  name: string
  title: string
  signedAt: string
}

/**
 * Body of POST /executive-summaries/:id/approve (audit-service).
 * signatureDate is YYYY-MM-DD; signatureImage must be a PNG or JPEG data URL (max 2 MB decoded).
 */
export interface ExecutiveSummaryApprovePayload {
  signaturePlace: string
  signatureDate: string
  signatureImage: string
  signerName: string
  signerRole: string
}
