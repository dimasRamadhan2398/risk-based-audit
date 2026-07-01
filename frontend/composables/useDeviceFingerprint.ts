import type { DeviceInfo } from '~/types/auth'

/**
 * Generates a stable device fingerprint based on browser/device properties.
 * Used for device registration and trusted device detection.
 */
export const useDeviceFingerprint = () => {
  const getDeviceFingerprint = (): DeviceInfo => {
    if (!import.meta.client) {
      return {
        deviceFingerprint: 'server-side',
        deviceName: 'Server',
        deviceType: 'server',
      }
    }

    // Build fingerprint from stable browser properties
    const components = [
      navigator.userAgent,
      navigator.language,
      screen.width + 'x' + screen.height,
      screen.colorDepth,
      new Date().getTimezoneOffset(),
      navigator.hardwareConcurrency ?? 0,
    ]

    // Simple hash of the components
    const raw = components.join('|')
    let hash = 0
    for (let i = 0; i < raw.length; i++) {
      const char = raw.charCodeAt(i)
      hash = (hash << 5) - hash + char
      hash = hash & hash // Convert to 32-bit integer
    }
    const fingerprint = Math.abs(hash).toString(16).padStart(8, '0')

    // Detect device name & type
    const ua = navigator.userAgent
    let deviceType = 'desktop'
    let deviceName = 'Unknown Browser'

    if (/Mobi|Android/i.test(ua)) {
      deviceType = 'mobile'
    } else if (/iPad|Tablet/i.test(ua)) {
      deviceType = 'tablet'
    }

    if (/Chrome/i.test(ua) && !/Chromium|Edge/i.test(ua)) {
      deviceName = 'Chrome'
    } else if (/Firefox/i.test(ua)) {
      deviceName = 'Firefox'
    } else if (/Safari/i.test(ua) && !/Chrome/i.test(ua)) {
      deviceName = 'Safari'
    } else if (/Edge/i.test(ua)) {
      deviceName = 'Edge'
    }

    const os = /Windows/i.test(ua)
      ? 'Windows'
      : /Mac OS X/i.test(ua)
        ? 'macOS'
        : /Linux/i.test(ua)
          ? 'Linux'
          : /Android/i.test(ua)
            ? 'Android'
            : /iOS/i.test(ua)
              ? 'iOS'
              : 'Unknown OS'

    return {
      deviceFingerprint: fingerprint,
      deviceName: `${deviceName} on ${os}`,
      deviceType,
    }
  }

  return { getDeviceFingerprint }
}
