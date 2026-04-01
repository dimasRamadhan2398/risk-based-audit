<template>
  <UCard variant="soft">
    <template #header>
      <div class="flex flex-row items-stretch justify-between gap-8">
        <div class="flex flex-col gap-2">
          <div class="flex flex-row items-center">
            <h1 class="text-3xl font-bold text-gray-900">Risk Profile</h1>
            <UButton
              icon="info"
              size="md"
              color="primary"
              variant="ghost"
              class="hover:cursor-pointer"
              @click="isInformationRiskProfileModalOpen = true"
            />
          </div>
          <h6>
            Daftar risiko prioritas perusahaan dengan tingkat dampak dan
            kemungkinan
          </h6>
        </div>
        <div class="self-start">
          <UButton
            icon="add"
            label="Tambah Risiko"
            variant="solid"
            color="primary"
            size="sm"
            @click="onAddRisk"
          >
          </UButton>
        </div>
      </div>
    </template>
    <div class="relative flex flex-col gap-4">
      <h5>Matriks Resiko</h5>
      <UTable
        :data="data"
        :columns="columns"
      />
    </div>
    <template #footer>
      <div class="relative flex flex-col gap-4">
        <div class="flex flex-row items-stretch justify-between gap-8">
          <h5>List Resiko</h5>
          <UButton
            label="Import"
            variant="solid"
            color="primary"
            size="sm"
            icon="import"
            @click="isAddRiskProfileFormOpen = true"
          >
          </UButton>
        </div>
        <UTable
          ref="table"
          :data="riskData"
          :columns="priorityRiskListColumn"
          :pagination-options="{
            getPaginationRowModel: getPaginationRowModel(),
          }"
          v-model:pagination-state="pagination"
        />
        <div class="flex flex-row items-end justify-end">
          <UPagination
            active-variant="solid"
            :page="(table?.tableApi?.getState().pagination.pageIndex || 0) + 1"
            :items-per-page="table?.tableApi?.getState().pagination.pageSize"
            :total="riskData.length"
            @update:page="(p) => table?.tableApi?.setPageIndex(p - 1)"
          />
        </div>
      </div>
    </template>
  </UCard>
  <AddRiskProfileForm
    v-model:isOpen="isAddRiskProfileFormOpen"
    v-model:editMode="onEditMode"
    v-model:riskData="selectedRiskValueRaw"
  />
  <DeleteConfirmationRiskItem
    v-model:isOpen="isDeleteRiskProfileModalOpen"
    v-model:riskName="selectedRiskName"
    v-model:riskId="selectedRiskId"
  />
  <DetailRiskProfileForm
    v-model:isOpen="isDetailRiskProfileModalOpen"
    v-model:riskData="selectedRiskValueRaw"
    v-model:quarterlyActive="isQuarterlyActive"
  />
  <InformationRiskProfile v-model:isOpen="isInformationRiskProfileModalOpen" />
</template>

<script setup lang="ts">
import { getPaginationRowModel } from '@tanstack/vue-table'
import type { TableColumn } from "@nuxt/ui";
import AddRiskProfileForm from "~/components/risk-profile/AddRiskProfileForm.vue";
import DeleteConfirmationRiskItem from "~/components/risk-profile/DeleteConfirmationRiskItem.vue";
import { useRiskProfileStore, type RiskListItem } from "~/stores/profile-risk";
import DetailRiskProfileForm from "~/components/risk-profile/DetailRiskProfileForm.vue";
import { riskColor } from "~/types/common";
import InformationRiskProfile from "~/components/risk-profile/InformationRiskProfile.vue";
import { useRoute } from 'vue-router'

const route = useRoute()

onMounted(() => {
  // Cek apakah ada instruksi untuk membuka detail dari halaman mitigasi
  const riskIdToOpen = route.query.openDetail as string
  
  if (riskIdToOpen) {
    const risk = riskProfileStore.getRiskById(riskIdToOpen)
    if (risk) {
      // Set data risiko yang dipilih ke state store/local
      riskProfileStore.setSelectedRiskValue(risk)
      // Buka modal secara otomatis
      isDetailRiskProfileModalOpen.value = true
    }
  }
})

const props = defineProps<{
  risks: [];
}>();

const table = useTemplateRef('table')

const UBadge = resolveComponent("UBadge");
const UDropdownMenu = resolveComponent("UDropdownMenu");
const UButton = resolveComponent("UButton");

const { copy } = useClipboard();
const toast = useToast();

const riskProfileStore = useRiskProfileStore();

const isAddRiskProfileFormOpen = ref(false);
const isDeleteRiskProfileModalOpen = ref(false);
const isDetailRiskProfileModalOpen = ref(false);
const isInformationRiskProfileModalOpen = ref(false);

const onEditMode = ref(false);

// Use store data instead of local refs
const data = computed(() => riskProfileStore.getRiskMatrix);
const riskData = computed(() => riskProfileStore.getRiskList);

const riskState = ref(riskProfileStore.riskState)

const selectedRiskValueRaw = computed(() => {
  const raw = toRaw(selectedRiskValue.value);
  return raw === null ? undefined : raw;
});

// Computed with getters/setters that sync with store
const selectedRiskId = computed({
  get: () => riskState.value.selectedRiskId,
  set: (value: string) => riskProfileStore.setSelectedRiskId(value),
});

const selectedRiskName = computed({
  get: () => riskState.value.selectedRiskName,
  set: (value: string) => riskProfileStore.setSelectedRiskName(value),
});

const selectedRiskValue = computed({
  get: () => riskState.value.selectedRiskValue,
  set: (value: RiskListItem | null) => riskProfileStore.setSelectedRiskValue(value),
});

const isQuarterlyActive = computed(() => {
  if (!selectedRiskValueRaw.value) {
    return false;
  }
  const length = selectedRiskValueRaw.value.list_residual_risks.length;
  if(length < 4){
    return false;
  }
  return true;
})

const pagination = ref({
  pageIndex: 0,
  pageSize: 0,
});

watch(riskData, () => {
  pagination.value = {
    pageIndex: 0,
    pageSize: 10,
  };
});

const onAddRisk = () => {
  isAddRiskProfileFormOpen.value = true
  onEditMode.value = false
}

const cleanData = () => {
  selectedRiskId.value = "";
  selectedRiskName.value = "";
  selectedRiskValue.value = null;
}

const getRiskCellValue = (risk: any) => {
  const selectedColor = Object.entries(riskColor).find(([key]) => {
    const clearedKey = risk.getValue().replace(/\s+\d+$/, "");
    return clearedKey === key;
  })?.[1];

  return h(
    UBadge,
    { class: "capitalize", variant: "soft", color: selectedColor },
    () => risk.getValue(),
  );
};

const deleteRisk = (id: string) => {
  selectedRiskId.value = id;
  isDeleteRiskProfileModalOpen.value = true;
};
const columns: TableColumn<Object>[] = [
  {
    accessorKey: "name",
    header: "Impact/Likelihood",
    cell: (risk: any) => {
      return risk.getValue();
    },
    meta: {
      class: {
        th: "text-center font-semibold bg-primary text-secondary-900",
        td: "text-center font-semibold bg-primary text-secondary-900 bg-primary",
      },
    },
  },
  {
    accessorKey: "riskId1",
    header: "1 Sangat Rendah",
    cell: getRiskCellValue,
    meta: {
      class: {
        th: "text-center font-semibold bg-primary text-secondary-900",
        td: "text-center font-semibold [&:nth-child(2):has(td:nth-child(2))]:bg-primary",
      },
    },
  },
  {
    accessorKey: "riskId2",
    header: "2 Rendah",
    cell: getRiskCellValue,
    meta: {
      class: {
        th: "text-center font-semibold bg-primary text-secondary-900",
        td: "text-center font-semibold [&:nth-child(2):has(td:nth-child(2))]:bg-primary",
      },
    },
  },
  {
    accessorKey: "riskId3",
    header: "3 Moderat",
    cell: getRiskCellValue,
    meta: {
      class: {
        th: "text-center font-semibold bg-primary text-secondary-900",
        td: "text-center font-semibold [&:nth-child(2):has(td:nth-child(2))]:bg-primary",
      },
    },
  },
  {
    accessorKey: "riskId4",
    header: "4 Tinggi",
    cell: getRiskCellValue,
    meta: {
      class: {
        th: "text-center font-semibold bg-primary text-secondary-900",
        td: "text-center font-semibold [&:nth-child(2):has(td:nth-child(2))]:bg-primary",
      },
    },
  },
  {
    accessorKey: "riskId5",
    header: "5 Sangat Tinggi",
    cell: getRiskCellValue,
    meta: {
      class: {
        th: "text-center font-semibold bg-primary text-secondary-900",
        td: "text-center font-semibold [&:nth-child(2):has(td:nth-child(2))]:bg-primary",
      },
    },
  },
];

const priorityRiskListColumn: TableColumn<RiskListItem>[] = [
  {
    accessorKey: "risk_id",
    header: "Risk ID",
    cell: (risk: any) => {
      return risk.getValue();
    },
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center text-black",
      },
    },
  },
  {
    accessorKey: "risk_name",
    header: "Risk Name",
    cell: (risk: any) => {
      return risk.getValue();
    },
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center text-black",
      },
    },
  },
  {
    accessorKey: "risk_category",
    header: "Risk Category",
    cell: (risk: any) => {
      return risk.getValue();
    },
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center text-black",
      },
    },
  },
  {
    accessorKey: "latest_impact_level",
    header: "Impact Level",
    cell: (risk: any) => {
      return risk.getValue();
    },
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center text-black",
      },
    },
  },
  {
    accessorKey: "latest_possibility_level",
    header: "Possibility Level",
    cell: (risk: any) => {
      return risk.getValue();
    },
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center text-black",
      },
    },
  },
  {
    accessorKey: "risk_level",
    header: "Risk Level",
    cell: getRiskCellValue,
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "place-self-center text-center",
      },
    },
  },
  {
    accessorKey: "actions",
    header: "Actions",
    cell: ({ row }) => {
      return h(
        UDropdownMenu,
        {
          content: {
            align: "end",
          },
          items: getRowItems(row),
          "aria-label": "Actions dropdown",
        },
        () =>
          h(UButton, {
            icon: "i-lucide-ellipsis-vertical",
            color: "primary",
            variant: "ghost",
            "aria-label": "Actions dropdown",
          }),
      );
    },
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 mx-auto text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center",
      },
    },
  },
];

function getRowItems(row: any) {
  return [
    {
      type: "label",
      label: "Actions",
    },
    {
      label: "Copy Risk ID",
      onSelect() {
        copy(row.original.risk_id);

        toast.add({
          title: "Risk ID copied to clipboard!",
          color: "success",
          icon: "i-lucide-circle-check",
        });
      },
    },
    {
      type: "separator",
    },
    {
      label: "Details",
      onSelect() {
        selectedRiskValue.value = row.original;
        isDetailRiskProfileModalOpen.value = true;
        onEditMode.value = false;
      },
    },
    {
      label: "Edit",
      onSelect() {
        selectedRiskValue.value = row.original;
        isAddRiskProfileFormOpen.value = true;
        onEditMode.value = true;
      },
    },
    {
      label: "Delete",
      onSelect() {
        selectedRiskName.value = row.original.risk_name;
        deleteRisk(row.original.risk_id);
      },
    },
  ];
}
</script>
