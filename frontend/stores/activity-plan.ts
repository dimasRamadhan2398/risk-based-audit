import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { TableColumn } from '@nuxt/ui'
import { type ActivityPlan, type ActivityPlanFormState, type AnnualAuditAttachment, AuditCategory } from '~/types/audit';
import { useAuthStore } from '~/stores/auth';
import { RiskLevel } from '~/types/risk';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';
import { formatPeriod } from '~/utils/dateConverter';
import { useI18n } from '~/composables/useI18n';
import { extractErrorMessage } from '~/utils/error';
import { getAuditServiceBaseUrl, getRiskServiceBaseUrl } from '~/composables/useApiUrl';
import { useDepartmentApi } from '~/composables/useDepartmentApi';
import { useRiskProfileStore } from '~/stores/risk-profile';

/** An entry of the "Associated Risk" dropdown, built from GET /risks. */
export interface RiskOption {
  name: string;
  category: string;
  riskLevel: RiskLevel;
}

// The generic list endpoint caps page_size at 100.
const PLAN_PAGE_SIZE = 100;

/** Number input value → number; "", null and unparseable values become 0. */
const toNumber = (value: unknown): number => {
  const n = typeof value === 'number' ? value : parseFloat(String(value ?? '').replace(/[,\s]/g, ''));
  return Number.isFinite(n) ? n : 0;
};

export const useActivityPlanStore = defineStore('activity-plan', () => {
  const { t, locale } = useI18n();
  const isModalOpen = ref(false);
  const isViewModalOpen = ref(false);
  const isEditMode = ref(false);
  const loading = ref(false);
  const errorMsg = ref('');
  const toast = useToastNotification()

  const executionStatusOptions = [
    { label: "Planned", value: "planned" },
    { label: "In Progress", value: "in_progress" },
    { label: "Completed", value: "completed" }
  ];

  const formatEnumKey = (key: string) => {
    return key
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ');
  };

  const riskLevelOptions = Object.entries(RiskLevel).map(([key, value]) => ({
    label: formatEnumKey(key),
    value: value,
  }));

  const priorityOptions = [
    { label: "P1", value: "p1" },
    { label: "P2", value: "p2" },
    { label: "P3", value: "p3" }
  ];

  // ── Dropdown data from the APIs ───────────────────────────────
  // Departments come from master-service and risks from risk-service. There is
  // no mock fallback: an empty or failed response leaves the dropdown empty.
  const departmentNames = ref<string[]>([]);
  const riskOptions = ref<RiskOption[]>([]);
  const optionsLoading = ref(false);

  // A plan saved with an older value (e.g. "IT" from the former fixed list)
  // keeps it selectable while it is being edited.
  const departmentOptions = computed(() => {
    const current = formState.value.department;
    return current && !departmentNames.value.includes(current)
      ? [current, ...departmentNames.value]
      : departmentNames.value;
  });

  const fetchRiskOptions = async (): Promise<RiskOption[]> => {
    try {
      const response: any = await $fetch(`${getRiskServiceBaseUrl()}/risks`);
      const items: any[] = Array.isArray(response?.data) ? response.data : [];
      const { getRiskLevel } = useRiskProfileStore();
      return items
        .filter(r => r?.name)
        .map(r => ({
          name: r.name,
          category: r.category || '',
          // /risks has no stored level; derive it from the risk's own likelihood × impact.
          riskLevel: getRiskLevel(Number(r.likelihood), Number(r.impact))
        }));
    } catch (error) {
      console.error('Failed to fetch risks for the activity plan form:', error);
      return [];
    }
  };

  const fetchFormOptions = async () => {
    optionsLoading.value = true;
    try {
      const [departments, risks] = await Promise.all([
        useDepartmentApi().getAllDepartments(),
        fetchRiskOptions()
      ]);
      departmentNames.value = departments
        .filter(d => d.is_active !== false && d.department_name)
        .map(d => d.department_name)
        .sort((a, b) => a.localeCompare(b));
      riskOptions.value = risks;
    } finally {
      optionsLoading.value = false;
    }
  };

  const getInitialFormState = (): ActivityPlanFormState => ({
    planTitle: '',
    planYear: new Date().getFullYear().toString(),
    planPeriodStart: '',
    planPeriodEnd: '',
    department: '',
    createdBy: '',
    creationDate: new Date().toISOString().split('T')[0]!,
    plannedActivities: [],
    resourceAuditors: [],
    budget: {
      totalEstimatedCost: 0,
      totalAllocatedBudget: 0,
      budgetNotes: ''
    },
    review: {
      creatorName: '',
      creatorPosition: '',
      approverName: '',
      approverPosition: '',
      approvalDate: '',
      additionalNotes: ''
    },
    attachmentCategory: '',
    attachmentUploadedBy: '',
    attachmentUploadDate: '',
    attachments: [],
    file: []
  });

  const formState = ref<ActivityPlanFormState>(getInitialFormState());

  const plans = ref<ActivityPlan[]>([]);
  const selectedPlan = ref<ActivityPlan | null>(null);



  const filteredPlans = computed(() => {
    return plans.value.map(plan => ({
      ...plan,
      period: formatPeriod(plan.planPeriodStart, plan.planPeriodEnd, locale.value),
      totalActivity: (plan.plannedActivities || []).length,
      totalAuditor: (plan.resourceAuditors || []).length,
      budgetEstimation: plan.budget?.totalEstimatedCost || 0,
      budgetAllocated: plan.budget?.totalAllocatedBudget || 0,
    }));
  });

  const columns: (TableColumn<ActivityPlan> & { class?: string })[] = [
    { accessorKey: 'planTitle', header: 'Title', class: 'max-w-[280px] whitespace-normal break-words font-medium' },
    { accessorKey: 'period', header: 'Period', class: 'min-w-[220px] whitespace-nowrap' },
    { accessorKey: 'department', header: 'Department/Unit', class: 'w-36' },
    { accessorKey: 'riskName', header: 'Risk Name', class: 'w-48' },
    { accessorKey: 'riskLevel', header: 'Risk Level', class: 'w-28 whitespace-nowrap' },
    { accessorKey: 'attachments', header: 'Attachment', class: 'w-28 whitespace-nowrap text-center' },
    { accessorKey: 'actions', header: 'Actions', class: 'w-28 whitespace-nowrap text-center' }
  ]

  const fetchPlans = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      // Without page/page_size the endpoint returns only the first 20 plans.
      const fetchPage = (page: number): Promise<any> => $fetch(`${baseUrl}/activity-plans`, {
        method: 'GET',
        query: { page, page_size: PLAN_PAGE_SIZE }
      });
      const first = await fetchPage(1);
      const totalPages = Number(first?.data?.pagination?.total_pages) || 1;
      const rest = totalPages > 1
        ? await Promise.all(Array.from({ length: totalPages - 1 }, (_, i) => fetchPage(i + 2)))
        : [];
      plans.value = [first, ...rest].flatMap((res: any) =>
        Array.isArray(res?.data?.items) ? res.data.items : Array.isArray(res?.items) ? res.items : []
      );
    } catch (error: any) {
      console.error('Failed to fetch activity plans:', error);
      errorMsg.value = extractErrorMessage(error, 'Failed to load activity plans.');
    } finally {
      loading.value = false;
    }
  }

  function openModal() {
    isEditMode.value = false;
    const form = getInitialFormState();
    // The creator is whoever is signed in, not a free-text name.
    const user = useAuthStore().user;
    form.createdBy = user?.fullName || '';
    form.review.creatorName = user?.fullName || '';
    form.review.creatorPosition = user?.position || '';
    formState.value = form;
    isModalOpen.value = true;
  }

  // Derived from the activities so it can never disagree with them.
  const totalEstimatedCost = computed(() =>
    formState.value.plannedActivities.reduce((sum, activity) => sum + toNumber(activity.budgetEstimation), 0)
  );

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

  function closeModal() {
    isModalOpen.value = false;
    isEditMode.value = false;
  }

  function openViewModal(plan: ActivityPlan) {
    selectedPlan.value = plan;
    isViewModalOpen.value = true;
  }

  function closeViewModal() {
    isViewModalOpen.value = false;
    selectedPlan.value = null;
  }

  function handleEdit(plan: ActivityPlan) {
    isEditMode.value = true;
    formState.value = JSON.parse(JSON.stringify(plan));
    isModalOpen.value = true;
  }

  const handleDelete = async (id: string) => {
    const { t } = useI18n();
    if (!await useGlobalModalStore().confirmDelete({ description: t("auditActivityPlan.deleteConfirm") })) return;
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/activity-plans/${id}`, {
        method: 'DELETE'
      });
      await fetchPlans();
      toast.showSuccess('Activity plan deleted successfully.');
    } catch (error: any) {
      console.error('Failed to delete activity plan:', error);
      const detail = extractErrorMessage(error, 'Failed to delete activity plan.');
      errorMsg.value = detail;
      toast.showError('Failed to delete activity plan.', detail);
    } finally {
      loading.value = false;
    }
  };

  const savePlan = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();

      const newFiles = formState.value.file || [];
      const fileList = newFiles.length > 0 ? await uploadAttachments(newFiles) : [];
      if (fileList.length > 0) {
        formState.value.attachmentUploadedBy = useAuthStore().user?.fullName || '';
        formState.value.attachmentUploadDate = new Date().toISOString().split('T')[0]!;
      }

      // Only the fields of the backend ActivityPlan model. Numbers are sent as
      // numbers: a cleared number input yields "", which the create endpoint
      // rejects and the update endpoint would store, breaking the plan list.
      const form = formState.value;
      const payload = {
        planTitle: form.planTitle,
        planYear: form.planYear,
        planPeriodStart: form.planPeriodStart,
        planPeriodEnd: form.planPeriodEnd,
        department: form.department,
        createdBy: form.createdBy,
        creationDate: form.creationDate,
        plannedActivities: form.plannedActivities.map(activity => ({
          ...activity,
          duration: toNumber(activity.duration),
          numberOfAuditors: toNumber(activity.numberOfAuditors),
          budgetEstimation: toNumber(activity.budgetEstimation)
        })),
        resourceAuditors: form.resourceAuditors,
        budget: {
          ...form.budget,
          totalEstimatedCost: totalEstimatedCost.value,
          totalAllocatedBudget: toNumber(form.budget.totalAllocatedBudget)
        },
        review: form.review,
        attachmentCategory: form.attachmentCategory,
        attachmentUploadedBy: form.attachmentUploadedBy,
        attachmentUploadDate: form.attachmentUploadDate,
        attachments: isEditMode.value ? (form.attachments || []).concat(fileList) : fileList
      };

      if (isEditMode.value) {
        const planId = (formState.value as ActivityPlan).id;
        await $fetch(`${baseUrl}/activity-plans/${planId}`, {
          method: 'PUT',
          body: payload
        });
      } else {
        // Planned activities are stored inside the plan itself. /audit-activities is a
        // different resource (annual plan, target unit, project code, dates) this
        // form does not collect, so it is not written from here.
        await $fetch(`${baseUrl}/activity-plans`, {
          method: 'POST',
          body: payload
        });
      }
      closeModal();
      toast.showSuccess('Activity plan saved successfully.');
      await fetchPlans();
    } catch (error: any) {
      console.error('Failed to save activity plan:', error);
      const detail = extractErrorMessage(error, 'Failed to save activity plan.');
      errorMsg.value = detail;
      toast.showError('Failed to save activity plan.', detail);
    } finally {
      loading.value = false;
    }
  }

  function addPlannedActivity() {
    formState.value.plannedActivities.push({
      id: Date.now().toString(),
      auditName: '',
      auditee: '',
      category: AuditCategory.ASSURANCE,
      riskLevel: RiskLevel.LOW,
      duration: 0,
      priority: '',
      numberOfAuditors: 1,
      estimatedSchedule: '',
      budgetEstimation: 0
    });
  }

  function removePlannedActivity(index: number) {
    formState.value.plannedActivities.splice(index, 1);
  }

  function addResourceAuditor() {
    formState.value.resourceAuditors.push({
      id: Date.now().toString(),
      name: '',
      position: '',
      competence: '',
      availability: ''
    });
  }

  function removeResourceAuditor(index: number) {
    formState.value.resourceAuditors.splice(index, 1);
  }

  const attachmentCategoryOptions = ['Plan', 'Evidence', 'Charter', 'Other'];

  const handleFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      errorMsg.value = "File terlalu besar! Maksimal 5MB.";
      formState.value.file = null;
      target.value = "";
      return;
    }

    const allowedTypes = [
      "application/pdf",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      "application/msword",
    ];

    if (!allowedTypes.includes(file.type)) {
      errorMsg.value = "Format file tidak valid. Gunakan PDF atau DOCX.";
      formState.value.file = null;
      target.value = "";
      return;
    }

    errorMsg.value = "";
    formState.value.file = [file];
  };

  const getRiskLevelColor = (level?: string) => {
    if (!level) return 'neutral'
    const lvl = level.toLowerCase()
    if (lvl.includes('high')) return 'error'
    if (lvl.includes('mod') || lvl.includes('medium')) return 'warning'
    if (lvl.includes('low')) return 'success'
    return 'neutral'
  }

  return {
    isModalOpen, isViewModalOpen, isEditMode, priorityOptions, riskLevelOptions,
    formState, plans, selectedPlan, columns, filteredPlans,
    openModal, closeModal, openViewModal, closeViewModal, handleEdit, handleDelete, savePlan, totalEstimatedCost,
    addPlannedActivity, removePlannedActivity, addResourceAuditor, removeResourceAuditor,
    fetchPlans, loading, errorMsg, getRiskLevelColor, attachmentCategoryOptions, handleFileChange,
    departmentOptions, riskOptions, optionsLoading, fetchFormOptions
  };
});
