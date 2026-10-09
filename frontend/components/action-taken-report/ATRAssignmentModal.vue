<template>
  <AtrModalShell
    v-model:open="store.showAssignment"
    :title="item?.pic_user_id ? t('actionTakenReport.assignment.titleReassign') : t('actionTakenReport.assignment.title')"
    :subtitle="item?.finding_title || ''"
    :busy="store.saving"
  >
    <template
      v-if="item"
      #meta
    >
      <AtrStatusBadge
        :status="item.status"
        :is-overdue="item.is_overdue"
        :overdue-days="item.overdue_days"
        size="md"
      />
    </template>

    <template v-if="item">
      <UAlert
        v-if="usersForbidden"
        color="error"
        variant="subtle"
        icon="i-lucide-shield-alert"
        :title="t('actionTakenReport.assignment.usersForbiddenTitle')"
        :description="t('actionTakenReport.assignment.usersForbidden')"
      />
      <UAlert
        v-else-if="usersError"
        color="error"
        variant="subtle"
        icon="i-lucide-alert-triangle"
        :title="t('actionTakenReport.errors.loadUsers')"
        :description="usersError"
        :actions="[{ label: t('actionTakenReport.table.retry'), color: 'neutral', variant: 'outline', onClick: () => loadCandidates(searchTerm) }]"
      />

      <AppFormField
        :label="t('actionTakenReport.assignment.pic')"
        :description="t('actionTakenReport.assignment.picHelp')"
        name="pic_user_id"
        required
        :error="errors.pic_user_id ? t(errors.pic_user_id) : undefined"
      >
        <template #default="{ id }">
          <ReusableSelectMenu
            :id="id"
            v-model="form.pic_user_id"
            v-model:search-term="searchTerm"
            :items="picItems"
            value-key="value"
            ignore-filter
            :loading="usersLoading"
            :disabled="usersForbidden || store.saving"
            :placeholder="t('actionTakenReport.assignment.picPlaceholder')"
            :search-input="{ placeholder: t('actionTakenReport.assignment.picSearch') }"
            icon="i-lucide-user"
          >
            <template #empty>
              <span class="text-sm text-gray-500">
                {{ usersLoading ? t('actionTakenReport.table.loading') : t('actionTakenReport.assignment.noUsers') }}
              </span>
            </template>
          </ReusableSelectMenu>
        </template>
      </AppFormField>

      <AppFormField
        :label="t('actionTakenReport.assignment.dueDate')"
        name="due_date"
        required
        :error="errors.due_date ? t(errors.due_date) : undefined"
      >
        <template #default="{ id }">
          <AppDatePicker
            :id="id"
            v-model="form.due_date"
            :disabled="store.saving"
            :placeholder="t('actionTakenReport.assignment.dueDatePlaceholder')"
          />
        </template>
      </AppFormField>
    </template>

    <template #footer>
      <div class="flex flex-col-reverse sm:flex-row sm:justify-end gap-2">
        <UButton
          :label="t('actionTakenReport.modal.close')"
          color="neutral"
          variant="ghost"
          class="justify-center"
          :disabled="store.saving"
          @click="store.closeModal()"
        />
        <UButton
          :label="t('actionTakenReport.assignment.save')"
          icon="i-lucide-user-check"
          color="primary"
          class="justify-center"
          :disabled="!canAssign || usersForbidden"
          :loading="store.saving"
          @click="save"
        />
      </div>
    </template>
  </AtrModalShell>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useActionTakenReportStore, type AtrPicCandidate } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { useAtrActions } from '~/composables/useAtrActions'
import { atrAssignmentSchema, atrDateOnly, atrFieldErrors } from '~/utils/actionTakenReport'
import AppFormField from '~/components/shared/AppFormField.vue'
import AppDatePicker from '~/components/shared/AppDatePicker.vue'
import ReusableSelectMenu from '~/components/shared/ReusableSelectMenu.vue'
import AtrModalShell from '~/components/action-taken-report/AtrModalShell.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'

const store = useActionTakenReportStore()
const { t } = useI18n()
const { availableActions } = useAtrActions()

const item = computed(() => store.current)
const canAssign = computed(() => availableActions(item.value).assign)

const form = reactive({ pic_user_id: '', due_date: '' })
const errors = ref<Record<string, string>>({})

// --- PIC candidates: server-side search on GET /users/assignable --------------------------------
const candidates = ref<AtrPicCandidate[]>([])
const usersLoading = ref(false)
const usersForbidden = ref(false)
const usersError = ref('')
const searchTerm = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null
let searchRequest = 0

const candidateLabel = (u: AtrPicCandidate) => {
  const meta = [u.position, u.department].filter(Boolean).join(', ')
  return meta ? `${u.full_name} — ${meta}` : u.full_name
}

const picItems = computed(() => {
  const list = candidates.value.map(u => ({ label: candidateLabel(u), value: u.id }))
  // Keep the current PIC selectable/labelled even when the search result does not include them.
  const pic = item.value
  if (pic?.pic_user_id && !list.some(i => i.value === pic.pic_user_id)) {
    list.unshift({ label: pic.pic_name || pic.pic_user_id, value: pic.pic_user_id })
  }
  return list
})

const loadCandidates = async (term: string) => {
  const id = ++searchRequest
  usersLoading.value = true
  usersError.value = ''
  const result = await store.fetchPicCandidates(term)
  if (id !== searchRequest) return
  candidates.value = result.users
  usersForbidden.value = result.forbidden
  usersError.value = result.error
  usersLoading.value = false
}

watch(searchTerm, (term) => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    searchTimer = null
    loadCandidates(term)
  }, 300)
})

onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
})

// Reset the form for the item being assigned each time the modal opens (and when the server copy arrives).
watch(
  () => [store.showAssignment, item.value?.id] as const,
  ([open], previous) => {
    if (!open || !item.value) return
    const wasOpen = previous?.[0]
    form.pic_user_id = item.value.pic_user_id ?? ''
    form.due_date = atrDateOnly(item.value.due_date) ?? ''
    errors.value = {}
    if (!wasOpen) {
      searchTerm.value = ''
      loadCandidates('')
    }
  },
  { immediate: true }
)

const save = async () => {
  if (!item.value) return
  const result = atrAssignmentSchema.safeParse({ pic_user_id: form.pic_user_id ?? '', due_date: form.due_date ?? '' })
  errors.value = atrFieldErrors(result)
  if (!result.success) return
  if (await store.assign(item.value.id, result.data)) store.closeModal()
}
</script>
