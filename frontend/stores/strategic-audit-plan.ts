import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import type { TableColumn } from "@nuxt/ui";
import type { StrategicAuditPlan } from "~/types/audit";
import { useToastNotification } from '~/components/shared/ToastNotification.vue';
import { extractErrorMessage } from '~/utils/error';
import { useI18n } from '~/composables/useI18n';
import { getFiscalYears } from '~/composables/useFiscalYear';
import { buildStrategicPlanPayload, strategicPlanErrorField, strategicPlanFormFromPlan } from '~/utils/strategicPlanPayload';

export const useStrategicPlanStore = defineStore('strategic-audit-plan', () => {

    const isAddModalOpen = ref(false);
    const isViewModalOpen = ref(false);
    const selectedViewObjective = ref<StrategicAuditPlan | null>(null);
    const isEditMode = ref(false);
    const loading = ref(false);
    // A save in flight (separate from `loading`, which is also true while the list loads).
    const saving = ref(false);
    const errorMsg = ref('');
    // Error of the last failed save, shown inside the form; `formFieldErrors` ties it to a field when the backend says which.
    const formError = ref('');
    const formFieldErrors = ref<Partial<Record<'category', string>>>({});
    const toast = useToastNotification();
    const { t } = useI18n();

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
    const yearOptions = getFiscalYears().map(y => ({ label: String(y), value: y }));

    const form = ref<Partial<StrategicAuditPlan>>({
        code: '',
        goalId: '',
        strategicObjective: '',
        kpi: '',
        unit: '',
        category: '',
        hibHig: 'HIG',
        periodType: 'Quartal',
        selectedPeriod: 'Q1',
        yearStart: currentYear,
        yearEnd: currentYear + 3,
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

    // Only what the API returned: no sample objectives before the first load, on an empty list or on an error.
    const strategicObjectives = ref<StrategicAuditPlan[]>([]);

    const fetchStrategicPlans = async () => {
        loading.value = true;
        errorMsg.value = '';
        try {
            const baseUrl = getAuditServiceBaseUrl();
            const response: any = await $fetch(`${baseUrl}/strategic-plans`, {
                method: 'GET',
                params: {
                    page: 1,
                    page_size: 100,
                    order: 'code ASC'
                }
            });
            let items: StrategicAuditPlan[] = [];
            if (response && response.data && Array.isArray(response.data.items)) {
                items = response.data.items;
            } else if (response && Array.isArray(response.items)) {
                items = response.items;
            } else if (Array.isArray(response)) {
                items = response;
            }

            strategicObjectives.value = items;
        } catch (error: any) {
            console.error('Failed to fetch strategic plans:', error);
            errorMsg.value = extractErrorMessage(error, 'Failed to load strategic plans.');
            strategicObjectives.value = [];
        } finally {
            loading.value = false;
        }
    }

    /** One plan by id, e.g. to edit a row that is not in the (first 100) loaded objectives. Null on failure. */
    const fetchStrategicPlanById = async (id: string | number): Promise<StrategicAuditPlan | null> => {
        try {
            const baseUrl = getAuditServiceBaseUrl();
            const response: any = await $fetch(`${baseUrl}/strategic-plans/${id}`);
            const plan = response?.data ?? response;
            return plan && plan.id ? plan as StrategicAuditPlan : null;
        } catch (error: any) {
            console.error('Failed to fetch strategic plan:', error);
            toast.showError(t('strategicPlan.toast.loadOneFailed'), extractErrorMessage(error, t('strategicPlan.toast.loadOneFailed')));
            return null;
        }
    };

    // Load on init
    fetchStrategicPlans();

    const columns: (TableColumn<StrategicAuditPlan> & { class?: string })[] = [
        { accessorKey: 'strategicObjective', header: 'Strategic Objective', class: 'max-w-[280px] whitespace-normal break-words font-medium' },
        { accessorKey: 'kpi', header: 'KPI Name', class: 'max-w-[240px] whitespace-normal break-words font-medium' },
        { accessorKey: 'unit', header: 'Unit', class: 'w-20 text-center whitespace-nowrap' },
        { accessorKey: 'selectedPeriod', header: 'Period', class: 'w-28 whitespace-nowrap' },
        { accessorKey: 'target', header: 'Target', class: 'w-24 whitespace-nowrap' },
        { accessorKey: 'actual', header: 'Actual', class: 'w-24 whitespace-nowrap' },
        { accessorKey: 'calculation', header: 'Hitungan', class: 'w-28 whitespace-nowrap' },
        { accessorKey: 'status', header: 'Keterangan', class: 'w-28 whitespace-nowrap' },
        { accessorKey: 'actions', header: 'Actions', class: 'w-24 whitespace-nowrap text-center' },
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
            category: '',
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
        clearFormErrors();
    };

    const clearFormErrors = () => {
        formError.value = '';
        formFieldErrors.value = {};
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

        // Only the fields the form edits (plus the id): the rest of the row is never sent back on save.
        form.value = {
            ...strategicPlanFormFromPlan(item),
            yearStart: startY,
            yearEnd: endY,
            kpiTargets: targets,
            kpiActuals: actuals,
        };
        clearFormErrors();
        isAddModalOpen.value = true;
    };

    const handleDelete = async (id: number | string) => {
        if (!await useGlobalModalStore().confirmDelete({ description: t('strategicPlan.toast.confirmDelete') })) return;
        loading.value = true;
        errorMsg.value = '';
        try {
            const baseUrl = getAuditServiceBaseUrl();
            await $fetch(`${baseUrl}/strategic-plans/${id}`, {
                method: 'DELETE'
            });
            toast.showSuccess(t('strategicPlan.toast.deletedTitle'), t('strategicPlan.toast.deletedDesc'));
            strategicObjectives.value = strategicObjectives.value.filter(o => o.id !== id);
            await fetchStrategicPlans();
        } catch (error: any) {
            // The row stays: it was not deleted on the server.
            console.error('Failed to delete strategic plan:', error);
            const detail = extractErrorMessage(error, t('strategicPlan.toast.deleteFailedFallback'));
            errorMsg.value = detail;
            toast.showError(t('strategicPlan.toast.deleteFailedTitle'), detail);
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
        if (saving.value) return;
        if (!form.value.strategicObjective) {
            toast.showWarning(t('strategicPlan.toast.validationTitle'), t('strategicPlan.toast.objectiveRequired'));
            return;
        }

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

        saving.value = true;
        loading.value = true;
        errorMsg.value = '';
        clearFormErrors();
        const editModeState = isEditMode.value;
        const savedName = form.value.strategicObjective || form.value.code || '';

        try {
            const baseUrl = getAuditServiceBaseUrl();
            const body = buildStrategicPlanPayload(form.value);
            if (editModeState) {
                await $fetch(`${baseUrl}/strategic-plans/${form.value.id}`, {
                    method: 'PUT',
                    body
                });
            } else {
                await $fetch(`${baseUrl}/strategic-plans`, {
                    method: 'POST',
                    body
                });
            }
        } catch (error: any) {
            // Nothing was saved: the list stays as it is and the modal stays open with the input, so the user can fix and retry.
            console.error('Failed to save strategic plan:', error);
            const detail = extractErrorMessage(error, t('strategicPlan.toast.saveFailedFallback'));
            formError.value = detail;
            const field = strategicPlanErrorField(error, detail);
            if (field) formFieldErrors.value = { [field]: detail };
            toast.showError(t('strategicPlan.toast.saveFailedTitle'), detail);
            saving.value = false;
            loading.value = false;
            return;
        }

        // Saved: show the list as the server has it now (with the server's id) rather than the local form.
        toast.showSuccess(
            editModeState ? t('strategicPlan.toast.updatedTitle') : t('strategicPlan.toast.createdTitle'),
            t('strategicPlan.toast.savedDesc', { name: savedName })
        );
        saving.value = false;
        closeModal();
        await fetchStrategicPlans();
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
        fetchStrategicPlans, fetchStrategicPlanById, loading, saving, errorMsg,
        formError, formFieldErrors, clearFormErrors
    };
});