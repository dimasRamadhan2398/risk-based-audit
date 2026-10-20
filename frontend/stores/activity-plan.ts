import { defineStore } from 'pinia';
import { ref, computed, reactive } from 'vue';
import {
  ACTIVITY_PLAN_ACTIVITY_TYPES,
  AnnualAuditPlanStatus,
  type ActivityPlanActivityType,
  type ActivityPlanDetail,
  type ActivityPlanFormState,
  type ActivityPlanSummary,
  type AnnualActivityOption,
  type AnnualAuditAttachment,
  type AnnualAuditPlan
} from '~/types/audit';
import type { Employee } from '~/types/master';
import { useAuthStore } from '~/stores/auth';
import { useAnnualPlanStore } from '~/stores/annual-audit';
import { useGlobalModalStore } from '~/stores/global-modal';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';
import { useI18n } from '~/composables/useI18n';
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';
import { useEmployeeApi } from '~/composables/useEmployeeApi';
import { extractErrorMessage, getUserErrorMessage } from '~/utils/error';
import { viewDocumentInNewTab } from '~/utils/documentViewer';
import {
  ACTIVITY_TYPE_LABEL_KEYS,
  AUDIT_KIND_SUGGESTIONS,
  buildActivityPlanPayload,
  categoryMatchesActivityType,
  emptyActivityPlanForm,
  formFromDetail,
  normalizeActivityPlanDetail,
  normalizeActivityPlanSummary,
  normalizeAnnualActivityOption,
  parseApiError,
  rowFromOption,
  stepForApiError,
  sumFormTotals,
  translateApiError,
  validateActivityPlanHeader,
  validateActivityPlanRows,
  type ActivityPlanIssue
} from '~/utils/activityPlanForm';

// The list endpoint caps page_size at 100.
const PLAN_PAGE_SIZE = 100;

/** i18n prefix of the /activity-plans error codes (section 6.7). */
const ERROR_PREFIX = 'auditActivityPlan.errors';

/** A blocking error of the form: shown in the step alert, with the step (and row) to focus. */
export interface ActivityPlanFormError {
  messages: string[];
  step?: number;
  row?: number;
  code?: string;
}

/**
 * Audit Activity Plan (Formulir Rencana Kegiatan Audit). Talks to audit-service /activity-plans
 * (docs/audit-activity-plan-rework.md, section 6). Fields owned by the Annual Audit Plan (PKAT)
 * are shown read-only and never sent.
 */
export const useActivityPlanStore = defineStore('activity-plan', () => {
  const { t } = useI18n();
  const annualPlanStore = useAnnualPlanStore();
  const toast = useToastNotification();

  // ── List ──────────────────────────────────────────────────────
  const plans = ref<ActivityPlanSummary[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');
  const filters = reactive<{ year?: number, activityType?: ActivityPlanActivityType }>({});

  const activityTypeLabel = (type?: string | null) =>
    type && ACTIVITY_TYPE_LABEL_KEYS[type as ActivityPlanActivityType] ? t(ACTIVITY_TYPE_LABEL_KEYS[type as ActivityPlanActivityType]) : (type || '');

  const activityTypeOptions = computed(() =>
    ACTIVITY_PLAN_ACTIVITY_TYPES.map(value => ({ value, label: activityTypeLabel(value) }))
  );

  /**
   * Loads every page of GET /activity-plans with the current filters. The first page tells the
   * total_pages; the rest are fetched in parallel.
   */
  const fetchPlans = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const query = {
        page_size: PLAN_PAGE_SIZE,
        year: filters.year || undefined,
        activity_type: filters.activityType || undefined
      };
      const fetchPage = (page: number): Promise<any> => $fetch(`${baseUrl}/activity-plans`, {
        method: 'GET',
        query: { ...query, page }
      });
      const first = await fetchPage(1);
      const totalPages = Number(first?.data?.pagination?.total_pages) || 1;
      const rest = totalPages > 1
        ? await Promise.all(Array.from({ length: totalPages - 1 }, (_, i) => fetchPage(i + 2)))
        : [];
      plans.value = [first, ...rest]
        .flatMap((res: any) => (Array.isArray(res?.data?.items) ? res.data.items : []))
        .map(normalizeActivityPlanSummary);
    } catch (error) {
      console.error('Failed to fetch activity plans:', error);
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'auditActivityPlan.loadError' });
    } finally {
      loading.value = false;
    }
  };

  const clearFilters = () => {
    filters.year = undefined;
    filters.activityType = undefined;
  };

  const fetchDetail = async (id: string): Promise<ActivityPlanDetail> => {
    const res: any = await $fetch(`${getAuditServiceBaseUrl()}/activity-plans/${id}`, { method: 'GET' });
    return normalizeActivityPlanDetail(res?.data ?? {});
  };

  // ── View ──────────────────────────────────────────────────────
  const isViewModalOpen = ref(false);
  const selectedPlan = ref<ActivityPlanDetail | null>(null);
  const viewingId = ref<string | null>(null);
  const viewLoading = ref(false);
  const viewError = ref('');

  const loadView = async () => {
    const id = viewingId.value;
    if (!id) return;
    viewLoading.value = true;
    viewError.value = '';
    try {
      const detail = await fetchDetail(id);
      // The user may have opened another plan meanwhile.
      if (viewingId.value === id) selectedPlan.value = detail;
    } catch (error) {
      console.error('Failed to load activity plan:', error);
      if (viewingId.value === id) viewError.value = getUserErrorMessage(error, t, { fallbackKey: 'auditActivityPlan.loadError' });
    } finally {
      if (viewingId.value === id) viewLoading.value = false;
    }
  };

  function openViewModal(plan: Pick<ActivityPlanSummary, 'id'>) {
    viewingId.value = plan.id;
    selectedPlan.value = null;
    isViewModalOpen.value = true;
    loadView();
  }

  function closeViewModal() {
    isViewModalOpen.value = false;
    viewingId.value = null;
    selectedPlan.value = null;
    viewError.value = '';
  }

  // ── Form ──────────────────────────────────────────────────────
  const isModalOpen = ref(false);
  const isEditMode = ref(false);
  const formState = ref<ActivityPlanFormState>(emptyActivityPlanForm());
  const formLoading = ref(false);
  const formLoadError = ref('');
  const editingId = ref<string | null>(null);
  const saving = ref(false);
  const formError = ref<ActivityPlanFormError | null>(null);

  const formTotals = computed(() => sumFormTotals(formState.value.rows));

  // PKAT rows that can be added (GET /activity-plans/annual-activity-options).
  const annualActivityOptions = ref<AnnualActivityOption[]>([]);
  const optionsLoading = ref(false);
  const optionsError = ref('');

  // PKAT choices for a new plan: deactivated PKATs cannot get a new Audit Activity Plan.
  const annualPlansLoading = ref(false);
  const annualPlanOptions = computed(() =>
    annualPlanStore.plans
      .filter((plan: AnnualAuditPlan) => plan.id && plan.isActive !== false)
      .map((plan: AnnualAuditPlan) => ({
        value: plan.id as string,
        label: [plan.year, plan.code, plan.title, annualPlanStore.getStatusLabel(plan.status)].filter(Boolean).join(' · '),
        plan
      }))
  );

  /** The PKAT of the form, for Tahun and the status warning. */
  const selectedAnnualPlan = computed(() => {
    const form = formState.value;
    if (!form.annualPlanId) return null;
    const fromList = annualPlanStore.plans.find((p: AnnualAuditPlan) => p.id === form.annualPlanId);
    if (fromList) {
      return { id: fromList.id as string, code: fromList.code, title: fromList.title || '', year: Number(fromList.year) || 0, status: fromList.status, isActive: fromList.isActive !== false };
    }
    return form.annualPlan;
  });

  const formYear = computed(() => selectedAnnualPlan.value?.year || 0);

  /** Non-blocking: the PKAT is not approved yet (drafting is still allowed). */
  const annualPlanNotApproved = computed(() =>
    !!selectedAnnualPlan.value && selectedAnnualPlan.value.status !== AnnualAuditPlanStatus.APPROVED
  );

  const loadAnnualPlans = async () => {
    annualPlansLoading.value = true;
    try {
      await annualPlanStore.fetchPlans();
    } finally {
      annualPlansLoading.value = false;
    }
  };

  // Ketua Tim picker (master-service employees).
  const employees = ref<Employee[]>([]);
  const employeesLoading = ref(false);
  const employeesError = ref('');

  const loadEmployees = async () => {
    employeesLoading.value = true;
    employeesError.value = '';
    try {
      const list = await useEmployeeApi().getAllEmployees();
      employees.value = (Array.isArray(list) ? list : [])
        .filter(e => e?.id && e.full_name && e.is_active !== false)
        .sort((a, b) => a.full_name.localeCompare(b.full_name));
    } catch (error) {
      console.error('Failed to load employees for the Ketua Tim picker:', error);
      employeesError.value = getUserErrorMessage(error, t, { fallbackKey: 'auditActivityPlan.rows.teamLeaderLoadError' });
    } finally {
      employeesLoading.value = false;
    }
  };

  const auditKindItems = computed(() =>
    AUDIT_KIND_SUGGESTIONS.map(s => ({ value: s.value, label: t(s.labelKey) }))
  );

  const resetForm = () => {
    formState.value = emptyActivityPlanForm();
    formError.value = null;
    formLoadError.value = '';
    annualActivityOptions.value = [];
    optionsError.value = '';
  };

  function openModal() {
    isEditMode.value = false;
    editingId.value = null;
    resetForm();
    isModalOpen.value = true;
    loadAnnualPlans();
    loadEmployees();
  }

  /** Loads the saved plan into the form. Legacy plans cannot be edited. */
  const loadEditing = async () => {
    const id = editingId.value;
    if (!id) return;
    formLoading.value = true;
    formLoadError.value = '';
    formError.value = null;
    try {
      const detail = await fetchDetail(id);
      if (editingId.value !== id) return;
      if (detail.isLegacy) {
        formLoadError.value = t('auditActivityPlan.errors.ACTIVITY_PLAN_LEGACY_READONLY');
        return;
      }
      formState.value = formFromDetail(detail);
    } catch (error) {
      console.error('Failed to load activity plan for editing:', error);
      if (editingId.value === id) formLoadError.value = getUserErrorMessage(error, t, { fallbackKey: 'auditActivityPlan.loadError' });
    } finally {
      if (editingId.value === id) formLoading.value = false;
    }
  };

  function handleEdit(plan: Pick<ActivityPlanSummary, 'id' | 'isLegacy'>) {
    if (plan.isLegacy) {
      toast.showWarning(t('auditActivityPlan.errors.ACTIVITY_PLAN_LEGACY_READONLY'), t('auditActivityPlan.legacy.notice'));
      return;
    }
    isEditMode.value = true;
    editingId.value = plan.id;
    resetForm();
    isModalOpen.value = true;
    loadEditing();
    loadAnnualPlans();
    loadEmployees();
  }

  function closeModal() {
    isModalOpen.value = false;
    isEditMode.value = false;
    editingId.value = null;
  }

  /**
   * Changes the PKAT of a new plan. Rows belong to the PKAT, so they are cleared after a
   * confirmation. The PKAT of a saved plan cannot change.
   */
  const setAnnualPlan = async (annualPlanId: string | undefined) => {
    const form = formState.value;
    const next = annualPlanId || '';
    if (isEditMode.value || next === form.annualPlanId) return;
    if (form.rows.length) {
      const confirmed = await useGlobalModalStore().confirmSubmit({
        description: t('auditActivityPlan.header.changeAnnualPlanConfirm'),
        confirmLabel: t('auditActivityPlan.header.changeAnnualPlanConfirmLabel')
      });
      if (!confirmed) return;
    }
    form.annualPlanId = next;
    form.annualPlan = null;
    form.rows = [];
    annualActivityOptions.value = [];
  };

  /**
   * Changes the Jenis Kegiatan Audit. Allowed only while every row's PKAT category still matches
   * the new type (the server rejects it otherwise). Returns whether it changed.
   */
  const setActivityType = (activityType: ActivityPlanActivityType | undefined): boolean => {
    const form = formState.value;
    const next = activityType || '';
    if (next === form.activityType) return true;
    if (next && form.rows.some(row => !categoryMatchesActivityType(row.pkat.auditCategory, next))) {
      formError.value = { messages: [t('auditActivityPlan.validation.activityTypeRowsMismatch')], step: 0 };
      return false;
    }
    if (!next && form.rows.length) {
      formError.value = { messages: [t('auditActivityPlan.validation.activityTypeRowsMismatch')], step: 0 };
      return false;
    }
    form.activityType = next;
    annualActivityOptions.value = [];
    if (formError.value?.step === 0) formError.value = null;
    return true;
  };

  const fetchAnnualActivityOptions = async () => {
    const form = formState.value;
    if (!form.annualPlanId || !form.activityType) {
      annualActivityOptions.value = [];
      return;
    }
    optionsLoading.value = true;
    optionsError.value = '';
    try {
      const res: any = await $fetch(`${getAuditServiceBaseUrl()}/activity-plans/annual-activity-options`, {
        method: 'GET',
        query: {
          annual_plan_id: form.annualPlanId,
          activity_type: form.activityType,
          exclude_plan_id: form.id || undefined
        }
      });
      const items = Array.isArray(res?.data?.items) ? res.data.items : [];
      annualActivityOptions.value = items.map(normalizeAnnualActivityOption);
    } catch (error) {
      console.error('Failed to load PKAT activities:', error);
      annualActivityOptions.value = [];
      optionsError.value = translateApiError(error, t, ERROR_PREFIX, 'auditActivityPlan.picker.loadError');
    } finally {
      optionsLoading.value = false;
    }
  };

  /** Adds the picked PKAT rows (in PKAT order), skipping ones already in the form or taken. */
  const addRowsFromOptions = (options: AnnualActivityOption[]) => {
    const present = new Set(formState.value.rows.map(r => r.annualActivityId));
    for (const option of options) {
      if (!option.available || present.has(option.annualActivityId)) continue;
      formState.value.rows.push(rowFromOption(option));
      present.add(option.annualActivityId);
    }
  };

  const removeRow = (index: number) => {
    const row = formState.value.rows[index];
    if (!row || row.inUse) return;
    formState.value.rows.splice(index, 1);
  };

  const moveRow = (index: number, direction: -1 | 1) => {
    const rows = formState.value.rows;
    const target = index + direction;
    if (index < 0 || index >= rows.length || target < 0 || target >= rows.length) return;
    const [row] = rows.splice(index, 1);
    rows.splice(target, 0, row!);
  };

  /** Blocking issues of a step (0 header, 1 rows, 2 attachments) as translated messages. */
  const validateStep = (step: number): ActivityPlanFormError | null => {
    let issues: ActivityPlanIssue[] = [];
    if (step === 0) issues = validateActivityPlanHeader(formState.value);
    else if (step === 1) issues = validateActivityPlanRows(formState.value, t);
    if (!issues.length) return null;
    return {
      messages: issues.map(issue => t(issue.key, issue.params)),
      step,
      row: issues.find(issue => issue.row)?.row
    };
  };

  // ── Attachments ───────────────────────────────────────────────
  const attachmentCategoryOptions = ['Plan', 'Evidence', 'Charter', 'Other'];

  // Stores the files with the audit-service media endpoint. The local provider
  // returns a /uploads/... path that the frontend proxies; Google Drive returns
  // a full link. One folder per save keeps same-named files from overwriting.
  const uploadAttachments = async (files: File[]): Promise<AnnualAuditAttachment[]> => {
    const baseUrl = getAuditServiceBaseUrl();
    const folder = `Auditsphere/activity-plans/${Date.now()}`;
    const uploaded: AnnualAuditAttachment[] = [];
    for (const file of files) {
      const body = new FormData();
      body.append('file', file);
      body.append('folder', folder);
      const failed = t('auditActivityPlan.form.uploadFailed', { name: file.name });
      let res: any;
      try {
        res = await $fetch(`${baseUrl}/media/upload`, { method: 'POST', body });
      } catch (error) {
        throw new Error(`${failed} ${extractErrorMessage(error, '')}`.trim());
      }
      if (!res?.data?.filePath) throw new Error(failed);
      uploaded.push({
        name: file.name,
        size: Math.round(file.size / 1024) + ' KB',
        url: res.data.filePath
      });
    }
    return uploaded;
  };

  /** Opens an attachment in a new tab, fetched as a blob so the request carries the token. */
  const openAttachment = async (file: AnnualAuditAttachment) => {
    try {
      await viewDocumentInNewTab(() => $fetch(file.url, { responseType: 'blob' }), file.name, file.name);
    } catch (error) {
      console.error('Failed to open attachment:', error);
      toast.showError(t('auditActivityPlan.view.openAttachmentError'), getUserErrorMessage(error, t));
    }
  };

  // ── Save / delete ─────────────────────────────────────────────

  /**
   * Validates every step, uploads new files and sends the section 6.4 payload. A server error is
   * mapped by its code to a message, with the step and row to focus. Resolves true when saved.
   */
  const savePlan = async (): Promise<boolean> => {
    for (const step of [0, 1]) {
      const error = validateStep(step);
      if (error) {
        formError.value = error;
        return false;
      }
    }
    saving.value = true;
    formError.value = null;
    try {
      const form = formState.value;
      const newFiles = form.file || [];
      let uploaded: AnnualAuditAttachment[] = [];
      try {
        uploaded = newFiles.length ? await uploadAttachments(newFiles) : [];
      } catch (error) {
        formError.value = { messages: [(error as Error).message], step: 2 };
        return false;
      }
      const payload = buildActivityPlanPayload(form, [...form.attachments, ...uploaded]);
      const baseUrl = getAuditServiceBaseUrl();
      if (isEditMode.value && form.id) {
        await $fetch(`${baseUrl}/activity-plans/${form.id}`, { method: 'PUT', body: payload });
      } else {
        await $fetch(`${baseUrl}/activity-plans`, { method: 'POST', body: payload });
      }
      closeModal();
      toast.showSuccess(t('auditActivityPlan.saveSuccess'));
      await fetchPlans();
      return true;
    } catch (error) {
      console.error('Failed to save activity plan:', error);
      const parsed = parseApiError(error);
      formError.value = {
        messages: [translateApiError(error, t, ERROR_PREFIX, 'auditActivityPlan.saveError')],
        step: stepForApiError(parsed),
        row: parsed.details.row,
        code: parsed.code
      };
      return false;
    } finally {
      saving.value = false;
    }
  };

  /** Deletes a plan after confirmation. Legacy plans can always be deleted. */
  const handleDelete = async (plan: Pick<ActivityPlanSummary, 'id' | 'title'>) => {
    const confirmed = await useGlobalModalStore().confirmDelete({
      description: t('auditActivityPlan.deleteConfirm'),
      body: plan.title ? [plan.title] : undefined
    });
    if (!confirmed) return;
    loading.value = true;
    try {
      await $fetch(`${getAuditServiceBaseUrl()}/activity-plans/${plan.id}`, { method: 'DELETE' });
      if (viewingId.value === plan.id) closeViewModal();
      toast.showSuccess(t('auditActivityPlan.deleteSuccess'));
      await fetchPlans();
    } catch (error) {
      console.error('Failed to delete activity plan:', error);
      toast.showError(t('auditActivityPlan.deleteError'), translateApiError(error, t, ERROR_PREFIX, 'auditActivityPlan.deleteError'));
    } finally {
      loading.value = false;
    }
  };

  /** The signed-in user (the server sets createdBy from the token; shown on a new plan). */
  const currentUserName = computed(() => useAuthStore().user?.fullName || '');

  return {
    // list
    plans, loading, errorMsg, filters, fetchPlans, clearFilters, activityTypeOptions, activityTypeLabel,
    // view
    isViewModalOpen, selectedPlan, viewLoading, viewError, openViewModal, closeViewModal, loadView,
    // form
    isModalOpen, isEditMode, formState, formLoading, formLoadError, saving, formError, formTotals,
    annualPlanOptions, annualPlansLoading, selectedAnnualPlan, formYear, annualPlanNotApproved,
    employees, employeesLoading, employeesError, auditKindItems, currentUserName,
    openModal, closeModal, handleEdit, loadEditing, setAnnualPlan, setActivityType,
    annualActivityOptions, optionsLoading, optionsError, fetchAnnualActivityOptions,
    addRowsFromOptions, removeRow, moveRow, validateStep, savePlan, handleDelete,
    // attachments
    attachmentCategoryOptions, uploadAttachments, openAttachment
  };
});
