import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { TableColumn } from '@nuxt/ui'
import type { ActivityPlan, ActivityPlanFormState, PlannedAuditActivity, ResourceAuditor } from '~/types/audit';

export const useActivityPlanStore = defineStore('activity-plan', () => {
  const isModalOpen = ref(false);
  const isViewModalOpen = ref(false);
  const isEditMode = ref(false);

  const getInitialFormState = (): ActivityPlanFormState => ({
    planTitle: '',
    planYear: new Date().getFullYear().toString(),
    planPeriodStart: '',
    planPeriodEnd: '',
    department: '',
    createdBy: 'Jamil Simamora',
    creationDate: new Date().toISOString().split('T')[0]!,
    plannedActivities: [],
    resourceAuditors: [],
    budget: {
      totalEstimatedCost: 0,
      totalAllocatedBudget: 0,
      budgetNotes: ''
    },
    review: {
      creatorName: 'Dimas',
      creatorPosition: 'CAE',
      approverName: '',
      approverPosition: '',
      approvalDate: '',
      additionalNotes: ''
    }
  });

  const formState = ref<ActivityPlanFormState>(getInitialFormState());

  const plans = ref<ActivityPlan[]>([]);
  const selectedPlan = ref<ActivityPlan | null>(null);
  const filteredPlans = ref<ActivityPlan[]>([])

  const columns: TableColumn<ActivityPlan>[] = [
    { accessorKey: 'planTitle', header: 'Judul Rencana' },
    { accessorKey: 'planYear', header: 'Tahun' },
    { accessorKey: 'department', header: 'Departemen/Unit' },
    { accessorKey: 'createdBy', header: 'Dibuat Oleh' },
    { accessorKey: 'creationDate', header: 'Tanggal' },
    { accessorKey: 'actions', header: 'Aksi' }
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
      auditor: '',
      area: '',
      executionStatus: 'Planned',
      riskLevel: 'Rendah',
      duration: 0,
      priority: 'Low',
      numberOfAuditors: 1,
      estimatedSchedule: ''
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
      department: '',
      availability: ''
    });
  }

  function removeResourceAuditor(index: number) {
    formState.value.resourceAuditors.splice(index, 1);
  }

  return {
    isModalOpen,
    isViewModalOpen,
    isEditMode,
    formState,
    plans,
    selectedPlan,
    columns,
    filteredPlans,
    openModal,
    closeModal,
    openViewModal,
    closeViewModal,
    handleEdit,
    savePlan,
    addPlannedActivity,
    removePlannedActivity,
    addResourceAuditor,
    removeResourceAuditor
  };
});
