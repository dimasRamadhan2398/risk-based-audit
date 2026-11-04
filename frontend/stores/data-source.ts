import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { getMasterServiceBaseUrl } from "~/composables/useApiUrl";

export interface DataSourceItem {
  id: string;
  source_id?: string;
  name: string;
  source_type: string;
  type?: string;
  host: string;
  port: number;
  database_name?: string;
  database?: string;
  environment?: string;
  status?: string;
  syncSchedule?: string;
  sync_schedule?: string;
  records_synced?: number;
  last_sync_at?: string;
  lastSync?: string;
  scopes?: string[];
  dataMappings?: any[];
}

const STORAGE_KEY_SELECTED = "auditsphere_selected_source_id";
const STORAGE_KEY_SOURCES = "auditsphere_registered_sources";

export const useDataSourceStore = defineStore("data-source", () => {
  const sources = ref<DataSourceItem[]>([]);
  const selectedSourceId = ref<string>("all");
  const isLoading = ref<boolean>(false);

  // Restore from localStorage if in client
  if (typeof window !== "undefined") {
    try {
      const savedSelected = localStorage.getItem(STORAGE_KEY_SELECTED);
      if (savedSelected) {
        selectedSourceId.value = savedSelected;
      }
      const savedSources = localStorage.getItem(STORAGE_KEY_SOURCES);
      if (savedSources) {
        const parsed = JSON.parse(savedSources);
        if (Array.isArray(parsed) && parsed.length > 0) {
          sources.value = parsed;
        }
      }
    } catch {
      // Ignore JSON parse errors
    }
  }

  const selectedSource = computed(() => {
    if (selectedSourceId.value === "all") return null;
    return (
      sources.value.find(
        (s) => s.id === selectedSourceId.value || s.source_id === selectedSourceId.value
      ) || null
    );
  });

  const isAllSources = computed(() => selectedSourceId.value === "all");

  function saveSourcesToStorage(list: DataSourceItem[]) {
    if (typeof window !== "undefined") {
      try {
        localStorage.setItem(STORAGE_KEY_SOURCES, JSON.stringify(list));
      } catch {
        // storage quota fallback
      }
    }
  }

  function setSelectedSourceId(id: string) {
    selectedSourceId.value = id;
    if (typeof window !== "undefined") {
      localStorage.setItem(STORAGE_KEY_SELECTED, id);
      const url = new URL(window.location.href);
      if (id === "all") {
        url.searchParams.delete("source_id");
      } else {
        url.searchParams.set("source_id", id);
      }
      window.history.replaceState({}, "", url.toString());
    }
  }

  function setSources(list: DataSourceItem[]) {
    sources.value = list;
    saveSourcesToStorage(list);
  }

  function upsertSource(item: DataSourceItem) {
    const sid = item.id || item.source_id;
    const idx = sources.value.findIndex((s) => s.id === sid || s.source_id === sid);
    if (idx !== -1) {
      sources.value[idx] = { ...sources.value[idx], ...item };
    } else {
      sources.value.unshift(item);
    }
    saveSourcesToStorage(sources.value);
  }

  function removeSource(id: string) {
    sources.value = sources.value.filter((s) => s.id !== id && s.source_id !== id);
    if (selectedSourceId.value === id) {
      setSelectedSourceId("all");
    }
    saveSourcesToStorage(sources.value);
  }

  async function fetchSources() {
    isLoading.value = true;
    try {
      let baseUrl = "";
      try {
        baseUrl = getMasterServiceBaseUrl();
      } catch {
        const config = useRuntimeConfig();
        baseUrl = config.public?.masterServiceUrl || config.public?.masterServiceBaseUrl || "/api/v1";
      }

      // Add 4s timeout to avoid freezing
      const controller = typeof AbortController !== "undefined" ? new AbortController() : null;
      const timeoutId = controller ? setTimeout(() => controller.abort(), 4000) : null;

      const res: any = await $fetch(`${baseUrl}/data-sources`, {
        signal: controller ? controller.signal : undefined,
      }).catch(() => null);

      if (timeoutId) clearTimeout(timeoutId);

      if (res && res.data && Array.isArray(res.data) && res.data.length > 0) {
        sources.value = res.data;
        saveSourcesToStorage(res.data);
      } else if (sources.value.length === 0) {
        // If API returned empty and no cached sources, keep empty array
        sources.value = [];
      }
    } catch (e) {
      console.warn("Failed to fetch sources from API, preserving cached state:", e);
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
    setSources,
    upsertSource,
    removeSource,
    fetchSources,
  };
});
