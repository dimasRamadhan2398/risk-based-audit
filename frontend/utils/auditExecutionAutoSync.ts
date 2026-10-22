export interface AutoSyncContext {
  /** The user may run the sync (assignment letter write access). */
  canSync: boolean
  /** The list fetch failed; do not sync on top of a broken load. */
  hasError: boolean
  /** Number of executions the list returned. */
  count: number
}

/**
 * Runs the assignment-letter sync at most once per page visit, and only when the list
 * loaded, came back empty and the user may sync. The flag is set before the sync runs,
 * so a sync that fails or still leaves the list empty (no letters yet) cannot loop.
 * Create one per page mount; a new visit gets a new guard.
 */
export function createAutoSyncOnce<T>(run: () => Promise<T>) {
  let attempted = false
  return {
    get attempted() {
      return attempted
    },
    async maybeRun(ctx: AutoSyncContext): Promise<T | undefined> {
      if (attempted || !ctx.canSync || ctx.hasError || ctx.count > 0) return undefined
      attempted = true
      return run()
    }
  }
}
