import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { TableColumn } from '@nuxt/ui'
import { type ActivityPlan, type ActivityPlanFormState, AuditCategory, AuditDepartment } from '~/types/audit';
import { RiskLevel } from '~/types/risk';

export const useActivityPlanStore = defineStore('activity-plan', () => {
  const isModalOpen = ref(false);
  const isViewModalOpen = ref(false);
  const isEditMode = ref(false);

  const executionStatusOptions = [
    { label: "Planned", value: "planned" },
    { label: "In Progress", value: "in_progress" },
    { label: "Completed", value: "completed" }
  ];

  // Helper untuk memformat key enum menjadi label yang mudah dibaca
  const formatEnumKey = (key: string) => {
    return key
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ');
  };

  // Buat opsi secara dinamis dari enum RiskLevel untuk konsistensi
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
    }
  });

  const formState = ref<ActivityPlanFormState>(getInitialFormState());

  const plans = ref<ActivityPlan[]>([]);
  const selectedPlan = ref<ActivityPlan | null>(null);
  const filteredPlans = computed(() => {
    return plans.value.map(plan => ({
      ...plan,
      period: `${plan.planPeriodStart} - ${plan.planPeriodEnd}`,
      totalActivity: plan.plannedActivities.length,
      totalAuditor: plan.resourceAuditors.length,
      budgetEstimation: plan.budget.totalEstimatedCost,
      budgetAllocated: plan.budget.totalAllocatedBudget,
    }));
  });

  const columns: TableColumn<ActivityPlan>[] = [
    { accessorKey: 'planTitle', header: 'Title' },
    { accessorKey: 'planYear', header: 'Year' },
    { accessorKey: 'period', header: 'Period' },
    { accessorKey: 'department', header: 'Department/Unit' },
    { accessorKey: 'totalActivity', header: 'Total Auditor Team' },
    { accessorKey: 'totalAuditor', header: 'Total Auditor' },
    { accessorKey: 'budgetEstimation', header: 'Budget Estimation' },
    { accessorKey: 'budgetAllocated', header: 'Budget Allocated' },
    { accessorKey: 'createdBy', header: 'Created By' },
    { accessorKey: 'actions', header: 'Actions' }
  ]

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

  const handleDelete = (id: string) => {
    if (confirm("Are you sure you want to delete this activity plan?")) {
      plans.value = plans.value.filter(
        (item) => item.id !== id
      );
    }
  };

  function savePlan() {
    if (isEditMode.value) {
      const index = plans.value.findIndex(p => p.id === (formState.value as ActivityPlan).id);
      if (index !== -1) {
        plans.value[index] = { ...formState.value, id: (formState.value as ActivityPlan).id, status: 'Updated', createdAt: (formState.value as ActivityPlan).createdAt } as ActivityPlan;
      }
    } else {
      plans.value.push({
        ...formState.value,
        id: Date.now().toString(),
        status: 'Submitted',
        createdAt: new Date().toISOString()
      });
    }
    closeModal();
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

  return {
    isModalOpen, isViewModalOpen, isEditMode, priorityOptions, riskLevelOptions,
    formState, plans, selectedPlan, columns, filteredPlans,
    openModal, closeModal, openViewModal, closeViewModal, handleEdit, handleDelete, savePlan,
    addPlannedActivity, removePlannedActivity, addResourceAuditor, removeResourceAuditor
  };
});
