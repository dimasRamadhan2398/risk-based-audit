import { defineStore } from 'pinia'

export interface GlobalModalOptions {
  itemName?: string;
  title?: string;
  description?: string;
}

export const useGlobalModalStore = defineStore('globalModal', {
  state: () => ({
    isOpen: false,
    options: null as GlobalModalOptions | null,
    resolvePromise: null as ((value: boolean) => void) | null,
  }),
  actions: {
    confirmDelete(options?: GlobalModalOptions | string): Promise<boolean> {
      let opts: GlobalModalOptions = {}
      if (typeof options === 'string') {
        opts.itemName = options
      } else if (options) {
        opts = options
      }
      
      this.options = opts
      this.isOpen = true
      
      return new Promise((resolve) => {
        this.resolvePromise = resolve
      })
    },
    resolve(result: boolean) {
      if (this.resolvePromise) {
        this.resolvePromise(result)
        this.resolvePromise = null
      }
      this.isOpen = false
    }
  }
})
