import { computed, ref, watch, type Ref, type WatchSource } from 'vue'
import { paginateList } from '~/utils/clientPagination'

// Client-side paging over an already filtered and sorted list. Any change to `resetOn`
// (search, filters, year) sends the list back to page 1.
export const usePagedList = <T>(
  list: Ref<ReadonlyArray<T>> | (() => ReadonlyArray<T>),
  resetOn: WatchSource[] = [],
  pageSize = 10
) => {
  const page = ref(1)
  const source = typeof list === 'function' ? list : () => list.value

  const pageData = computed(() => paginateList(source(), page.value, pageSize))
  // 1-based range of the rows on the current page; 0-0 when the list is empty.
  const from = computed(() => (pageData.value.total === 0 ? 0 : (pageData.value.page - 1) * pageSize + 1))
  const to = computed(() => Math.min(pageData.value.page * pageSize, pageData.value.total))

  if (resetOn.length) {
    watch(resetOn, () => {
      page.value = 1
    })
  }

  return { page, pageData, from, to, pageSize }
}
