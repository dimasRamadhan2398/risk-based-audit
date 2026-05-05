<template>
  <UModal
    title="Perbedaan Dampak vs Tingkat Kemungkinan"
    :open="props.isOpen"
    :close="{
      color: 'primary',
      variant: 'outline',
      class: 'rounded-full',
      onClick: closeModal,
    }"
    variant="default"
    scrollable
    dismissible
  >
    <template #body>
      <div class="grid gap-16">
        <UTable :data="data" :columns="columns" />
      </div>
    </template>
  </UModal>
</template>
<script setup lang="ts">
import type { ColumnDef } from "@tanstack/vue-table";

var riskProfileStore = useRiskProfileStore();
const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  "update:isOpen": [value: boolean];
}>();

const closeModal = () => {
  console.log("close");
  emit("update:isOpen", false);
};

const columns: ColumnDef<Object>[] = [
  {
    accessorKey: "impact",
    header: ({ column: any }) => {
      return h("div", { class: "flex items-center gap-3 justify-center " }, [
        h(
          "p",
          {
            class: "text-center font-semibold text-secondary-900 bg-primary",
          },
          "Impact", // Use a static string or get from column definition
        ),
      ]);
    },
    cell: (info) => info.getValue(),
    meta: {
      class: {
        th: "text-center font-semibold bg-primary text-secondary-900 text-center",
        td: " font-medium text-center",
      },
    },
  },
  {
    accessorKey: "likelihood",
    header: ({ column: any }) => {
      return h("div", { class: "flex items-center justify-center gap-3" }, [
        h(
          "p",
          {
            class: "font-semibold text-secondary-900 bg-primary text-right",
          },
          "Likelihood", // Use a static string or get from column definition
        ),
      ]);
    },
    cell: (info) => info.getValue(),
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900",
        td: " font-medium text-center",
      },
    },
  },
];
const data = riskProfileStore.getImpactLikelihoodExplanation;
</script>
