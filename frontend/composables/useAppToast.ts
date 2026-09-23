import { useToast } from '#imports'

export interface ToastOptions {
  title?: string
  description?: string
  color?: 'success' | 'error' | 'warning' | 'info' | 'primary' | 'secondary' | 'neutral'
  icon?: string
  duration?: number
}

/**
 * Reusable Toast Notification Composable
 * Provides clean helpers for success, error, warning, and info toasts.
 * Duration progress bar is hidden via app.config.ts styling.
 */
export function useAppToast() {
  const toast = useToast()

  const success = (title: string, description?: string, options?: Partial<ToastOptions>) => {
    return toast.add({
      title,
      description,
      color: 'success',
      icon: 'i-lucide-check-circle-2',
      ...options,
    })
  }

  const error = (title: string, description?: string, options?: Partial<ToastOptions>) => {
    return toast.add({
      title,
      description,
      color: 'error',
      icon: 'i-lucide-alert-circle',
      ...options,
    })
  }

  const warning = (title: string, description?: string, options?: Partial<ToastOptions>) => {
    return toast.add({
      title,
      description,
      color: 'warning',
      icon: 'i-lucide-alert-triangle',
      ...options,
    })
  }

  const info = (title: string, description?: string, options?: Partial<ToastOptions>) => {
    return toast.add({
      title,
      description,
      color: 'info',
      icon: 'i-lucide-info',
      ...options,
    })
  }

  return {
    toast,
    success,
    error,
    warning,
    info,
    add: toast.add,
    remove: toast.remove,
    clear: toast.clear,
  }
}
