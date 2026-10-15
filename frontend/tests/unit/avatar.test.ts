import { describe, it, expect } from 'vitest'
import {
  AVATAR_MAX_DATA_URL_LENGTH,
  AvatarError,
  compressAvatar,
  coverCrop,
  getDataUrlMime,
  isMimeHonored,
  isUploadableAvatar,
  validateAvatarFile,
  type AvatarEncoder
} from '../../utils/avatar'

const dataUrl = (mime: string, length = 100) =>
  `data:${mime};base64,${'A'.repeat(Math.max(0, length - `data:${mime};base64,`.length))}`

describe('validateAvatarFile', () => {
  it('accepts png, jpeg and webp up to 5 MB', () => {
    for (const type of ['image/png', 'image/jpeg', 'image/webp']) {
      expect(validateAvatarFile({ type, size: 5 * 1024 * 1024 })).toBeNull()
    }
  })

  it('rejects svg and gif', () => {
    expect(validateAvatarFile({ type: 'image/svg+xml', size: 10 })).toBe('invalidType')
    expect(validateAvatarFile({ type: 'image/gif', size: 10 })).toBe('invalidType')
    expect(validateAvatarFile({ type: '', size: 10 })).toBe('invalidType')
  })

  it('rejects files over 5 MB', () => {
    expect(validateAvatarFile({ type: 'image/png', size: 5 * 1024 * 1024 + 1 })).toBe('tooLargeInput')
  })
})

describe('data URL checks', () => {
  it('reads the mime type', () => {
    expect(getDataUrlMime(dataUrl('image/webp'))).toBe('image/webp')
    expect(getDataUrlMime('nope')).toBe('')
  })

  it('detects when webp was not honored', () => {
    expect(isMimeHonored(dataUrl('image/png'), 'image/webp')).toBe(false)
    expect(isMimeHonored(dataUrl('image/webp'), 'image/webp')).toBe(true)
  })

  it('checks format and length', () => {
    expect(isUploadableAvatar(dataUrl('image/jpeg'))).toBe(true)
    expect(isUploadableAvatar(dataUrl('image/gif'))).toBe(false)
    expect(isUploadableAvatar(dataUrl('image/png', AVATAR_MAX_DATA_URL_LENGTH + 1))).toBe(false)
    expect(isUploadableAvatar(dataUrl('image/png', AVATAR_MAX_DATA_URL_LENGTH))).toBe(true)
  })
})

describe('coverCrop', () => {
  it('centers a square on landscape and portrait images', () => {
    expect(coverCrop(400, 200)).toEqual({ sx: 100, sy: 0, side: 200 })
    expect(coverCrop(200, 500)).toEqual({ sx: 0, sy: 150, side: 200 })
  })
})

describe('compressAvatar', () => {
  it('uses webp when the browser honors it', () => {
    const calls: Array<[number, string, number]> = []
    const encoder: AvatarEncoder = {
      encode: (side, mime, quality) => {
        calls.push([side, mime, quality])
        return dataUrl(mime, 5000)
      }
    }
    expect(getDataUrlMime(compressAvatar(encoder))).toBe('image/webp')
    expect(calls).toEqual([[256, 'image/webp', 0.85]])
  })

  it('falls back to jpeg when webp comes back as png (Safari)', () => {
    const encoder: AvatarEncoder = {
      encode: (_side, mime) => dataUrl(mime === 'image/webp' ? 'image/png' : mime, 5000)
    }
    expect(getDataUrlMime(compressAvatar(encoder))).toBe('image/jpeg')
  })

  it('lowers quality and then size until it fits', () => {
    const encoder: AvatarEncoder = {
      encode: (side, mime, quality) => dataUrl(mime, Math.round(side * quality * 1000))
    }
    const out = compressAvatar(encoder)
    expect(out.length).toBeLessThanOrEqual(AVATAR_MAX_DATA_URL_LENGTH)
  })

  it('throws tooLargeOutput when nothing fits', () => {
    const encoder: AvatarEncoder = { encode: (_s, mime) => dataUrl(mime, AVATAR_MAX_DATA_URL_LENGTH + 10) }
    expect(() => compressAvatar(encoder)).toThrow(AvatarError)
    try {
      compressAvatar(encoder)
    } catch (e) {
      expect((e as AvatarError).code).toBe('tooLargeOutput')
    }
  })
})
