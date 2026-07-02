import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import type { TableColumn } from "@nuxt/ui";
import type { StrategicAuditPlan } from "~/types/audit";

export const useStrategicPlanStore = defineStore('strategic-audit-plan', () => {

    const isAddModalOpen = ref(false);
    const isEditMode = ref(false);
    const loading = ref(false);
    const errorMsg = ref('');

    const unitOptions = [
        { label: 'Percentage (%)', value: '%' },
        { label: 'Rupiah (Rp)', value: 'Rp' },
        { label: 'Amount', value: 'Amount' },
        { label: 'Score', value: 'Score' },
        { label: 'Hour', value: 'Hour' },
        { label: 'Day', value: 'Day' },
    ];

    const currentYear = new Date().getFullYear();
    const yearOptions = Array.from({ length: 20 }, (_, i) => {
        const year = currentYear - 5 + i;
        return { label: String(year), value: year };
    });

    const form = ref<Partial<StrategicAuditPlan>>({
        code: '',
        strategicObjective: '',
        kpi: '',
        unit: '',
        hibHig: 'HIG',
        periodType: 'Quartal',
        selectedPeriod: 'Q1',
        yearStart: currentYear,
        yearEnd: currentYear + 4,
        kpiTargets: {},
        internalAuditSO: '',
        actual: '',
        target: '',
        calculation: '',
        status: '',
    });

    const availablePeriods = computed(() => {
        if (form.value.periodType === 'Quartal') {
            return ['Q1', 'Q2', 'Q3', 'Q4'];
        } else {
            const yearStart = form.value.yearStart || currentYear;
            const yearEnd = form.value.yearEnd || currentYear + 4;
            const years: string[] = [];
            for (let y = yearStart; y <= yearEnd; y++) {
                years.push(String(y));
            }
            return years;
        }
    });

    const computedCalculation = computed(() => {
        const actual = parseFloat(form.value.actual || '0');
        const target = parseFloat(form.value.target || '0');
        if (!form.value.actual || !form.value.target) return '';

        let result = 0;
        if (form.value.hibHig === 'HIG') {
            if (target === 0) return '';
            result = (actual / target) * 100;
        } else {
            if (actual === 0) return '';
            result = (target / actual) * 100;
        }
        
        return `${result.toFixed(2)}%`;
    });

    const computedStatus = computed(() => {
        const actual = parseFloat(form.value.actual || '0');
        const target = parseFloat(form.value.target || '0');
        if (!form.value.actual || !form.value.target) return '';

        let ratio = 0;
        if (form.value.hibHig === 'HIG') {
            if (target === 0) return '';
            ratio = actual / target;
        } else {
            if (actual === 0) return '';
            ratio = target / actual;
        }

        if (ratio >= 1) return 'Good';
        if (ratio >= 0.7) return 'Moderate';
        return 'Poor';
    });

    const mockObjectives: StrategicAuditPlan[] = [
        {
            id: '1',
            code: 'SO-IA01',
            strategicObjective: 'Improve Audit Efficiency',
            kpi: 'Revenue Operational Cost',
            unit: '%',
            hibHig: 'HIG',
            periodType: 'Quartal',
            selectedPeriod: 'Q1',
            kpiTargets: { 2024: '300', 2025: '350', 2026: '400', 2027: '450', 2028: '500' },
            internalAuditSO: 'Optimize resource allocation',
            actual: '100',
            target: '300',
            calculation: '33.33%',
            status: 'Poor',
        },
        {
            id: '2',
            code: 'SO-IA02',
            strategicObjective: 'Strengthen Internal Control',
            kpi: 'Customer Satisfaction Index',
            unit: 'Score',
            hibHig: 'HIG',
            kpiTargets: {},
            internalAuditSO: '',
            periodType: 'Yearly',
            selectedPeriod: '2025',
            yearStart: 2022,
            yearEnd: 2026,
            actual: '85',
            target: '90',
            calculation: '94.44%',
            status: 'Good',
        },
        {
            id: '3',
            code: 'SO-IA03',
            strategicObjective: 'Improve Compliance',
            kpi: 'Audit Response Time',
            unit: 'Hour',
            hibHig: 'HIB',
            kpiTargets: {},
            internalAuditSO: '',
            periodType: 'Quartal',
            selectedPeriod: 'Q2',
            actual: '24',
            target: '48',
            calculation: '50.00%',
            status: 'Good',
        },
    ];

    const strategicObjectives = ref<StrategicAuditPlan[]>([...mockObjectives]);

    const getAuditServiceBaseUrl = () => {
        const config = useRuntimeConfig()
        return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
    }

    const fetchStrategicPlans = async () => {
        loading.value = true;
        errorMsg.value = '';
        try {
            const baseUrl = getAuditServiceBaseUrl();
            const response: any = await $fetch(`${baseUrl}/strategic-plans`, {
                method: 'GET'
            });
            let items: StrategicAuditPlan[] = [];
            if (response && response.data && Array.isArray(response.data.items)) {
                items = response.data.items;
            } else if (response && Array.isArray(response.items)) {
                items = response.items;
            } else if (Array.isArray(response)) {
                items = response;
            }

            if (items.length > 0) {
                strategicObjectives.value = items;
            } else {
                strategicObjectives.value = [];
            }
        } catch (error: any) {
            console.error('Failed to fetch strategic plans:', error);
            errorMsg.value = 'Failed to load strategic plans.';
            strategicObjectives.value = [];
        } finally {
            loading.value = false;
        }
    }

    // Load on init
    fetchStrategicPlans();

    const columns: TableColumn<StrategicAuditPlan>[] = [
        {
            accessorKey: 'code',
            header: 'Objective ID',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'strategicObjective',
            header: 'Strategic Objective',
        },
        {
            accessorKey: 'kpi',
            header: 'KPI Name',
        },
        {
            accessorKey: 'unit',
            header: 'Unit',
        },
        {
            accessorKey: 'selectedPeriod',
            header: 'Period',
        },
        {
            accessorKey: 'target',
            header: 'Target',
        },
        {
            accessorKey: 'actual',
            header: 'Actual',
        },
        {
            accessorKey: 'calculation',
            header: 'Hitungan',
        },
        {
            accessorKey: 'status',
            header: 'Keterangan',
            cell: 'status-cell',
        },
        {
            accessorKey: 'actions',
            header: 'Actions',
            cell: 'actions-cell',
        },
    ];

    const getRowActions = (row: any) => [
        [
            {
                type: "label" as const,
                label: "Actions",
            },
            {
                label: "Edit",
                onSelect: () => handleEdit(row.original),
            },
            {
                label: "Delete",
                onSelect: () => handleDelete(row.original.id),
            },
        ],
    ];

    const resetForm = () => {
        form.value = {
            code: '',
            strategicObjective: '',
            kpi: '',
            unit: '',
            hibHig: 'HIG',
            periodType: 'Quartal',
            selectedPeriod: 'Q1',
            yearStart: currentYear,
            yearEnd: currentYear + 4,
            kpiTargets: {},
            internalAuditSO: '',
            actual: '',
            target: '',
            calculation: '',
            status: '',
        };
    };

    const openModal = () => {
        isEditMode.value = false;
        resetForm();
        isAddModalOpen.value = true;
    };

    const closeModal = () => {
        isAddModalOpen.value = false;
        isEditMode.value = false;
        resetForm();
    };

    const handleEdit = (item: any) => {
        isEditMode.value = true;
        form.value = { ...item };
        isAddModalOpen.value = true;
    };

    const handleDelete = async (id: number | string) => {
        if (!confirm("Are you sure you want to delete this Strategic Plan?")) return;
        loading.value = true;
        errorMsg.value = '';
        try {
            const baseUrl = getAuditServiceBaseUrl();
            await $fetch(`${baseUrl}/strategic-plans/${id}`, {
                method: 'DELETE'
            });
            await fetchStrategicPlans();
        } catch (error: any) {
            console.error('Failed to delete strategic plan:', error);
            errorMsg.value = 'Failed to delete strategic plan.';
        } finally {
            loading.value = false;
        }
    };

    const handleSubmit = async () => {
        form.value.calculation = computedCalculation.value;
        form.value.status = computedStatus.value;
        
        loading.value = true;
        errorMsg.value = '';
        try {
            const baseUrl = getAuditServiceBaseUrl();
            if (isEditMode.value) {
                await $fetch(`${baseUrl}/strategic-plans/${form.value.id}`, {
                    method: 'PUT',
                    body: form.value
                });
            } else {
                await $fetch(`${baseUrl}/strategic-plans`, {
                    method: 'POST',
                    body: form.value
                });
            }
            closeModal();
            await fetchStrategicPlans();
        } catch (error: any) {
            console.error('Failed to save strategic plan:', error);
            errorMsg.value = 'Failed to save strategic plan.';
        } finally {
            loading.value = false;
        }
    };

    watch(() => form.value.periodType, (newType) => {
        if (newType === 'Quartal') {
            form.value.selectedPeriod = 'Q1';
        } else {
            form.value.selectedPeriod = String(form.value.yearStart || currentYear);
        }
    });

    watch([(() => form.value.yearStart), (() => form.value.yearEnd)], () => {
        if (form.value.periodType === 'Yearly') {
            const periods = availablePeriods.value;
            if (!periods.includes(form.value.selectedPeriod || '')) {
                form.value.selectedPeriod = periods[0] || '';
            }
        }
    });

    return {
        columns, strategicObjectives, isAddModalOpen, isEditMode, form,
        unitOptions, yearOptions, availablePeriods, computedCalculation, computedStatus,
        getRowActions, openModal, closeModal, handleEdit, handleDelete, handleSubmit,
        fetchStrategicPlans, loading, errorMsg
    };
});