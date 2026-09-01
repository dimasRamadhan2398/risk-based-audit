import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import type { TableColumn } from "@nuxt/ui";
import type { StrategicAuditPlan } from "~/types/audit";

export const useStrategicPlanStore = defineStore('strategic-audit-plan', () => {

    const isAddModalOpen = ref(false);
    const isViewModalOpen = ref(false);
    const selectedViewObjective = ref<StrategicAuditPlan | null>(null);
    const isEditMode = ref(false);
    const loading = ref(false);
    const errorMsg = ref('');

    const openViewModal = (item: StrategicAuditPlan) => {
        selectedViewObjective.value = item;
        isViewModalOpen.value = true;
    };

    const closeViewModal = () => {
        isViewModalOpen.value = false;
        selectedViewObjective.value = null;
    };

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
        goalId: '',
        strategicObjective: '',
        kpi: '',
        unit: '',
        hibHig: 'HIG',
        periodType: 'Quartal',
        selectedPeriod: 'Q1',
        yearStart: currentYear,
        yearEnd: currentYear + 4,
        kpiTargets: {},
        kpiActuals: {},
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
            strategicObjective: 'Penyelesaian Rencana Audit / Audit Plan Completion',
            kpi: 'Audit Completion Rate',
            unit: '%',
            hibHig: 'HIG',
            periodType: 'Quartal',
            selectedPeriod: 'Q1-2026',
            yearStart: 2026,
            yearEnd: 2027,
            kpiTargets: { 'Q1-2026': '95', 'Q2-2026': '95', 'Q3-2026': '95', 'Q4-2026': '95', 'Q1-2027': '98', 'Q2-2027': '98', 'Q3-2027': '98', 'Q4-2027': '98' },
            kpiActuals: { 'Q1-2026': '97', 'Q2-2026': '', 'Q3-2026': '', 'Q4-2026': '', 'Q1-2027': '', 'Q2-2027': '', 'Q3-2027': '', 'Q4-2027': '' },
            internalAuditSO: 'Tingkatkan efisiensi pelaksanaan audit tahunan',
            actual: '97',
            target: '95',
            calculation: '102.11%',
            status: 'Good',
        },
        {
            id: '2',
            code: 'SO-IA02',
            strategicObjective: 'Ketepatan Waktu Penyampaian Laporan / Report Timeliness',
            kpi: 'Report Timeliness',
            unit: '%',
            hibHig: 'HIG',
            yearStart: 2026,
            yearEnd: 2027,
            kpiTargets: { 'Q1-2026': '95', 'Q2-2026': '95', 'Q3-2026': '95', 'Q4-2026': '95', 'Q1-2027': '98', 'Q2-2027': '98', 'Q3-2027': '98', 'Q4-2027': '98' },
            kpiActuals: { 'Q1-2026': '98', 'Q2-2026': '', 'Q3-2026': '', 'Q4-2026': '', 'Q1-2027': '', 'Q2-2027': '', 'Q3-2027': '', 'Q4-2027': '' },
            internalAuditSO: 'Percepat penyampaian LHA ke pihak manajemen',
            periodType: 'Quartal',
            selectedPeriod: 'Q1-2026',
            actual: '98',
            target: '95',
            calculation: '103.16%',
            status: 'Good',
        },
        {
            id: '3',
            code: 'SO-IA03',
            strategicObjective: 'Kepuasan Auditee & Klien / Client Satisfaction',
            kpi: 'Client Satisfaction',
            unit: 'Score',
            hibHig: 'HIG',
            yearStart: 2024,
            yearEnd: 2028,
            kpiTargets: { '2024': '4.0', '2025': '4.2', '2026': '4.5', '2027': '4.8', '2028': '5.0' },
            kpiActuals: { '2024': '4.2', '2025': '4.5', '2026': '4.7', '2027': '', '2028': '' },
            internalAuditSO: 'Tingkatkan kualitas layanan & rekomendasi audit',
            periodType: 'Yearly',
            selectedPeriod: '2026',
            actual: '4.7',
            target: '4.5',
            calculation: '104.44%',
            status: 'Good',
        },
        {
            id: '4',
            code: 'SO-IA04',
            strategicObjective: 'Penyelesaian Tindak Lanjut / Action Plan Closed',
            kpi: 'Action Plan Closed',
            unit: '%',
            hibHig: 'HIG',
            yearStart: 2026,
            yearEnd: 2027,
            kpiTargets: { 'Q1-2026': '90', 'Q2-2026': '90', 'Q3-2026': '90', 'Q4-2026': '90', 'Q1-2027': '95', 'Q2-2027': '95', 'Q3-2027': '95', 'Q4-2027': '95' },
            kpiActuals: { 'Q1-2026': '87', 'Q2-2026': '', 'Q3-2026': '', 'Q4-2026': '', 'Q1-2027': '', 'Q2-2027': '', 'Q3-2027': '', 'Q4-2027': '' },
            internalAuditSO: 'Memastikan seluruh temuan audit ditindaklanjuti auditee',
            periodType: 'Quartal',
            selectedPeriod: 'Q1-2026',
            actual: '87',
            target: '90',
            calculation: '96.67%',
            status: 'Moderate',
        },
        {
            id: '5',
            code: 'SO-IA05',
            strategicObjective: 'Tingkat Penyelesaian Bulanan / Monthly Completion Rate',
            kpi: 'Monthly Completion Rate',
            unit: '%',
            hibHig: 'HIG',
            yearStart: 2026,
            yearEnd: 2027,
            kpiTargets: { 'Q1-2026': '90', 'Q2-2026': '90', 'Q3-2026': '90', 'Q4-2026': '90', 'Q1-2027': '95', 'Q2-2027': '95', 'Q3-2027': '95', 'Q4-2027': '95' },
            kpiActuals: { 'Q1-2026': '95', 'Q2-2026': '', 'Q3-2026': '', 'Q4-2026': '', 'Q1-2027': '', 'Q2-2027': '', 'Q3-2027': '', 'Q4-2027': '' },
            internalAuditSO: 'Monitoring progres penyelesaian audit per bulan',
            periodType: 'Quartal',
            selectedPeriod: 'Q1-2026',
            actual: '95',
            target: '90',
            calculation: '105.56%',
            status: 'Good',
        },
    ];

    const strategicObjectives = ref<StrategicAuditPlan[]>([...mockObjectives]);

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
                strategicObjectives.value = [...mockObjectives];
            }
        } catch (error: any) {
            console.error('Failed to fetch strategic plans:', error);
            errorMsg.value = 'Failed to load strategic plans.';
            strategicObjectives.value = [...mockObjectives];
        } finally {
            loading.value = false;
        }
    }

    // Load on init
    fetchStrategicPlans();

    const columns: TableColumn<StrategicAuditPlan>[] = [
        { accessorKey: 'code', header: 'Objective ID' },
        { accessorKey: 'strategicObjective', header: 'Strategic Objective' },
        { accessorKey: 'kpi', header: 'KPI Name' },
        { accessorKey: 'unit', header: 'Unit' },
        { accessorKey: 'selectedPeriod', header: 'Period' },
        { accessorKey: 'target', header: 'Target' },
        { accessorKey: 'actual', header: 'Actual' },
        { accessorKey: 'calculation', header: 'Hitungan' },
        { accessorKey: 'status', header: 'Keterangan' },
        { accessorKey: 'actions', header: 'Actions' },
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
        const startY = currentYear;
        const initialTargets: Record<string, string> = {};
        const initialActuals: Record<string, string> = {};
        for (let i = 0; i < 5; i++) {
            initialTargets[startY + i] = '';
            initialActuals[startY + i] = '';
        }
        form.value = {
            code: '',
            goalId: '',
            strategicObjective: '',
            kpi: '',
            unit: '%',
            hibHig: 'HIG',
            periodType: 'Yearly',
            selectedPeriod: String(startY),
            yearStart: startY,
            yearEnd: startY + 4,
            kpiTargets: initialTargets,
            kpiActuals: initialActuals,
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
        const startY = item.yearStart || currentYear;
        const endY = item.yearEnd || (startY + 4);
        const targets: Record<string, string> = { ...(item.kpiTargets || {}) };
        const actuals: Record<string, string> = { ...(item.kpiActuals || {}) };

        form.value = {
            ...item,
            yearStart: startY,
            yearEnd: endY,
            kpiTargets: targets,
            kpiActuals: actuals,
        };
        isAddModalOpen.value = true;
    };

    const handleDelete = async (id: number | string) => {
        if (!await useGlobalModalStore().confirmDelete({ description: "Are you sure you want to delete this Strategic Plan?" })) return;
        loading.value = true;
        errorMsg.value = '';
        try {
            const baseUrl = getAuditServiceBaseUrl();
            await $fetch(`${baseUrl}/strategic-plans/${id}`, {
                method: 'DELETE'
            });
            strategicObjectives.value = strategicObjectives.value.filter(o => o.id !== id);
            await fetchStrategicPlans();
        } catch (error: any) {
            console.error('Failed to delete strategic plan:', error);
            strategicObjectives.value = strategicObjectives.value.filter(o => o.id !== id);
        } finally {
            loading.value = false;
        }
    };

    const cleanKpiMaps = () => {
        if (!form.value.kpiTargets) form.value.kpiTargets = {};
        if (!form.value.kpiActuals) form.value.kpiActuals = {};

        const cleanedTargets: Record<string, string> = {};
        const cleanedActuals: Record<string, string> = {};

        if (form.value.periodType === 'Quartal') {
            for (const [key, val] of Object.entries(form.value.kpiTargets)) {
                if (!/^\d{4}$/.test(key) && val !== undefined && val !== null) {
                    cleanedTargets[key] = String(val);
                }
            }
            for (const [key, val] of Object.entries(form.value.kpiActuals)) {
                if (!/^\d{4}$/.test(key) && val !== undefined && val !== null) {
                    cleanedActuals[key] = String(val);
                }
            }
        } else {
            for (const [key, val] of Object.entries(form.value.kpiTargets)) {
                if (/^\d{4}$/.test(key) && val !== undefined && val !== null) {
                    cleanedTargets[key] = String(val);
                }
            }
            for (const [key, val] of Object.entries(form.value.kpiActuals)) {
                if (/^\d{4}$/.test(key) && val !== undefined && val !== null) {
                    cleanedActuals[key] = String(val);
                }
            }
        }

        form.value.kpiTargets = cleanedTargets;
        form.value.kpiActuals = cleanedActuals;
    };

    const handleSubmit = async () => {
        cleanKpiMaps();

        if (!form.value.code) {
            form.value.code = `SO-IA${String(strategicObjectives.value.length + 1).padStart(2, '0')}`;
        }
        
        const startY = form.value.yearStart || currentYear;
        const endY = form.value.yearEnd || (startY + 4);
        const targets = form.value.kpiTargets as Record<string, string>;
        const actuals = form.value.kpiActuals as Record<string, string>;

        const selPeriod = form.value.selectedPeriod || (form.value.periodType === 'Quartal' ? `Q1-${startY}` : String(startY));
        let currentTarget = targets[selPeriod] || targets['Q1-' + startY] || targets[String(startY)] || form.value.target;
        let currentActual = actuals[selPeriod] || actuals['Q1-' + startY] || actuals[String(startY)] || form.value.actual;

        if (form.value.periodType === 'Quartal') {
            if (!currentTarget) {
                for (let y = startY; y <= endY; y++) {
                    for (const q of ['Q1', 'Q2', 'Q3', 'Q4']) {
                        const k = `${q}-${y}`;
                        if (targets[k]) { currentTarget = targets[k]; break; }
                    }
                    if (currentTarget) break;
                }
            }
            if (!currentActual) {
                for (let y = startY; y <= endY; y++) {
                    for (const q of ['Q1', 'Q2', 'Q3', 'Q4']) {
                        const k = `${q}-${y}`;
                        if (actuals[k]) { currentActual = actuals[k]; break; }
                    }
                    if (currentActual) break;
                }
            }
        }
        
        if (currentTarget !== undefined && currentTarget !== '') {
            form.value.target = String(currentTarget);
        }
        if (currentActual !== undefined && currentActual !== '') {
            form.value.actual = String(currentActual);
        }
        form.value.calculation = computedCalculation.value;
        form.value.status = computedStatus.value;
        
        loading.value = true;
        errorMsg.value = '';
        const appToast = useAppToast();
        const editModeState = isEditMode.value;

        try {
            const baseUrl = getAuditServiceBaseUrl();
            if (editModeState) {
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
            
            const idx = strategicObjectives.value.findIndex(o => String(o.id) === String(form.value.id));
            if (idx !== -1) {
                strategicObjectives.value[idx] = { ...form.value } as StrategicAuditPlan;
            } else {
                strategicObjectives.value.push({ ...form.value, id: String(Date.now()) } as StrategicAuditPlan);
            }

            appToast.success(
                editModeState ? 'Rencana Strategis Diperbarui' : 'Rencana Strategis Dibuat',
                `Rencana Strategis "${form.value.strategicObjective || form.value.code}" berhasil disimpan.`
            );

            closeModal();
            await fetchStrategicPlans();
        } catch (error: any) {
            console.error('Failed to save strategic plan:', error);
            const detail = error.data?.error?.message || error.message || 'Gagal menyimpan rencana strategis.';
            appToast.error('Gagal Menyimpan Rencana Strategis', detail);

            const idx = strategicObjectives.value.findIndex(o => String(o.id) === String(form.value.id));
            if (idx !== -1) {
                strategicObjectives.value[idx] = { ...form.value } as StrategicAuditPlan;
            } else {
                strategicObjectives.value.push({ ...form.value, id: String(Date.now()) } as StrategicAuditPlan);
            }
            closeModal();
        } finally {
            loading.value = false;
        }
    };

    watch(() => form.value.periodType, (newType) => {
        cleanKpiMaps();
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
        isViewModalOpen, selectedViewObjective, openViewModal, closeViewModal,
        unitOptions, yearOptions, availablePeriods, computedCalculation, computedStatus,
        getRowActions, openModal, closeModal, handleEdit, handleDelete, handleSubmit,
        fetchStrategicPlans, loading, errorMsg
    };
});