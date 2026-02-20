import { defineStore } from "pinia";
import type {
  RiskProfile,
  RiskProfileItem,
  ImpactLevel,
  PossibilityLevel,
} from "~/types/risk";
import { generatedRiskProfileId } from "~/utils/structuredId";
import {
  getPrompt,
  type PromptKey,
  type PromptVariables,
} from "~/utils/prompts";
import type { OpenAICompletionResponse } from "~/server/api/openai";

export interface RiskMatrixCell {
  name: string;
  riskId1: string;
  riskId2: string;
  riskId3: string;
  riskId4: string;
  riskId5: string;
}

export interface RiskListItem {
  risk_id: string;
  risk_name: string;
  risk_category: string;
  risk_level: string;
  list_residual_risks: RiskPoint[];
  latest_impact_level: number;
  latest_possibility_level: number;
  conclusion: string;
}

export interface RiskPoint {
  impact_level: number;
  possibility_level: number;
}

export class RiskPointModel implements RiskPoint {
  constructor(
    public impact_level: number,
    public possibility_level: number,
  ) {}

  get risk_level(): number {
    return this.impact_level * this.possibility_level;
  }
}

export interface CreateRiskProfilePayload {
  risk_name: string;
  risk_category: string;
  list_residual_risks: RiskPoint[];
}

export interface ImpactLikelihoodExplanation {
  impact: string;
  likelihood: string;
}

export interface RiskActiveState{
  selectedRiskName: string;
  selectedRiskId: string;
  selectedRiskValue: RiskListItem | null;
}

interface RiskProfileState {
  registeredRisk: RiskListItem[];
  riskMatrix: RiskMatrixCell[];
  riskList: RiskListItem[];
  riskState: RiskActiveState;
  impactLikelihoodExplanation: ImpactLikelihoodExplanation[];
  loading: boolean;
  error: string | null;
}

export const useRiskProfileStore = defineStore("risk-profile", {
  state: (): RiskProfileState => ({
    riskState: {
      selectedRiskName: "",
      selectedRiskId: "",
      selectedRiskValue: null
    },
    registeredRisk: [],
    riskMatrix: [
      {
        name: "Hampir Pasti Terjadi",
        riskId1: "Low to Moderate 7",
        riskId2: "Moderate 12",
        riskId3: "Moderate to High 17",
        riskId4: "High 22",
        riskId5: "High 25",
      },
      {
        name: "Sangat Mungkin Terjadi",
        riskId1: "Low 4",
        riskId2: "Low to Moderate 9",
        riskId3: "Moderate 14",
        riskId4: "Moderate to High 19",
        riskId5: "High 24",
      },
      {
        name: "Bisa Terjadi",
        riskId1: "Low 3",
        riskId2: "Low to Moderate 8",
        riskId3: "Moderate 13",
        riskId4: "Moderate to High 18",
        riskId5: "High 23",
      },
      {
        name: "Jarang Terjadi",
        riskId1: "Low 2",
        riskId2: "Low to Moderate 6",
        riskId3: "Low to Moderate 11",
        riskId4: "Moderate to High 16",
        riskId5: "High 21",
      },
      {
        name: "Sangat Jarang Terjadi",
        riskId1: "Low 1",
        riskId2: "Low 5",
        riskId3: "Low to Moderate 10",
        riskId4: "Moderate 15",
        riskId5: "High 20",
      },
    ],
    riskList: [
      {
        risk_id: "FIN-001",
        risk_name: "Fluktuasi Nilai Tukar Mata Uang",
        risk_category: "Financial",
        risk_level: "High",
        list_residual_risks: [
          { impact_level: 5, possibility_level: 5 },
          { impact_level: 4, possibility_level: 3 },
          { impact_level: 4, possibility_level: 4 },
          { impact_level: 5, possibility_level: 2 },
        ],
        latest_impact_level: 5,
        latest_possibility_level: 5,
        conclusion: "",
      },
      {
        risk_id: "OPS-001",
        risk_name: "Gangguan Sistem IT",
        risk_category: "Operational",
        risk_level: "Moderate to High",
        list_residual_risks: [{ impact_level: 4, possibility_level: 4 }],
        latest_impact_level: 4,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "COM-001",
        risk_name: "Ketidakpatuhan Regulasi GDPR",
        risk_category: "Compliance",
        risk_level: "Moderate",
        list_residual_risks: [{ impact_level: 3, possibility_level: 3 }],
        latest_impact_level: 3,
        latest_possibility_level: 3,
        conclusion: "",
      },
      {
        risk_id: "STR-001",
        risk_name: "Perubahan Strategi Kompetitor",
        risk_category: "Strategic",
        risk_level: "Low to Moderate",
        list_residual_risks: [{ impact_level: 2, possibility_level: 4 }],
        latest_impact_level: 2,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "OPS-002",
        risk_name: "Kegagalan Rantai Pasokan",
        risk_category: "Operational",
        risk_level: "High",
        list_residual_risks: [{ impact_level: 5, possibility_level: 4 }],
        latest_impact_level: 5,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "FIN-002",
        risk_name: "Risiko Kredit Pelanggan",
        risk_category: "Financial",
        risk_level: "Moderate",
        list_residual_risks: [{ impact_level: 3, possibility_level: 4 }],
        latest_impact_level: 3,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "SEC-001",
        risk_name: "Serangan Siber dan Ransomware",
        risk_category: "Security",
        risk_level: "High",
        list_residual_risks: [{ impact_level: 5, possibility_level: 4 }],
        latest_impact_level: 5,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "REP-001",
        risk_name: "Penurunan Reputasi Brand",
        risk_category: "Reputational",
        risk_level: "Moderate to High",
        list_residual_risks: [{ impact_level: 4, possibility_level: 3 }],
        latest_impact_level: 4,
        latest_possibility_level: 3,
        conclusion: "",
      },
      {
        risk_id: "ENV-001",
        risk_name: "Dampak Perubahan Iklim",
        risk_category: "Environmental",
        risk_level: "Low to Moderate",
        list_residual_risks: [{ impact_level: 3, possibility_level: 2 }],
        latest_impact_level: 3,
        latest_possibility_level: 2,
        conclusion: "",
      },
      {
        risk_id: "HR-001",
        risk_name: "Tingkat Turnover Karyawan Tinggi",
        risk_category: "Human Resources",
        risk_level: "Moderate",
        list_residual_risks: [{ impact_level: 3, possibility_level: 5 }],
        latest_impact_level: 3,
        latest_possibility_level: 5,
        conclusion: "",
      },
      {
        risk_id: "FIN-003",
        risk_name: "Kenaikan Harga Bahan Baku",
        risk_category: "Financial",
        risk_level: "Moderate to High",
        list_residual_risks: [{ impact_level: 4, possibility_level: 4 }],
        latest_impact_level: 4,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "OPS-003",
        risk_name: "Kerusakan Peralatan Produksi",
        risk_category: "Operational",
        risk_level: "Moderate",
        list_residual_risks: [{ impact_level: 3, possibility_level: 3 }],
        latest_impact_level: 3,
        latest_possibility_level: 3,
        conclusion: "",
      },
      {
        risk_id: "COM-002",
        risk_name: "Ketidakpatuhan Pajak",
        risk_category: "Compliance",
        risk_level: "High",
        list_residual_risks: [{ impact_level: 5, possibility_level: 3 }],
        latest_impact_level: 5,
        latest_possibility_level: 3,
        conclusion: "",
      },
      {
        risk_id: "STR-002",
        risk_name: "Kegagalan Inovasi Produk",
        risk_category: "Strategic",
        risk_level: "Moderate",
        list_residual_risks: [{ impact_level: 4, possibility_level: 3 }],
        latest_impact_level: 4,
        latest_possibility_level: 3,
        conclusion: "",
      },
      {
        risk_id: "SEC-002",
        risk_name: "Kebocoran Data Pelanggan",
        risk_category: "Security",
        risk_level: "High",
        list_residual_risks: [{ impact_level: 5, possibility_level: 4 }],
        latest_impact_level: 5,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "REP-002",
        risk_name: "Krisis Hubungan Masyarakat",
        risk_category: "Reputational",
        risk_level: "Moderate to High",
        list_residual_risks: [{ impact_level: 4, possibility_level: 4 }],
        latest_impact_level: 4,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "ENV-002",
        risk_name: "Pembuangan Limbah Tidak Sesuai Standar",
        risk_category: "Environmental",
        risk_level: "High",
        list_residual_risks: [{ impact_level: 5, possibility_level: 3 }],
        latest_impact_level: 5,
        latest_possibility_level: 3,
        conclusion: "",
      },
      {
        risk_id: "HR-002",
        risk_name: "Kekurangan Tenaga Kerja Terampil",
        risk_category: "Human Resources",
        risk_level: "Moderate",
        list_residual_risks: [{ impact_level: 3, possibility_level: 4 }],
        latest_impact_level: 3,
        latest_possibility_level: 4,
        conclusion: "",
      },
      {
        risk_id: "LEG-001",
        risk_name: "Sengketa Kekayaan Intelektual",
        risk_category: "Legal",
        risk_level: "Moderate to High",
        list_residual_risks: [{ impact_level: 4, possibility_level: 4 }],
        latest_impact_level: 4,
        latest_possibility_level: 4,
        conclusion: "",
      },
    ],
    impactLikelihoodExplanation: [
      {
        impact: "Membutuhkan biaya untuk penyelesaian",
        likelihood:
          "Perubahan stabilitas, termasuk ketidakpastian akibat risiko-risiko",
      },
      {
        impact: "Berpotensi merusak reputasi perusahaan",
        likelihood: "Kemungkinan terjadi berdasarkan siklus atau pengalaman",
      },
      {
        impact: "Berpengaruh dari segi materi",
        likelihood: "Menunjukkan Efektivitas dari proses kontrol",
      },
    ],
    loading: false,
    error: null,

  }),

  getters: {
    getRiskMatrix: (state) => state.riskMatrix,
    getRiskList: (state) => state.riskList,
    getRiskState: (state) => state.riskState,
    isLoading: (state) => state.loading,
    getError: (state) => state.error,

    getRiskById: (state) => (id: string) => {
      return state.riskList.find((risk) => risk.risk_id === id);
    },

    getRisksByCategory: (state) => (category: string) => {
      return state.riskList.filter((risk) => risk.risk_category === category);
    },

    getHighRisks: (state) => {
      return state.riskList.filter(
        (risk) =>
          risk.risk_level === "High" || risk.risk_level === "Moderate to High",
      );
    },
    getImpactLikelihoodExplanation: (state) =>
      state.impactLikelihoodExplanation,

    getRegisteredRisks: (state) => state.registeredRisk
  },

  actions: {
    async fetchRiskProfiles() {
      this.loading = true;
      this.error = null;

      try {
        const response = await $fetch<RiskListItem[]>("/api/risk-profile", {
          method: "GET",
        });

        this.riskList = response;
      } catch (error: any) {
        this.error = error.message || "Failed to fetch risk profiles";
        console.error("Error fetching risk profiles:", error);
      } finally {
        this.loading = false;
      }
    },

    async createRiskProfile(payload: CreateRiskProfilePayload) {
      this.loading = true;
      this.error = null;

      try {
        // const response = await $fetch<RiskListItem>("/api/risk-profile", {
        //   method: "POST",
        //   body: payload,
        // });

        const id = generatedRiskProfileId(
          payload.risk_category,
          this.riskList.length,
        );

        const riskPoint =
          payload.list_residual_risks[payload.list_residual_risks.length - 1];

        const conclusion = await this.createConclusion("risk-conclusion", {
          category: payload.risk_category,
          name: payload.risk_name,
          impact: riskPoint?.impact_level!,
          possibility: riskPoint?.possibility_level!,
          riskLevel: this.calculateRiskLevel(
            riskPoint?.impact_level!,
            riskPoint?.possibility_level!,
          ),
          historyCount: payload.list_residual_risks.length,
          trend: "stable",
        });

        this.riskList.push({
          risk_id: id,
          risk_name: payload.risk_name,
          risk_category: payload.risk_category,
          risk_level: this.calculateRiskLevel(
            riskPoint?.impact_level!,
            riskPoint?.possibility_level!,
          ),
          list_residual_risks: payload.list_residual_risks,
          latest_impact_level: riskPoint?.impact_level!,
          latest_possibility_level: riskPoint?.possibility_level!,
          conclusion: conclusion,
        });
      } catch (error: any) {
        this.error = error.message || "Failed to create risk profile";
        console.error("Error creating risk profile:", error);
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async updateRiskProfile(
      id: string,
      payload: Partial<CreateRiskProfilePayload>,
    ) {
      this.loading = true;
      this.error = null;

      try {
        const response = await $fetch<RiskListItem>(`/api/risk-profile/${id}`, {
          method: "PUT",
          body: payload,
        });

        const index = this.riskList.findIndex((risk) => risk.risk_id === id);
        if (index !== -1) {
          this.riskList[index] = response;
        }

        return response;
      } catch (error: any) {
        this.error = error.message || "Failed to update risk profile";
        console.error("Error updating risk profile:", error);
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async deleteRiskProfile(id: string) {
      this.loading = true;
      this.error = null;

      try {
        // await $fetch(`/api/risk-profile/${id}`, {
        //   method: "DELETE",
        // });
        this.riskList = this.riskList.filter((risk) => risk.risk_id !== id);
      } catch (error: any) {
        this.error = error.message || "Failed to delete risk profile";
        console.error("Error deleting risk profile:", error);
        throw error;
      } finally {
        this.loading = false;
      }
    },

    calculateRiskLevel(impactLevel: number, possibilityLevel: number): string {
      const riskScore = impactLevel * possibilityLevel;

      if (riskScore <= 4) return "Low";
      if (riskScore <= 10) return "Low to Moderate";
      if (riskScore <= 15) return "Moderate";
      if (riskScore <= 20) return "Moderate to High";
      return "High";
    },

    createRiskPoint(
      impact_level: number,
      possibility_level: number,
    ): RiskPoint {
      return {
        impact_level,
        possibility_level,
      };
    },

    async createConclusion(promptKey: PromptKey, variables: PromptVariables) {
      try {
        // Generate the prompt with variables interpolated
        const prompt = getPrompt(promptKey, variables);

        const result = await $fetch<OpenAICompletionResponse>(
          "api/completion",
          {
            method: "POST",
            body: { prompt },
          },
        );
        return result.choices[0]!.text;
      } catch (e) {
        console.error("Error creating conclusion:", e);
        throw e;
      }
    },

    async generateRiskConclusion(riskId: string) {
      const risk = this.getRiskById(riskId);
      if (!risk) {
        throw new Error(`Risk with ID ${riskId} not found`);
      }

      const conclusion = await this.createConclusion("risk-conclusion", {
        category: risk.risk_category,
        name: risk.risk_name,
        impact: risk.latest_impact_level,
        possibility: risk.latest_possibility_level,
        riskLevel: risk.risk_level,
        historyCount: risk.list_residual_risks.length,
        trend: "stable",
      });

      return conclusion;
    },

    clearError() {
      this.error = null;
    },

    setSelectedRiskId(id: string) {
      this.riskState.selectedRiskId = id;
    },

    setSelectedRiskName(name: string) {
      this.riskState.selectedRiskName = name;
    },

    setSelectedRiskValue(risk: RiskListItem | null) {
      this.riskState.selectedRiskValue = risk;
      if (risk) {
        this.riskState.selectedRiskId = risk.risk_id;
        this.riskState.selectedRiskName = risk.risk_name;
      }
    },

    clearSelectedRisk() {
      this.riskState.selectedRiskId = "";
      this.riskState.selectedRiskName = "";
      this.riskState.selectedRiskValue = null;
    },
  },
});
