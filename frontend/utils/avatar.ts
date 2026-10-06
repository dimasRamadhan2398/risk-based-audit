/**
 * Profile picture helpers.
 *
 * The auth-service accepts `avatar_url` only as a data URL
 * (`data:image/(png|jpeg|webp);base64,...`, <= 300000 chars). The browser
 * therefore center-crops and downscales the chosen file before upload.
 * Validation and size/prefix checks are pure; the canvas work sits behind the
 * `AvatarEncoder` interface so it can be replaced in tests.
 */

export const AVATAR_ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/webp'] as const
export const AVATAR_MAX_INPUT_MB = 5
export const AVATAR_MAX_INPUT_BYTES = AVATAR_MAX_INPUT_MB * 1024 * 1024
export const AVATAR_MAX_SIDE = 256
/** Our own ceiling; the backend rejects anything above 300000 characters */
export const AVATAR_MAX_DATA_URL_LENGTH = 250_000

const DATA_URL_PATTERN = /^data:image\/(png|jpeg|webp);base64,/

export type AvatarErrorCode = 'invalidType' | 'tooLargeInput' | 'readFailed' | 'tooLargeOutput'

export class AvatarError extends Error {
  code: AvatarErrorCode
  constructor(code: AvatarErrorCode) {
    super(code)
    this.name = 'AvatarError'
    this.code = code
  }
}

/** Returns an error code, or null when the file may be used */
export const validateAvatarFile = (
  file: { type: string, size: number },
  maxBytes = AVATAR_MAX_INPUT_BYTES
): AvatarErrorCode | null => {
  if (!(AVATAR_ALLOWED_TYPES as readonly string[]).includes(file.type)) return 'invalidType'
  if (file.size > maxBytes) return 'tooLargeInput'
  return null
}

/** `image/webp` out of `data:image/webp;base64,...`, or '' when it is not a data URL */
export const getDataUrlMime = (dataUrl: string): string => {
  const match = /^data:([^;,]+)[;,]/.exec(dataUrl)
  return match?.[1] ?? ''
}

/** Whether the encoder really produced the requested type (Safari silently returns PNG for webp) */
export const isMimeHonored = (dataUrl: string, requestedMime: string): boolean =>
  getDataUrlMime(dataUrl) === requestedMime

/** Format accepted by the backend and within our size budget */
export const isUploadableAvatar = (dataUrl: string, maxLength = AVATAR_MAX_DATA_URL_LENGTH): boolean =>
  DATA_URL_PATTERN.test(dataUrl) && dataUrl.length <= maxLength

/** Largest square taken from the center of a w x h image (cover crop) */
export const coverCrop = (width: number, height: number) => {
  const side = Math.min(width, height)
  return { sx: Math.floor((width - side) / 2), sy: Math.floor((height - side) / 2), side }
}

/** Draws the source cropped to a `side` x `side` square and returns the data URL for `mime` */
export interface AvatarEncoder {
  encode(side: number, mime: string, quality: number): string
}

const SIDES = [AVATAR_MAX_SIDE, 192, 128, 96, 64]
const QUALITIES = [0.85, 0.75, 0.65, 0.5]

/**
 * Tries webp first, uses jpeg when the browser does not honor webp, and
 * reduces quality, then dimensions, until the data URL fits.
 */
export const compressAvatar = (
  encoder: AvatarEncoder,
  maxLength = AVATAR_MAX_DATA_URL_LENGTH
): string => {
  let mime = 'image/webp'
  for (const side of SIDES) {
    for (const quality of QUALITIES) {
      let dataUrl = encoder.encode(side, mime, quality)
      if (mime === 'image/webp' && !isMimeHonored(dataUrl, mime)) {
        mime = 'image/jpeg'
        dataUrl = encoder.encode(side, mime, quality)
      }
      if (isMimeHonored(dataUrl, mime) && isUploadableAvatar(dataUrl, maxLength)) return dataUrl
    }
  }
  throw new AvatarError('tooLargeOutput')
}

/** Canvas-backed encoder for a browser File */
export const createCanvasEncoder = async (file: Blob): Promise<AvatarEncoder> => {
  let source: CanvasImageSource & { width: number, height: number }
  let release = () => {}
  try {
    if (typeof createImageBitmap === 'function') {
      const bitmap = await createImageBitmap(file)
      source = bitmap
      release = () => bitmap.close()
    } else {
      const url = URL.createObjectURL(file)
      const img = new Image()
      await new Promise<void>((resolve, reject) => {
        img.onload = () => resolve()
        img.onerror = () => reject(new Error('decode'))
        img.src = url
      })
      URL.revokeObjectURL(url)
      source = img
    }
  } catch {
    throw new AvatarError('readFailed')
  }

  const { sx, sy, side: cropSide } = coverCrop(source.width, source.height)
  if (!cropSide) {
    release()
    throw new AvatarError('readFailed')
  }

  return {
    encode(side, mime, quality) {
      const canvas = document.createElement('canvas')
      canvas.width = side
      canvas.height = side
      const ctx = canvas.getContext('2d')
      if (!ctx) throw new AvatarError('readFailed')
      if (mime === 'image/jpeg') {
        // JPEG has no alpha; avoid black backgrounds for transparent PNGs
        ctx.fillStyle = '#ffffff'
        ctx.fillRect(0, 0, side, side)
      }
      ctx.imageSmoothingQuality = 'high'
      ctx.drawImage(source, sx, sy, cropSide, cropSide, 0, 0, side, side)
      return canvas.toDataURL(mime, quality)
    }
  }
}

/** Validates, crops, downscales and encodes a chosen file; throws AvatarError */
export const prepareAvatar = async (file: File): Promise<string> => {
  const invalid = validateAvatarFile(file)
  if (invalid) throw new AvatarError(invalid)
  const encoder = await createCanvasEncoder(file)
  return compressAvatar(encoder)
}
