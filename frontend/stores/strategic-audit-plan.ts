import type { TableColumn } from "@nuxt/ui";
import type { StrategicAuditPlan, KPITargetYear } from "~/types/audit";
import { h } from 'vue'; // Import the 'h' function

export const useStrategicPlanStore = defineStore('strategic-audit-plan', () => {

    // State
    const isAddModalOpen = ref(false);
    const isEditMode = ref(false);

    // Form Data
    const form = ref<Partial<StrategicAuditPlan>>({
        number: "",
        objectives: "",
        kpi: "",
        characteristicData: "",
        kpiTarget: [],
        unit: "",
    });

    // Sifat Data Options
    const characteristicDataOptions = [
        { label: "HIG", value: "hig" },
        { label: "HIB", value: "hib" },
    ];

    // Mock Data
    const strategicObjectives = ref<StrategicAuditPlan[]>([
        {
            id: 1,
            number: "1",
            objectives: "Enhance operational efficiency across all departments",
            kpi: "Cost reduction ratio",
            characteristicData: "HIG",
            kpiTarget: [
                { year: 2025, value: "5" },
                { year: 2026, value: "10" },
                { year: 2027, value: "15" }
            ],
            unit: "%",
        },
        {
            id: 2,
            number: "2",
            objectives: "Strengthen internal control systems",
            kpi: "Control effectiveness score",
            characteristicData: "HIG",
            kpiTarget: [
                { year: 2025, value: "85" },
                { year: 2026, value: "90" }
            ],
            unit: "%",
        },
        {
            id: 3,
            number: "3",
            objectives: "Improve risk management practices",
            kpi: "Risk response time",
            characteristicData: "HIB",
            kpiTarget: [
                { year: 2025, value: "48" }
            ],
            unit: "Hours",
        },
        {
            id: 4,
            number: "4",
            objectives: "Enhance compliance with regulations",
            kpi: "Compliance audit score",
            characteristicData: "HIG",
            kpiTarget: [
                { year: 2025, value: "100" }
            ],
            unit: "Score",
        },
    ]);

    // Table Columns
    const columns: TableColumn<StrategicAuditPlan>[] = [
        {
            accessorKey: "number",
            header: "No.",
            cell: (row) => row.getValue(),
            // meta: {
            //     class: {
            //         th: "bg-primary border-rounded text-secondary-900 text-center w-16",
            //         td: "text-center font-semibold text-gray-900",
            //     },
            // },
        },
        {
            accessorKey: "objectives",
            header: "Corporate Strategic Objectives",
            cell: (row) => row.getValue(),

        },
        {
            accessorKey: "kpi",
            header: "KPI",
            cell: (row) => row.getValue(),

        },
        {
            accessorKey: "characteristicData",
            header: "Characteristic Data",
            cell: (row) => row.getValue(),

        },
        {
            accessorKey: "kpiTarget",
            header: "KPI Target",
            cell: (row) => {
                const targets = row.getValue() as KPITargetYear[];
                const unit = row.row.original.unit;
                return h('div', { class: 'flex flex-col gap-1' },
                    targets.map(t => h('div', { class: 'text-xs' }, `${t.year}: ${t.value} ${unit}`))
                );
            },

        },
        {
            accessorKey: "unit",
            header: "Unit",
            cell: (row) => row.getValue(),

        },
        {
            accessorKey: "actions",
            header: "Actions",
            cell: "actions-cell",

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
    const openModal = () => {
        isEditMode.value = false;

        // Reset Form
        form.value = {
            number: '',
            objectives: '',
            kpi: '',
            characteristicData: '',
            kpiTarget: [],
            unit: '',
        };
        isAddModalOpen.value = true;
    }

    const closeModal = () => {
        isAddModalOpen.value = false;
        isEditMode.value = false;
        form.value = {
            number: "",
            objectives: "",
            kpi: "",
            characteristicData: "",
            kpiTarget: [],
            unit: "",
        };
    };

    const handleEdit = (item: any) => {
        isEditMode.value = true;
        form.value = { ...item };
        isAddModalOpen.value = true;
    };

    const handleDelete = (id: number) => {
        if (confirm("Are you sure you want to delete this strategic objective?")) {
            strategicObjectives.value = strategicObjectives.value.filter(
                (item) => item.id !== id
            );
        }
    };

    const addKpiTargetYear = () => {
        if (!form.value.kpiTarget) {
            form.value.kpiTarget = [];
        }
        const currentYear = new Date().getFullYear();
        const nextYear = form.value.kpiTarget.length > 0
            ? Math.max(...form.value.kpiTarget.map(t => t.year)) + 1
            : currentYear;

        form.value.kpiTarget.push({ year: nextYear, value: "" });
    };

    const removeKpiTargetYear = (index: number) => {
        form.value.kpiTarget?.splice(index, 1);
    };

    const handleSubmit = () => {
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
                status: "In Progress",
            } as StrategicAuditPlan);
        }
        closeModal();
    };

    return {
        columns, strategicObjectives, isAddModalOpen, isEditMode, form, characteristicDataOptions,
        getRowActions, openModal, closeModal, handleEdit, handleDelete, handleSubmit,
        addKpiTargetYear, removeKpiTargetYear
    }
});