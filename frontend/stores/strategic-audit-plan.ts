import type { TableColumn } from "@nuxt/ui";
import type { StrategicAuditPlan } from "~/types/audit";

export const useStrategicPlanStore = defineStore('strategic-audit-plan', () => {

    // State
    const isAddModalOpen = ref(false);
    const isEditMode = ref(false);

    // Unit options for dropdown
    const unitOptions = [
        { label: 'Percentage (%)', value: '%' },
        { label: 'Rupiah (Rp)', value: 'Rp' },
        { label: 'Amount', value: 'Amount' },
        { label: 'Score', value: 'Score' },
        { label: 'Hour', value: 'Hour' },
        { label: 'Day', value: 'Day' },
    ];

    // Year range for dropdowns
    const currentYear = new Date().getFullYear();
    const yearOptions = Array.from({ length: 20 }, (_, i) => {
        const year = currentYear - 5 + i;
        return { label: String(year), value: year };
    });

    // Form Data
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
        actual: '',
        target: '',
        calculation: '',
        status: '',
    });

    // Computed: available periods based on periodType
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

    // Computed: Hitungan (calculation based on actual and target)
    const computedCalculation = computed(() => {
        const actual = parseFloat(form.value.actual || '0');
        const target = parseFloat(form.value.target || '0');
        if (target === 0) return '';
        const result = (1 - (actual / target)) * 100;
        return `${result.toFixed(2)}%`;
    });

    // Computed: Keterangan based on HIB/HIG and hitungan
    const computedStatus = computed(() => {
        const actual = parseFloat(form.value.actual || '0');
        const target = parseFloat(form.value.target || '0');
        if (target === 0 || !form.value.actual || !form.value.target) return '';

        const ratio = actual / target;
        if (form.value.hibHig === 'HIG') {
            // High is Good: higher actual is better
            if (ratio >= 1) return 'Good';
            if (ratio >= 0.7) return 'Moderate';
            return 'Poor';
        } else {
            // High is Bad: lower actual is better
            if (ratio <= 1) return 'Good';
            if (ratio <= 1.3) return 'Moderate';
            return 'Poor';
        }
    });

    // Mock Data
    const strategicObjectives = ref<StrategicAuditPlan[]>([
        {
            id: 1,
            code: 'SO-IA01',
            strategicObjective: 'Enhance Operational Efficiency',
            kpi: 'Revenue Operational Cost',
            unit: '%',
            hibHig: 'HIG',
            periodType: 'Quartal',
            selectedPeriod: 'Q1',
            actual: '100',
            target: '300',
            calculation: '33.33%',
            status: 'Poor',
        },
        {
            id: 2,
            code: 'SO-IA02',
            strategicObjective: 'Strengthen Internal Control',
            kpi: 'Customer Satisfaction Index',
            unit: 'Score',
            hibHig: 'HIG',
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
            id: 3,
            code: 'SO-IA03',
            strategicObjective: 'Improve Compliance',
            kpi: 'Audit Response Time',
            unit: 'Hour',
            hibHig: 'HIB',
            periodType: 'Quartal',
            selectedPeriod: 'Q2',
            actual: '24',
            target: '48',
            calculation: '50.00%',
            status: 'Good',
        },
    ]);

    // Table Columns
    const columns: TableColumn<StrategicAuditPlan>[] = [
        {
            accessorKey: 'code',
            header: 'Objective ID',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'strategicObjective',
            header: 'Strategic Objective',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'kpi',
            header: 'KPI',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'unit',
            header: 'Unit',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'hibHig',
            header: 'HIB/HIG',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'periodType',
            header: 'Period',
            cell: (row) => {
                const periodType = row.getValue() as string;
                const period = row.row.original.selectedPeriod;
                return `${periodType} - ${period}`;
            },
        },
        {
            accessorKey: 'actual',
            header: 'Actual',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'target',
            header: 'Target',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'calculation',
            header: 'Calculation',
            cell: (row) => row.getValue(),
        },
        {
            accessorKey: 'status',
            header: 'Status',
            cell: (row) => row.getValue(),
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

    // Methods
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

    const handleDelete = (id: number) => {
        if (confirm("Are you sure you want to delete this Strategic Plan?")) {
            strategicObjectives.value = strategicObjectives.value.filter(
                (item) => item.id !== id
            );
        }
    };

    const handleSubmit = () => {
        // Set computed fields before saving
        form.value.calculation = computedCalculation.value;
        form.value.status = computedStatus.value;

        if (isEditMode.value) {
            const index = strategicObjectives.value.findIndex(
                (item) => item.id === form.value.id,
            );
            if (index !== -1) {
                strategicObjectives.value[index] = { ...form.value } as StrategicAuditPlan;
            }
        } else {
            strategicObjectives.value.push({
                ...form.value,
                id: Date.now(),
            } as StrategicAuditPlan);
        }
        closeModal();
    };

    // When periodType changes, reset selectedPeriod
    watch(() => form.value.periodType, (newType) => {
        if (newType === 'Quartal') {
            form.value.selectedPeriod = 'Q1';
        } else {
            form.value.selectedPeriod = String(form.value.yearStart || currentYear);
        }
    });

    // When yearStart or yearEnd changes for Yearly, reset selectedPeriod if needed
    watch([() => form.value.yearStart, () => form.value.yearEnd], () => {
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
    };
});