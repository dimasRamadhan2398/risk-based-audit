<template>
  <UModal
    title="Detail Risk Data"
    description=""
    :open="props.isOpen"
    :close="{
      color: 'primary',
      variant: 'outline',
      class: 'rounded-full',
      onClick: closeModal,
    }"
    :ui="{
      content: 'min-w-[95vw]',
    }"
    variant="default"
    scrollable
    dismissible
  >
    <template #header>
      <div class="flex justify-between flex-row w-full">
        <h5>Detail Risk Data</h5>
        <div class="gap-4 flex flex-row">
          <UButton label="Edit" icon="edit" color="secondary" variant="outline"></UButton>
          <UButton
            color="primary"
            variant="outline"
            icon="close"
            v-on:click="closeModal"
          />
        </div>
      </div>
    </template>
    <template #body>
      <div class="overflow-x-auto">
        <UTable :data="tableData" :columns="tableColumns"> </UTable>
        <div
          v-if="!quarterlyActive"
          class="min-h-max flex items-center text-center flex-row justify-center"
        >
          <UButton
            variant="solid"
            icon="add"
            :ui="{
              leadingIcon: `font-bold`,
            }"
            >Add Risk Control</UButton
          >
        </div>
      </div>
    </template>
    <template #footer>
      <div></div>
      <UButton @click="handleSubmit" color="primary">Close</UButton>
    </template>
  </UModal>
</template>
<script setup lang="ts">
import { toRaw } from "vue";
import type { TableColumn, TabsItem } from "@nuxt/ui";
import { riskColor, type Quarter, type ResidualFields } from "~/types/common";

const UBadge = resolveComponent("UBadge");

const props = defineProps<{
  isOpen: boolean;
  riskData: any;
  quarterlyActive: boolean;
}>();

interface riskData {
  risk_name?: string;
  risk_category?: string;
  impact_level: number;
  possibility_level: number;
}

const riskData = ref<RiskListItem>(props.riskData);
const quarterlyActive = ref(props.quarterlyActive);

watch(
  () => props.riskData,
  (newValue) => {
    console.log(newValue);
    if (newValue) {
      riskData.value = newValue;
    }
  },
  { immediate: true },
);

watch(
  () => props.quarterlyActive,
  (newValue) => {
    quarterlyActive.value = newValue;
    console.log(newValue);
  },
  { immediate: true },
)

interface RiskDetailTableRow {
  no: string;
  risk_name: string;
  risk_category: string;
  impact_q1: number;
  impact_q2: number;
  impact_q3: number;
  impact_q4: number;
  possibility_q1: number;
  possibility_q2: number;
  possibility_q3: number;
  possibility_q4: number;
  risk_level_q1: string;
  risk_level_q2: string;
  risk_level_q3: string;
  risk_level_q4: string;
}

const emit = defineEmits<{
  "update:isOpen": [value: boolean];
}>();

const formattedRiskData = computed<RiskDetailTableRow>(() => {
  const risk_residual_objects: Record<string, any> = {
    impact_q1: 0,
    impact_q2: 0,
    impact_q3: 0,
    impact_q4: 0,
    possibility_q1: 0,
    possibility_q2: 0,
    possibility_q3: 0,
    possibility_q4: 0,
    risk_level_q1: "",
    risk_level_q2: "",
    risk_level_q3: "",
    risk_level_q4: "",
  };

  if (riskData.value.list_residual_risks) {
    riskData.value.list_residual_risks.forEach(
      (residual: any, index: number) => {
        const quarter = index + 1;
        risk_residual_objects[`impact_q${quarter}`] =
          residual.impact_level || 0;
        risk_residual_objects[`possibility_q${quarter}`] =
          residual.possibility_level || 0;
        risk_residual_objects[`risk_level_q${quarter}`] =
          residual.risk_level || "";
      },
    );
  }

  return {
    no: riskData.value.risk_id,
    risk_name: riskData.value.risk_name,
    risk_category: riskData.value.risk_category,
    ...risk_residual_objects,
  } as RiskDetailTableRow;
});

// Sample table data
const tableData = computed(() => [formattedRiskData.value]);
const riskProfileStore = useRiskProfileStore();
const tabActiveIndex = ref(0);

const tabActive = computed({
  get: () => tabActiveIndex.value,

  set: (value: string | number) => {
    tabActiveIndex.value =
      typeof value === "string" ? parseInt(value, 10) : value;
  },
});

const quarterlyRiskProfiles = ref<riskData[]>([
  {
    impact_level: 1,
    possibility_level: 1,
  },
  {
    impact_level: 1,
    possibility_level: 1,
  },
  {
    impact_level: 1,
    possibility_level: 1,
  },
  {
    impact_level: 1,
    possibility_level: 1,
  },
]);

const confirmRiskProfile = async () => {
  try {
    let payload: riskData[] | CreateRiskProfilePayload =
      quarterlyRiskProfiles.value;
    if (!quarterlyActive.value) {
      payload = {
        risk_name: riskData.value.risk_name!,
        risk_category: riskData.value.risk_category!,
        list_residual_risks: riskData.value.list_residual_risks,
      };
      await riskProfileStore.createRiskProfile(
        payload as CreateRiskProfilePayload,
      );
    }
    closeModal();
  } catch (error) {
    console.error("Error adding risk profile:", error);
    // You can add a toast notification here if needed
  }
};

const closeModal = () => {
  console.log("close");
  emit("update:isOpen", false);
  riskData.value = props.riskData;
};

// const currentImpactLevel = computed({
//   get: () =>
//     quarterlyActive.value
//       ? quarterlyRiskProfiles.value![tabActiveIndex.value]!.impact_level
//       : riskData.value.impact_level,
//   set: (value: number) => {
//     if (quarterlyActive.value) {
//       quarterlyRiskProfiles.value![tabActiveIndex.value]!.impact_level = value;
//     } else {
//       riskData.value.impact_level = value;
//     }
//   },
// });

// const currentPossibilityLevel = computed({
//   get: () =>
//     quarterlyActive.value
//       ? quarterlyRiskProfiles.value![tabActiveIndex.value]!.possibility_level
//       : riskData.value.possibility_level,
//   set: (value: number) => {
//     if (quarterlyActive.value) {
//       quarterlyRiskProfiles.value![tabActiveIndex.value]!.possibility_level =
//         value;
//     } else {
//       riskData.value.possibility_level = value;
//     }
//   },
// });

const handleSubmit = () => {
  closeModal();
};

const tableColumns: TableColumn<RiskDetailTableRow>[] = [
  {
    accessorKey: "risk_id",
    header: "Risk ID",
    cell: (risk: any) => {
      const rawObject = toRaw(risk.row.original);
      return rawObject.no;
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
      const rawObject = toRaw(risk.row.original);
      return rawObject.risk_name;
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
      const rawObject = toRaw(risk.row.original);
      return rawObject.risk_category;
    },
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center text-black",
      },
    },
  },
  {
    header: "Risk Residual Status",
    meta: {
      class: {
        th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
        td: "text-center place-self-center text-black",
      },
    },
    columns: [
      {
        header: "Impact",
        meta: {
          class: {
            th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
            td: "text-center place-self-center text-black",
          },
        },
        columns: [
          {
            accessorKey: "impact_q1",
            header: "Q1",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.impact_q1;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "impact_q2",
            header: "Q2",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.impact_q2;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "impact_q3",
            header: "Q3",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.impact_q3;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "impact_q4",
            header: "Q4",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.impact_q4;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
        ],
      },
      {
        header: "Possibility",
        meta: {
          class: {
            th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
            td: "text-center place-self-center text-black",
          },
        },
        columns: [
          {
            accessorKey: "possibility_q1",
            header: "Q1",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.possibility_q1;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "possibility_q2",
            header: "Q2",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.possibility_q2;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "possibility_q3",
            header: "Q3",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.possibility_q3;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "possibility_q4",
            header: "Q4",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              return rawObject.possibility_q4;
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
        ],
      },
      {
        header: "Risk Level",
        meta: {
          class: {
            th: "bg-primary text-secondary-900 text-center place-self-center px-6 py-3 text-xs leading-4 font-medium uppercase tracking-wider",
            td: "text-center place-self-center text-black",
          },
        },
        columns: [
          {
            accessorKey: "risk_level_q1",
            header: "Q1",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              var result = calculateRiskLevel(
                rawObject.impact_q1,
                rawObject.possibility_q1,
              );
              var label = getRiskLevelOptions().findLast(
                (opt, index) => opt.value === result,
              )?.label;

              const selectedColor = Object.entries(riskColor).find(([key]) => {
                const clearedKey = label?.replace(/\s+\d+$/, "");
                return clearedKey === key;
              })?.[1];

              return h(
                UBadge,
                { class: "capitalize", variant: "soft", color: selectedColor },
                () => label,
              );
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "risk_level_q2",
            header: "Q2",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              var result = calculateRiskLevel(
                rawObject.impact_q2,
                rawObject.possibility_q2,
              );

              var label = getRiskLevelOptions().findLast(
                (opt, index) => opt.value === result,
              )?.label;

              const selectedColor = Object.entries(riskColor).find(([key]) => {
                const clearedKey = label?.replace(/\s+\d+$/, "");
                return clearedKey === key;
              })?.[1];

              return h(
                UBadge,
                { class: "capitalize", variant: "soft", color: selectedColor },
                () => label,
              );
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "risk_level_q3",
            header: "Q3",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              var result = calculateRiskLevel(
                rawObject.impact_q3,
                rawObject.possibility_q3,
              );
              var label = getRiskLevelOptions().findLast(
                (opt, index) => opt.value === result,
              )?.label;

              const selectedColor = Object.entries(riskColor).find(([key]) => {
                const clearedKey = label?.replace(/\s+\d+$/, "");
                return clearedKey === key;
              })?.[1];

              return h(
                UBadge,
                { class: "capitalize", variant: "soft", color: selectedColor },
                () => label,
              );
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
          {
            accessorKey: "risk_level_q4",
            header: "Q4",
            cell: (risk: any) => {
              const rawObject = toRaw(risk.row.original);
              var result = calculateRiskLevel(
                rawObject.impact_q4,
                rawObject.possibility_q4,
              );

              var label = getRiskLevelOptions().findLast(
                (opt, index) => opt.value === result,
              )?.label;

              const selectedColor = Object.entries(riskColor).find(([key]) => {
                const clearedKey = label?.replace(/\s+\d+$/, "");
                return clearedKey === key;
              })?.[1];

              return h(
                UBadge,
                { class: "capitalize", variant: "soft", color: selectedColor },
                () => label,
              );
            },
            meta: {
              class: {
                th: "bg-primary text-secondary-900 text-center px-2 py-3 text-xs font-medium uppercase",
                td: "text-center text-black",
              },
            },
          },
        ],
      },
    ],
  },
];
</script>
