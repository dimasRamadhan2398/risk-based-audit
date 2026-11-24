import { ref, onBeforeUnmount } from 'vue'
import { getAuthServiceBaseUrl } from '~/composables/useApiUrl'
import {
  createJobWatcher,
  createWatcherState,
  mergeJobSnapshot,
  type CreateJobBody,
  type NextPorts,
  type ProvisioningJob,
  type WatcherState
} from '~/utils/provisioning'

/**
 * Tenant provisioning API (auth-service /api/v1/provisioning/*, see docs/provisioning-job-contract.md).
 * All calls go through the global $fetch, so plugins/api-auth.client.ts attaches the bearer token
 * (the SSE stream included: it is read with responseType 'stream') and handles 401.
 */

const base = () => `${getAuthServiceBaseUrl()}/provisioning`

/** Accepts both a bare payload and the usual { data } envelope. */
const unwrap = <T>(res: unknown): T =>
  (res && typeof res === 'object' && !Array.isArray(res) && 'data' in res ? (res as { data: unknown }).data : res) as T

export const useProvisioningApi = () => {
  const createJob = async (body: CreateJobBody) =>
    unwrap<ProvisioningJob>(await $fetch(`${base()}/jobs`, { method: 'POST', body }))

  const listJobs = async (): Promise<ProvisioningJob[]> => {
    const res = unwrap<ProvisioningJob[] | { jobs?: ProvisioningJob[], items?: ProvisioningJob[] } | null>(await $fetch(`${base()}/jobs`))
    if (Array.isArray(res)) return res
    return res?.jobs || res?.items || []
  }

  const getJob = async (id: string) =>
    unwrap<ProvisioningJob>(await $fetch(`${base()}/jobs/${encodeURIComponent(id)}`))

  const getNextPorts = async () =>
    unwrap<NextPorts>(await $fetch(`${base()}/ports/next`))

  const openEventStream = async (id: string, since: number, signal: AbortSignal) => {
    const body = await $fetch<ReadableStream<Uint8Array>>(`${base()}/jobs/${encodeURIComponent(id)}/events`, {
      query: since > 0 ? { since } : undefined,
      headers: { Accept: 'text/event-stream' },
      responseType: 'stream',
      retry: 0,
      signal
    })
    if (!body || typeof (body as Partial<ReadableStream>).getReader !== 'function') {
      throw new Error('Streaming is not supported by this browser')
    }
    return body
  }

  return { createJob, listJobs, getJob, getNextPorts, openEventStream }
}

/**
 * Live view of one provisioning job: SSE stream with ?since= resume after a reconnect,
 * falling back to polling GET /jobs/{id} every 2s. `state` is a ref to the reactive watcher state.
 * The watcher stops when the job ends, on detach()/reset(), or when the calling component unmounts.
 */
export const useProvisioningJob = () => {
  const api = useProvisioningApi()
  // A fresh state object per attach, so a stopped watcher's late response cannot touch the next job
  const state = ref<WatcherState>(createWatcherState())
  let watcher: ReturnType<typeof createJobWatcher> | null = null

  const detach = () => {
    watcher?.stop()
    watcher = null
  }

  const reset = () => {
    detach()
    state.value = createWatcherState()
  }

  const attach = (jobId: string, initial?: ProvisioningJob) => {
    reset()
    if (initial) mergeJobSnapshot(state.value, initial)
    watcher = createJobWatcher(jobId, {
      getJob: api.getJob,
      openStream: api.openEventStream
    }, state.value)
    void watcher.start()
  }

  onBeforeUnmount(detach)

  return { state, attach, detach, reset }
}
