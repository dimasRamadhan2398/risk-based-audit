import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { TableColumn } from '@nuxt/ui'
import { type ActivityPlan, type ActivityPlanFormState, AuditCategory, AuditDepartment } from '~/types/audit';
import { RiskLevel } from '~/types/risk';

export const useActivityPlanStore = defineStore('activity-plan', () => {
  const isModalOpen = ref(false);
  const isViewModalOpen = ref(false);
  const isEditMode = ref(false);
  const loading = ref(false);
  const errorMsg = ref('');

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

  const getInitialFormState = (): ActivityPlanFormState => ({
    planTitle: '',
    planYear: new Date().getFullYear().toString(),
    planPeriodStart: '',
    planPeriodEnd: '',
    department: AuditDepartment.IT,
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

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const filteredPlans = computed(() => {
    return plans.value.map(plan => ({
      ...plan,
      period: `${plan.planPeriodStart} - ${plan.planPeriodEnd}`,
      totalActivity: (plan.plannedActivities || []).length,
      totalAuditor: (plan.resourceAuditors || []).length,
      budgetEstimation: plan.budget?.totalEstimatedCost || 0,
      budgetAllocated: plan.budget?.totalAllocatedBudget || 0,
    }));
  });

  const columns: TableColumn<ActivityPlan>[] = [
    { accessorKey: 'planTitle', header: 'Title' },
    { accessorKey: 'planYear', header: 'Year' },
    { accessorKey: 'period', header: 'Period' },
    { accessorKey: 'department', header: 'Department/Unit' },
    { accessorKey: 'riskName', header: 'Risk Name' },
    { accessorKey: 'riskLevel', header: 'Risk Level' },
    { accessorKey: 'attachments', header: 'Attachment' },
    { accessorKey: 'actions', header: 'Actions' }
  ]

  const fetchPlans = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/activity-plans`, {
        method: 'GET'
      });
      if (response && response.data && Array.isArray(response.data.items)) {
        plans.value = response.data.items
      } else if (response && Array.isArray(response.items)) {
        plans.value = response.items;
      } else if (Array.isArray(response)) {
        plans.value = response;
      }
    } catch (error: any) {
      console.error('Failed to fetch activity plans:', error);
      errorMsg.value = 'Failed to load activity plans.';
    } finally {
      loading.value = false;
    }
  }

  function openModal() {
    isEditMode.value = false;
    formState.value = getInitialFormState();
    isModalOpen.value = true;
  }

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
    if (!confirm("Are you sure you want to delete this activity plan?")) return;
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/activity-plans/${id}`, {
        method: 'DELETE'
      });
      await fetchPlans();
    } catch (error: any) {
      console.error('Failed to delete activity plan:', error);
      errorMsg.value = 'Failed to delete activity plan.';
    } finally {
      loading.value = false;
    }
  };

  const savePlan = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();

      // Transform planned activities to match backend expectations if needed
      // but for now we follow existing structure and just ensure fields are there

      const fileList = formState.value.file && formState.value.file.length > 0 
        ? formState.value.file.map((f: any) => ({
            name: f.name,
            size: Math.round(f.size / 1024) + ' KB',
            url: '#'
          }))
        : [];

      const payload = {
        ...formState.value,
        attachments: isEditMode.value ? (formState.value.attachments || []).concat(fileList) : fileList
      };

      if (isEditMode.value) {
        const planId = (formState.value as ActivityPlan).id;
        await $fetch(`${baseUrl}/activity-plans/${planId}`, {
          method: 'PUT',
          body: payload
        });
      } else {
        await $fetch(`${baseUrl}/activity-plans`, {
          method: 'POST',
          body: payload
        });

        for (const act of formState.value.plannedActivities) {
           await $fetch(`${baseUrl}/audit-activities`, {
             method: 'POST',
             body: {
                title: act.auditName,
                engagement_subject: act.auditee,
                audit_type: act.category,
                justification: act.priority, // or mapped differently
                audit_purpose: 'Standard Audit',
                team_size: act.numberOfAuditors,
                status: 'PLANNED'
             }
           });
        }
      }
      closeModal();
      await fetchPlans();
    } catch (error: any) {
      console.error('Failed to save activity plan:', error);
      errorMsg.value = 'Failed to save activity plan.';
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
      budgetEstimation: ''
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
    openModal, closeModal, openViewModal, closeViewModal, handleEdit, handleDelete, savePlan,
    addPlannedActivity, removePlannedActivity, addResourceAuditor, removeResourceAuditor,
    fetchPlans, loading, errorMsg, getRiskLevelColor, attachmentCategoryOptions, handleFileChange
  };
});
