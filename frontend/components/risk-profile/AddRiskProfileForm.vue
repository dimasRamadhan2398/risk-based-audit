<template>
  <UModal
    :title="formTitle"
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
      <UForm @submit.prevent="addRiskProfile">
        <div class="grid gap-16">
          <div class="flex flex-col justify-between gap-8">
            <UFormField
              orientation="horizontal"
              label="Peristiwa Resiko"
              help="Masukkan kejadian yang ingin ditinjau yang berpotensi menimbulkan dampak negatif terhadap tujuan, proses, atau aset."
              class="text-sm font-medium"
              size="xl"
              :ui="{
                root: 'wrap-break-word flex flex-row flex-1 w-full basis-0.5',
                container: ' gap-4 items-start max-w-full',
                label: 'block font-medium text-black min-w-24 max-w-24',
              }"
            >
              <UInput
                type="text"
                name="risk_name"
                id="risk_name"
                class="mt-1 rounded flex basis-0.5"
                v-model="riskProfile.risk_name"
              />
            </UFormField>
            <!-- <div>
            <label
              for="impact_level"
              class="block text-sm font-medium text-gray-700"
              >Impact Level</label
            >
            <select
              id="impact_level"
              name="impact_level"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring focus:ring-indigo-500"
              v-model="riskProfile.impact_level"
            >
              <option :value="1">Very Low</option>
              <option :value="2">Low</option>
              <option :value="3">Moderate</option>
              <option :value="4">High</option>
              <option :value="5">Very High</option>
            </select> -->
            <!-- </div> -->
            <UFormField
              orientation="horizontal"
              label="Kategori Peristiwa"
              help="Pengelompokan jenis risiko berdasarkan sumber atau sifatnya (mis. operasional, keuangan, kepatuhan, strategis)."
              class="text-sm font-medium"
              size="xl"
              :ui="{
                root: 'wrap-break-word flex flex-row flex-1 w-full',
                container: ' gap-4 items-start max-w-full',
                label: 'block font-medium text-black min-w-24 max-w-24',
              }"
            >
              <UInput
                type="text"
                name="risk_category"
                id="risk_category"
                class="mt-1 rounded flex basis-1/2"
                v-model="riskProfile.risk_category"
              />
            </UFormField>
          </div>
          <div class="flex flex-col justify-start gap-4">
            <UCheckbox
              label="Buatkan mode pengisian per kuartal"
              color="primary"
              v-model="quarterlyActive"
            >
            </UCheckbox>
            <UTabs
              :content="false"
              :items="tabItems"
              v-if="quarterlyActive"
              v-model="tabActive"
            />
            <UFormField
              label="Skala Dampak"
              description="Ukuran besarnya akibat yang ditimbulkan jika risiko terjadi"
              class="block text-sm font-medium text-gray-700"
            >
              <USelect
                name="impact_level"
                class="mt-1 block w-full min-w-max rounded text-left"
                v-model="currentImpactLevel"
                :items="impactLevelItems"
              >
              </USelect>
            </UFormField>
            <UFormField
              label="Skala Kemungkinan"
              description="Tingkat peluang terjadinya risiko dalam periode tertentu"
              class="block text-sm font-medium text-gray-700"
            >
              <USelect
                name="possibility_level"
                class="mt-1 block w-full min-w-max rounded text-left"
                v-model="currentPossibilityLevel"
                :items="possibilityLevelItems"
              >
              </USelect>
            </UFormField>
          </div>
        </div>
      </UForm>
    </template>
    <template #footer>
      <div></div>
      <UButton @click="handleSubmit" color="primary">"Confirm Risk Profile</UButton>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import type { TabsItem } from "@nuxt/ui";
import { ref, computed, watch } from "vue";
import { useRiskProfileStore } from "~/stores/profile-risk";

const props = defineProps<{
  isOpen: boolean;
  editMode: boolean;
  riskData?: RiskListItem;
}>();

watch(props, () => {
  if (props.editMode) {
    riskProfile.value = {
      impact_level: props.riskData!.latest_impact_level,
      possibility_level: props.riskData!.latest_possibility_level,
      risk_name: props.riskData!.risk_name,
      risk_category: props.riskData!.risk_category,
    };

    if (props.riskData!.list_residual_risks.length > 1) {
      quarterlyActive.value = true;
      quarterlyRiskProfiles.value = props.riskData!.list_residual_risks.map(
        (risk) => {
          return {
            impact_level: risk.impact_level,
            possibility_level: risk.possibility_level,
          };
        },
      );
    }
  }else{
    riskProfile.value = {
      risk_name: "",
      risk_category: "",
      impact_level: 1,
      possibility_level: 1,
    }
    quarterlyActive.value = false;
  }
});

const emit = defineEmits<{
  "update:isOpen": [value: boolean];
}>();

const riskProfileStore = useRiskProfileStore();

const formTitle = computed(() => {
  return props.editMode ? "Edit Risk Priority List" : "Add Risk Priority List";
});

onMounted(() => {
  console.log("AddRiskProfileForm mounted");
});
interface RiskProfile {
  risk_name?: string;
  risk_category?: string;
  impact_level: number;
  possibility_level: number;
}

const impactLevelItems = computed(() => [
  {
    value: 1,
    label: "1 - Very Low",
  },
  {
    value: 2,
    label: "2 - Low",
  },
  {
    value: 3,
    label: "3 - Moderate",
  },
  {
    value: 4,
    label: "4 - High",
  },
  {
    value: 5,
    label: "5 - Very High",
  },
]);

const possibilityLevelItems = computed(() => [
  {
    value: 1,
    label: "1 - Rare",
  },
  {
    value: 2,
    label: "2 - Unlikely",
  },
  {
    value: 3,
    label: "3 - Possible",
  },
  {
    value: 4,
    label: "4 - Very Likely",
  },
  {
    value: 5,
    label: "5 - Almost Certain",
  },
]);

const tabItems = ref<TabsItem[]>([
  {
    label: "Q1",
  },
  {
    label: "Q2",
  },
  {
    label: "Q3",
  },
  {
    label: "Q4",
  },
]);

const riskProfile = ref<RiskProfile>({
  risk_name: "",
  risk_category: "",
  impact_level: 4,
  possibility_level: 4,
});

const quarterlyRiskProfiles = ref<RiskProfile[]>([
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

const quarterlyActive = ref(false);
const tabActiveIndex = ref(0);

const tabActive = computed({
  get: () => tabActiveIndex.value.toString(),
  set: (value: string | number) => {
    tabActiveIndex.value =
      typeof value === "string" ? parseInt(value, 10) : value;
  },
});

watch(tabActive, (value) => {
  console.log(value);
})

const currentImpactLevel = computed<number>(
  {
    get: () =>
      quarterlyActive.value
        ? quarterlyRiskProfiles.value![tabActiveIndex.value]!.impact_level
        : riskProfile.value.impact_level,
    set: (value: number) => {
      if (quarterlyActive.value) {
        quarterlyRiskProfiles.value![tabActiveIndex.value]!.impact_level = value;
      } else {
        riskProfile.value.impact_level = value;
      }
    },
  }
)

const currentPossibilityLevel = computed<number>(
    {
    get: () =>
      quarterlyActive.value
        ? quarterlyRiskProfiles.value![tabActiveIndex.value]!.possibility_level
        : riskProfile.value.possibility_level,
    set: (value: number) => {
      if (quarterlyActive.value) {
        quarterlyRiskProfiles.value![tabActiveIndex.value]!.possibility_level = value;
      } else {
        riskProfile.value.possibility_level = value;
      }
    },
  }
);

// // Sync values when switching between quarterly/non-quarterly or changing tabs
// watch([quarterlyActive, tabActiveIndex], () => {
//   if (quarterlyActive.value) {
//     currentImpactLevel.value =
//       quarterlyRiskProfiles.value![tabActiveIndex.value]!.impact_level;
//     currentPossibilityLevel.value =
//       quarterlyRiskProfiles.value![tabActiveIndex.value]!.possibility_level;
//   } else {
//     currentImpactLevel.value = riskProfile.value.impact_level;
//     currentPossibilityLevel.value = riskProfile.value.possibility_level;
//   }
// });

// Sync impact level changes back to source
watch(currentImpactLevel, (newValue) => {
  if (quarterlyActive.value) {
    quarterlyRiskProfiles.value![tabActiveIndex.value]!.impact_level = newValue;
  } else {
    riskProfile.value.impact_level = newValue;
  }
});

// Sync possibility level changes back to source
watch(currentPossibilityLevel, (newValue) => {
  if (quarterlyActive.value) {
    quarterlyRiskProfiles.value![tabActiveIndex.value]!.possibility_level =
      newValue;
  } else {
    riskProfile.value.possibility_level = newValue;
  }
});

const closeModal = () => {
  console.log("close");
  emit("update:isOpen", false);
  riskProfile.value = {
    risk_name: "",
    risk_category: "",
    impact_level: 1,
    possibility_level: 1,
  };
};

const handleSubmit = () => {
  addRiskProfile();
};

const addRiskProfile = async () => {
  try {
    let payload: RiskProfile[] | CreateRiskProfilePayload =
      quarterlyRiskProfiles.value;
    if (!quarterlyActive.value) {
      payload = {
        risk_name: riskProfile.value.risk_name!,
        risk_category: riskProfile.value.risk_category!,
        list_residual_risks: [
          {
            impact_level: riskProfile.value.impact_level,
            possibility_level: riskProfile.value.possibility_level,
          },
        ],
      };
      await riskProfileStore.createRiskProfile(
        payload as CreateRiskProfilePayload,
      );
    }
    // For quarterly, creation logic may be implemented separately

    // Reset form and close modal on success
    riskProfile.value = {
      risk_name: "",
      risk_category: "",
      impact_level: 1,
      possibility_level: 1,
    };
    closeModal();
  } catch (error) {
    console.error("Error adding risk profile:", error);
    // You can add a toast notification here if needed
  }
};
</script>
