// Stand-ins for Nuxt's #components virtual module, which only exists in a Nuxt build.
import { defineComponent, h } from 'vue'

export const NuxtLink = defineComponent({
  name: 'NuxtLink',
  props: ['to'],
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.())
})
