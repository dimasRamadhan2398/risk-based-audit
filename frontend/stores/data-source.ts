import { defineStore } from "pinia";
import { ref, computed } from "vue";

export interface DataSourceItem {
  id: string;
  source_id?: string;
  name: string;
  source_type: string;
  host: string;
  port: number;
  database_name?: string;
  database?: string;
  status?: string;
  records_synced?: number;
  last_sync_at?: string;
}

export const useDataSourceStore = defineStore("data-source", () => {
  const sources = ref<DataSourceItem[]>([]);
  const selectedSourceId = ref<string>("all");
  const isLoading = ref<boolean>(false);

  // Restore from localStorage if in client
  if (typeof window !== "undefined") {
    const saved = localStorage.getItem("auditsphere_selected_source_id");
    if (saved) {
      selectedSourceId.value = saved;
    }
  }

  const selectedSource = computed(() => {
    if (selectedSourceId.value === "all") return null;
    return sources.value.find(
      (s) => s.id === selectedSourceId.value || s.source_id === selectedSourceId.value
    ) || null;
  });

  const isAllSources = computed(() => selectedSourceId.value === "all");

  function setSelectedSourceId(id: string) {
    selectedSourceId.value = id;
    if (typeof window !== "undefined") {
      localStorage.setItem("auditsphere_selected_source_id", id);
      // Sync with URL query parameter smoothly
      const url = new URL(window.location.href);
      if (id === "all") {
        url.searchParams.delete("source_id");
      } else {
        url.searchParams.set("source_id", id);
      }
      window.history.replaceState({}, "", url.toString());
    }
  }

  async function fetchSources() {
    isLoading.value = true;
    try {
      const config = useRuntimeConfig();
      const baseUrl = config.public?.masterServiceUrl || "http://localhost:8001/api/v1";
      const res: any = await $fetch(`${baseUrl}/data-sources`).catch(() => null);
      if (res && res.data && Array.isArray(res.data)) {
        sources.value = res.data;
      } else {
        // Fallback default simulator source if API not yet populated
        sources.value = [
          {
            id: "cbs_simulator",
            source_id: "cbs_simulator",
            name: "Core Banking Simulator",
            source_type: "postgres",
            host: "postgres-timescale",
            port: 5432,
            database: "auditsphere_cbs",
            status: "synced"
          }
        ];
      }
    } catch (e) {
      console.warn("Failed to fetch sources, using fallback", e);
    } finally {
      isLoading.value = false;
    }
  }

  return {
    sources,
    selectedSourceId,
    selectedSource,
    isAllSources,
    isLoading,
    setSelectedSourceId,
    fetchSources
  };
});
