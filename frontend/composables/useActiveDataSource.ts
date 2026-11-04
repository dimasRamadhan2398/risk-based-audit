import { storeToRefs } from "pinia";
import { useDataSourceStore } from "~/stores/data-source";

export function useActiveDataSource() {
  const store = useDataSourceStore();
  const { sources, selectedSourceId, selectedSource, isAllSources, isLoading } = storeToRefs(store);

  return {
    sources,
    selectedSourceId,
    selectedSource,
    isAllSources,
    isLoading,
    setSelectedSourceId: store.setSelectedSourceId,
    setSources: store.setSources,
    upsertSource: store.upsertSource,
    removeSource: store.removeSource,
    fetchSources: store.fetchSources
  };
}
