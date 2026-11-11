// Client-side paging for lists derived from the full data set (pure helper).

export interface ClientPage<T> {
  items: T[]
  page: number
  pageSize: number
  total: number
  totalPages: number
}

// The page is clamped into 1..totalPages, so a list that shrank never shows an empty page.
export const paginateList = <T>(list: ReadonlyArray<T>, page: number, pageSize: number): ClientPage<T> => {
  const total = list.length
  const totalPages = Math.ceil(total / pageSize)
  const current = Math.min(Math.max(1, Math.floor(page) || 1), Math.max(1, totalPages))
  const start = (current - 1) * pageSize
  return { items: list.slice(start, start + pageSize), page: current, pageSize, total, totalPages }
}
