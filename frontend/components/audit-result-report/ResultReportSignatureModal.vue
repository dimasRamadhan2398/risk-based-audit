<template>
  <UModal
    :open="open"
    dismissible
    :ui="{
      content: 'sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-6 space-y-6 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(92vh-140px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'bg-gray-900/60 dark:bg-black/85 backdrop-blur-md'
    }"
    @update:open="emit('update:open', $event)"
  >
    <template #content>
      <div class="flex flex-col h-full max-h-[92vh] bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center bg-slate-50/70 dark:bg-gray-900">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-xl bg-primary-100 dark:bg-primary-950/60 text-primary-600 dark:text-primary-400">
              <UIcon name="i-heroicons-pencil-square" class="size-6" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white"> {{ t('auditResultReport.editor.signature.title') }} </h3>
              <p class="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-2 mt-0.5">
                <span>{{ t('auditResultReport.editor.signature.assignmentLetter') }} <strong class="text-primary-600 dark:text-primary-400">{{ assignmentLetterId || '-' }}</strong></span>
                <span class="text-gray-300 dark:text-gray-600">•</span>
                <span>{{ t('auditResultReport.editor.signature.reportNumber') }} <strong class="font-mono text-gray-700 dark:text-gray-300">{{ reportNumber || '-' }}</strong></span>
              </p>
            </div>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark"
            @click="emit('update:open', false)"
          />
        </div>

        <!-- Body -->
        <div class="p-6 overflow-y-auto flex-1 space-y-6 bg-white dark:bg-gray-900">
          <!-- Notice Banner -->
          <div class="p-4 bg-primary-50/70 dark:bg-primary-950/40 rounded-xl border border-primary-200 dark:border-primary-900/60 flex items-start gap-3">
            <UIcon name="i-heroicons-information-circle" class="size-5 text-primary-600 dark:text-primary-400 shrink-0 mt-0.5" />
            <div class="text-xs text-slate-700 dark:text-slate-300 leading-relaxed">
              {{ t('auditResultReport.editor.signature.introPrefix') }} <strong class="text-slate-900 dark:text-white">{{ t('auditResultReport.editor.signature.member') }}</strong> {{ t('auditResultReport.editor.signature.introMiddle') }} <strong class="text-primary-700 dark:text-primary-300">{{ assignmentLetterId }}</strong>.
              {{ t('auditResultReport.editor.signature.introAfterLetter') }} <strong class="text-slate-900 dark:text-white">{{ t('auditResultReport.editor.signature.lastPage') }}</strong> {{ t('auditResultReport.editor.signature.introSuffix') }}
            </div>
          </div>

          <!-- Metadata Pengesahan: Tempat, Tanggal, Perusahaan -->
          <div class="p-5 bg-slate-50/80 dark:bg-gray-800/60 rounded-xl border border-slate-200 dark:border-gray-800 space-y-4">
            <h4 class="text-xs font-bold uppercase tracking-wider text-gray-600 dark:text-gray-400 flex items-center gap-2">
              <UIcon name="i-heroicons-map-pin" class="size-4 text-primary-600 dark:text-primary-400" /> {{ t('auditResultReport.editor.signature.approvalInfo') }} </h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <UFormField :label="t('executiveSummaryForm.signature.place')" required>
                <UInput
                  v-model="localPlace"
                  :placeholder="t('executiveSummaryForm.signature.placePlaceholder')"
                  class="w-full"
                  required
                />
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.signature.date')" required>
                <AppDatePicker
                  v-model="localDate"
                  class="w-full"
                  required
                />
              </UFormField>

              <UFormField :label="t('masterData.company.columns.name')" required>
                <UInput
                  v-model="localCompanyName"
                  :placeholder="t('auditResultReport.editor.signature.companyExample')"
                  class="w-full"
                  required
                />
              </UFormField>
            </div>
          </div>

          <!-- Section: Tanda Tangan Team Member -->
          <div class="space-y-4">
            <div class="flex justify-between items-center">
              <div>
                <h4 class="text-sm font-bold text-gray-900 dark:text-white uppercase tracking-wide flex items-center gap-2">
                  <UIcon name="i-heroicons-user-group" class="text-primary-600 dark:text-primary-400" /> {{ t('auditResultReport.editor.signature.memberSignatures') }} <UBadge color="primary" variant="subtle" size="sm">
                    {{ memberSignatures.length }} {{ t('workingPaper.headerForm.member') }} </UBadge>
                </h4>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5"> {{ t('auditResultReport.editor.signature.automaticList') }} </p>
              </div>
            </div>

            <!-- Member Signature Cards Grid -->
            <div class="space-y-4">
              <div
                v-for="(member, idx) in memberSignatures"
                :key="idx"
                class="p-5 bg-white dark:bg-gray-800/60 rounded-xl border border-slate-200 dark:border-gray-700/70 shadow-sm space-y-3 hover:border-primary-300 dark:hover:border-primary-500/60 transition"
              >
                <!-- Member Header Info -->
                <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-3 border-b border-gray-100 dark:border-gray-700/60">
                  <div class="flex items-center gap-3">
                    <div class="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-950/80 border border-primary-200 dark:border-primary-800/80 flex items-center justify-center text-primary-700 dark:text-primary-300 font-bold text-xs">
                      {{ idx + 1 }}
                    </div>
                    <div>
                      <div class="flex items-center gap-2">
                        <span class="font-bold text-sm text-gray-900 dark:text-white">{{ member.name }}</span>
                        <UBadge :color="getRoleBadgeColor(member.role)" variant="subtle" size="xs">
                          {{ member.role || 'Team Member' }}
                        </UBadge>
                      </div>
                      <span class="text-[11px] text-gray-400 dark:text-gray-500">{{ t('auditResultReport.editor.signature.memberNumber') }}{{ idx + 1 }}</span>
                    </div>
                  </div>
                </div>

                <!-- Signature Pad -->
                <SignaturePad
                  v-model="member.signature"
                  :default-name="member.name"
                  :height="130"
                />
              </div>

              <!-- Empty State if no team members -->
              <div
                v-if="memberSignatures.length === 0"
                class="text-center py-8 bg-slate-50/80 dark:bg-gray-800/40 rounded-xl border border-dashed border-slate-300 dark:border-gray-700/60 space-y-2"
              >
                <UIcon name="i-heroicons-user-group" class="size-8 text-gray-400 dark:text-gray-500 mx-auto" />
                <p class="text-sm font-semibold text-gray-700 dark:text-gray-300"> {{ t('auditResultReport.editor.signature.empty') }} </p>
                <p class="text-xs text-gray-500 dark:text-gray-400 max-w-sm mx-auto"> {{ t('auditResultReport.editor.signature.emptyDescription') }} </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-100 dark:border-gray-800 flex justify-between items-center bg-slate-50/70 dark:bg-gray-900">
          <span class="text-xs text-gray-500 dark:text-gray-400 mr-2 hidden sm:inline">
            {{ signedCount }} {{ t('auditResultReport.editor.signature.of') }} {{ memberSignatures.length }} {{ t('auditResultReport.editor.signature.completed') }} </span>

          <div class="flex items-center gap-2">
            <UButton
              :label="t('auditResultReport.editor.signature.back')"
              color="neutral"
              variant="ghost"
              @click="emit('update:open', false)"
            />
            <UButton
              :label="t('auditResultReport.editor.signature.confirm')"
              color="primary"
              icon="i-heroicons-check"
              :loading="loading"
              @click="handleConfirmSave"
            />
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
const { t } = useI18n()
import { ref, computed, watch } from 'vue'
import SignaturePad from '~/components/shared/SignaturePad.vue'
import type { MemberSignature } from '~/stores/audit-result-report'

interface TeamMemberItem {
  name: string
  role?: string
}

const props = defineProps<{
  open: boolean
  assignmentLetterId?: string
  reportNumber?: string
  companyName?: string
  initialPlace?: string
  initialDate?: string
  existingSignatures?: MemberSignature[]
  teamMembers?: TeamMemberItem[]
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (
    e: 'confirm',
    payload: {
      signaturePlace: string
      signatureDate: string
      companyName: string
      signatures: MemberSignature[]
    }
  ): void
}>()

const localPlace = ref(props.initialPlace || 'Jakarta')
const localDate = ref(props.initialDate || new Date().toISOString().split('T')[0])
const localCompanyName = ref(props.companyName || 'PT AIFL Indonesia')
const memberSignatures = ref<MemberSignature[]>([])

const signedCount = computed(() => {
  return memberSignatures.value.filter(m => Boolean(m.signature)).length
})

const getRoleBadgeColor = (role?: string) => {
  const r = (role || '').toLowerCase()
  if (r.includes('chair') || r.includes('ketua') || r.includes('lead')) return 'primary'
  if (r.includes('super') || r.includes('pengawas')) return 'warning'
  if (r.includes('pic') || r.includes('charge')) return 'info'
  return 'neutral'
}

// Sync signatures when modal opens
const syncSignatures = () => {
  localPlace.value = props.initialPlace || 'Jakarta'
  localDate.value = props.initialDate || new Date().toISOString().split('T')[0]
  localCompanyName.value = props.companyName || 'PT AIFL Indonesia'

  const existingMap = new Map((props.existingSignatures || []).map(s => [s.name, s]))
  const result: MemberSignature[] = []

  // Team members are strictly derived from Assignment Letter
  if (props.teamMembers && props.teamMembers.length > 0) {
    props.teamMembers.forEach(tm => {
      const existing = existingMap.get(tm.name)
      result.push({
        name: tm.name,
        role: tm.role || 'Member',
        signature: existing?.signature || '',
        signedAt: existing?.signedAt || ''
      })
    })
  } else if (props.existingSignatures && props.existingSignatures.length > 0) {
    // If existing signatures exist on already-created report, retain them
    props.existingSignatures.forEach(s => {
      result.push({
        name: s.name,
        role: s.role || 'Member',
        signature: s.signature || '',
        signedAt: s.signedAt || ''
      })
    })
  }

  memberSignatures.value = result
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      syncSignatures()
    }
  },
  { immediate: true }
)

const handleConfirmSave = () => {
  const cleanedSignatures: MemberSignature[] = memberSignatures.value.map(m => ({
    name: m.name,
    role: m.role || 'Member',
    signature: m.signature || '',
    signedAt: m.signature ? (m.signedAt || new Date().toISOString()) : ''
  }))

  emit('confirm', {
    signaturePlace: localPlace.value.trim() || 'Jakarta',
    signatureDate: localDate.value || new Date().toISOString().split('T')[0],
    companyName: localCompanyName.value.trim() || 'PT AIFL Indonesia',
    signatures: cleanedSignatures
  })
}
</script>
