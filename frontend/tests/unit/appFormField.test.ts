import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import AppFormField from '../../components/shared/AppFormField.vue'

describe('AppFormField', () => {
  const mountComponent = (props: Record<string, unknown> = {}, slots: Record<string, unknown> = {}) => {
    const container = document.createElement('div')
    document.body.appendChild(container)
    const app = createApp({
      render() {
        return h(AppFormField, props, slots)
      }
    })
    // Stub UIcon component
    app.component('UIcon', {
      props: ['name'],
      render() {
        return h('span', { class: this.name })
      }
    })
    const vm = app.mount(container)
    return {
      container,
      vm,
      destroy: () => {
        app.unmount()
        container.remove()
      }
    }
  }

  it('renders normal label within 100 characters', () => {
    const title = 'Document Title'
    const { container, destroy } = mountComponent({ label: title })

    expect(container.textContent).toContain('Document Title')
    const label = container.querySelector('label')
    expect(label?.textContent?.trim()).toBe('Document Title')
    expect(label?.getAttribute('title')).toBeNull() // Not truncated
    destroy()
  })

  it('enforces maximum 100 characters for form field title and truncates with ellipsis', () => {
    // 120 character string
    const longTitle = 'A'.repeat(120)
    const { container, destroy } = mountComponent({ label: longTitle })

    const label = container.querySelector('label')
    // Should be truncated to 100 chars + '...'
    expect(label?.textContent?.trim()).toBe('A'.repeat(100) + '...')
    // Full raw title preserved in tooltip/title attribute
    expect(label?.getAttribute('title')).toBe(longTitle)

    // Warning badge showing over limit
    expect(container.textContent).toContain('120/100 max')
    destroy()
  })

  it('supports alias prop title instead of label', () => {
    const title = 'Audit Plan Title'
    const { container, destroy } = mountComponent({ title })

    expect(container.textContent).toContain('Audit Plan Title')
    destroy()
  })

  it('does not truncate when label is exactly 100 characters', () => {
    const title = 'B'.repeat(100)
    const { container, destroy } = mountComponent({ label: title })

    const label = container.querySelector('label')
    expect(label?.textContent?.trim()).toBe('B'.repeat(100))
    expect(container.textContent).not.toContain('max')
    destroy()
  })

  it('renders required asterisk when required is true', () => {
    const { container, destroy } = mountComponent({ label: 'Required Field', required: true })

    expect(container.textContent).toContain('*')
    const asterisk = container.querySelector('[aria-hidden="true"]')
    expect(asterisk?.textContent?.trim()).toBe('*')
    destroy()
  })

  it('renders optional badge when optional is true', () => {
    const { container, destroy } = mountComponent({ label: 'Optional Field', optional: true })

    expect(container.textContent).toContain('Optional')
    destroy()
  })

  it('renders custom optional string when provided', () => {
    const { container, destroy } = mountComponent({ label: 'Field', optional: '(Tidak Wajib)' })

    expect(container.textContent).toContain('(Tidak Wajib)')
    destroy()
  })

  it('displays character counter when counter or showCount is enabled', () => {
    const { container, destroy } = mountComponent({
      label: 'Title',
      counter: true,
      maxCount: 100,
      modelValue: 'Short title'
    })

    expect(container.textContent).toContain('11')
    expect(container.textContent).toContain('/100')
    destroy()
  })

  it('displays error message when error prop is provided', () => {
    const errorMsg = 'Title is required'
    const { container, destroy } = mountComponent({ label: 'Title', error: errorMsg })

    const errorEl = container.querySelector('[role="alert"]')
    expect(errorEl?.textContent).toContain(errorMsg)
    destroy()
  })

  it('renders info tooltip when tooltip or info prop is provided', () => {
    const tooltipText = 'Enter the official audit document title'
    const { container, destroy } = mountComponent({ label: 'Title', tooltip: tooltipText })

    expect(container.innerHTML).toContain(tooltipText)
    destroy()
  })

  it('renders custom slot contents for default input', () => {
    const { container, destroy } = mountComponent(
      { label: 'Title' },
      {
        default: () => h('input', { id: 'custom-input', placeholder: 'Enter text...' })
      }
    )

    const input = container.querySelector('#custom-input')
    expect(input).not.toBeNull()
    expect(input?.getAttribute('placeholder')).toBe('Enter text...')
    destroy()
  })

  it('renders custom actions slot', () => {
    const { container, destroy } = mountComponent(
      { label: 'Title' },
      {
        actions: () => h('button', { id: 'clear-btn' }, 'Clear')
      }
    )

    const btn = container.querySelector('#clear-btn')
    expect(btn).not.toBeNull()
    expect(btn?.textContent).toBe('Clear')
    destroy()
  })
})

describe('ReusableFormField', () => {
  it('imports and renders properly as an alias of AppFormField', async () => {
    const ReusableFormField = (await import('../../components/shared/ReusableFormField.vue')).default
    const container = document.createElement('div')
    document.body.appendChild(container)
    const app = createApp({
      render() {
        return h(ReusableFormField, { label: 'Reusable Title', required: true })
      }
    })
    app.component('UIcon', {
      props: ['name'],
      render() {
        return h('span', { class: this.name })
      }
    })
    app.mount(container)

    expect(container.textContent).toContain('Reusable Title')
    expect(container.textContent).toContain('*')
    app.unmount()
    container.remove()
  })
})
