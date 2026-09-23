<template>
  <div class="flex flex-col gap-1.5 items-start w-full">
    <UBadge 
      v-for="(member, idx) in formattedMembers"
      :key="member.id || member.name || idx"
      :color="color" 
      :variant="variant" 
      :size="size"
      class="flex flex-col items-start px-2 py-1 max-w-full"
    >
      <span class="font-bold text-primary truncate max-w-full">{{ member.name }}</span>
      <span v-if="member.role" class="text-[10px] opacity-70 italic truncate max-w-full">{{ member.role }}</span>
    </UBadge>
    <span v-if="!formattedMembers.length" class="text-gray-400">-</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

export interface TeamMemberItem {
  id?: string | number
  name?: string
  role?: string
  [key: string]: unknown
}

const props = withDefaults(
  defineProps<{
    members?: (TeamMemberItem | string)[] | null
    color?: string
    variant?: string
    size?: string
  }>(),
  {
    members: () => [],
    color: 'primary',
    variant: 'outline',
    size: 'lg'
  }
)

const formattedMembers = computed<TeamMemberItem[]>(() => {
  if (!props.members || !Array.isArray(props.members)) return []
  return props.members
    .filter(Boolean)
    .map((member) => {
      if (typeof member === 'string') {
        return { name: member, role: '' }
      }
      return member
    })
})
</script>
