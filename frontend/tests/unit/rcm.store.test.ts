// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useRCMStore } from '~/stores/rcm'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      riskServiceBaseUrl: 'http://localhost:8080/api/v1'
    }
  })
}))

global.$fetch = vi.fn()

describe('RCM Store - API handling', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    if (typeof localStorage !== 'undefined') {
      localStorage.clear()
    }
  })

  it('addRCMItem should send valid payload to backend and update item with server UUID', async () => {
    const store = useRCMStore()
    const serverCreatedItem = {
      id: '550e8400-e29b-41d4-a716-446655440000',
      risk_code: 'FIN-001',
      risk_event: 'Target revenue missed',
      control_code: 'CTL-FIN-001',
      control_description: 'Monthly review',
      control_type: 'Preventive',
      control_owner: 'Finance Lead',
      department: 'Head Office',
      year: 2026,
      design_effectiveness_rating: 4,
      operating_effectiveness_rating: 3,
      coverage_completeness_rating: 4,
      timeliness_rating: 3,
      automation_monitoring_rating: 2,
      total_weighted_score: 64
    }

    vi.mocked($fetch).mockResolvedValueOnce({
      success: true,
      message: 'RCM item created successfully',
      data: serverCreatedItem
    })

    await store.addRCMItem({
      risk_id: 'non-uuid-mock-1', // Non-UUID mock ID
      risk_code: 'FIN-001',
      risk_event: 'Target revenue missed',
      control_code: 'CTL-FIN-001',
      control_description: 'Monthly review',
      control_type: 'Preventive',
      control_owner: 'Finance Lead',
      department: 'Head Office',
      year: 2026,
      design_effectiveness_weight: 20,
      design_effectiveness_rating: 4,
      operating_effectiveness_weight: 20,
      operating_effectiveness_rating: 3,
      coverage_completeness_weight: 20,
      coverage_completeness_rating: 4,
      timeliness_weight: 20,
      timeliness_rating: 3,
      automation_monitoring_weight: 20,
      automation_monitoring_rating: 2,
      notes: ''
    })

    // Verify $fetch was called with clean payload (no invalid UUID risk_id, no temporary id)
    expect($fetch).toHaveBeenCalledWith(
      expect.stringContaining('/rcm'),
      expect.objectContaining({
        method: 'POST',
        body: expect.objectContaining({
          risk_code: 'FIN-001',
          control_code: 'CTL-FIN-001',
          control_type: 'Preventive'
        })
      })
    )

    const postCall = vi.mocked($fetch).mock.calls.find(c => c[1]?.method === 'POST')
    expect(postCall).toBeDefined()
    const callBody = postCall![1]?.body
    expect(callBody.id).toBeUndefined()
    expect(callBody.risk_id).toBeUndefined()

    // Verify local item has been updated with real server UUID
    const savedItem = store.rcmList.find(i => i.control_code === 'CTL-FIN-001')
    expect(savedItem).toBeDefined()
    expect(savedItem?.id).toBe('550e8400-e29b-41d4-a716-446655440000')
  })

  it('fetchRCMList should populate items from backend API', async () => {
    const store = useRCMStore()
    const mockList = [
      {
        id: '11111111-2222-3333-4444-555555555555',
        risk_code: 'TEC-003',
        risk_event: 'Cyber Security threat',
        control_code: 'CTL-TEC-003',
        control_description: 'MFA implementation',
        control_type: 'Preventive',
        control_owner: 'IT Security',
        department: 'Head Office',
        year: 2026,
        design_effectiveness_rating: 5,
        operating_effectiveness_rating: 4,
        coverage_completeness_rating: 4,
        timeliness_rating: 5,
        automation_monitoring_rating: 4
      }
    ]

    vi.mocked($fetch).mockResolvedValueOnce({
      success: true,
      data: mockList
    })

    await store.fetchRCMList()

    expect($fetch).toHaveBeenCalledWith(expect.stringContaining('/rcm'))
    expect(store.rcmList.length).toBe(1)
    expect(store.rcmList[0].id).toBe('11111111-2222-3333-4444-555555555555')
    expect(store.rcmList[0].total_weighted_score).toBe(88) // 20 + 16 + 16 + 20 + 16 = 88%
  })

  it('deleteRCMItem should send DELETE to backend and remove item from list', async () => {
    const store = useRCMStore()
    store.rcmList = [
      {
        id: 'del-item-id',
        risk_code: 'FIN-001',
        risk_event: 'Event',
        control_code: 'CTL-1',
        control_description: 'Desc',
        control_owner: 'Owner',
        department: 'Head Office',
        year: 2026,
        design_effectiveness_weight: 20,
        design_effectiveness_rating: 3,
        operating_effectiveness_weight: 20,
        operating_effectiveness_rating: 3,
        coverage_completeness_weight: 20,
        coverage_completeness_rating: 3,
        timeliness_weight: 20,
        timeliness_rating: 3,
        automation_monitoring_weight: 20,
        automation_monitoring_rating: 3,
        total_weighted_score: 60
      }
    ]

    vi.mocked($fetch).mockResolvedValueOnce({
      success: true,
      message: 'RCM item deleted successfully'
    })

    await store.deleteRCMItem('del-item-id')

    expect($fetch).toHaveBeenCalledWith(
      expect.stringContaining('/rcm/del-item-id'),
      expect.objectContaining({ method: 'DELETE' })
    )
    expect(store.rcmList.find(i => i.id === 'del-item-id')).toBeUndefined()
  })
})

const row = (id: string, extra: any = {}) => ({
  id,
  risk_code: 'FIN-001',
  risk_event: 'Event',
  control_code: `CTL-${id}`,
  control_description: 'Desc',
  control_owner: 'Owner',
  department: 'Head Office',
  year: 2026,
  design_effectiveness_weight: 20,
  design_effectiveness_rating: 3,
  operating_effectiveness_weight: 20,
  operating_effectiveness_rating: 3,
  coverage_completeness_weight: 20,
  coverage_completeness_rating: 3,
  timeliness_weight: 20,
  timeliness_rating: 3,
  automation_monitoring_weight: 20,
  automation_monitoring_rating: 3,
  total_weighted_score: 60,
  ...extra
})

describe('RCM Store - no dummy data, server is the source of truth', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    if (typeof localStorage !== 'undefined') {
      localStorage.clear()
    }
  })

  it('starts empty (no dummy rows) before anything is fetched', () => {
    const store = useRCMStore()
    expect(store.rcmList).toEqual([])
    expect(store.filteredRCMList).toEqual([])
  })

  it('ignores and clears the old localStorage cache (rcm_items_v2)', () => {
    localStorage.setItem('rcm_items_v2', JSON.stringify([row('cached-1')]))
    const store = useRCMStore()
    expect(store.rcmList).toEqual([])
    expect(localStorage.getItem('rcm_items_v2')).toBeNull()
  })

  it('an empty API response gives an empty list', async () => {
    const store = useRCMStore()
    store.rcmList = [row('old-1')]
    vi.mocked($fetch).mockResolvedValueOnce({ success: true, data: [] })
    await store.fetchRCMList()
    expect(store.rcmList).toEqual([])
  })

  it('an API error gives an empty list and an error message', async () => {
    const store = useRCMStore()
    store.rcmList = [row('old-1')]
    vi.mocked($fetch).mockRejectedValueOnce(new Error('Network error'))
    await store.fetchRCMList()
    expect(store.rcmList).toEqual([])
    expect(store.errorMsg).toBeTruthy()
  })

  it('a failed create does not add a row', async () => {
    const store = useRCMStore()
    store.rcmList = [row('a')]
    vi.mocked($fetch).mockRejectedValueOnce(new Error('500'))
    const { id, total_weighted_score, ...input } = row('new')
    await expect(store.addRCMItem(input)).rejects.toThrow()
    expect(store.rcmList.map(i => i.id)).toEqual(['a'])
  })

  it('a failed update leaves the row unchanged', async () => {
    const store = useRCMStore()
    store.rcmList = [row('a', { control_description: 'Original' })]
    vi.mocked($fetch).mockRejectedValueOnce(new Error('500'))
    await expect(store.updateRCMItem(row('a', { control_description: 'Changed' }))).rejects.toThrow()
    expect(store.rcmList[0].control_description).toBe('Original')
  })

  it('a successful update replaces the row with the server version', async () => {
    const store = useRCMStore()
    store.rcmList = [row('a', { control_description: 'Original' })]
    vi.mocked($fetch).mockResolvedValueOnce({ success: true, data: row('a', { control_description: 'From server', design_effectiveness_rating: 5 }) })
    await store.updateRCMItem(row('a', { control_description: 'Changed' }))
    const call = vi.mocked($fetch).mock.calls.find(c => c[1]?.method === 'PUT')
    expect(call[0]).toContain('/rcm/a')
    expect(call[1]).toMatchObject({ method: 'PUT' })
    expect(call[1].body.id).toBeUndefined()
    expect(store.rcmList[0].control_description).toBe('From server')
    expect(store.rcmList[0].total_weighted_score).toBe(20 + 12 * 4)
  })

  it('a failed delete keeps the row', async () => {
    const store = useRCMStore()
    store.rcmList = [row('a')]
    vi.mocked($fetch).mockRejectedValueOnce(new Error('500'))
    await expect(store.deleteRCMItem('a')).rejects.toThrow()
    expect(store.rcmList.map(i => i.id)).toEqual(['a'])
  })

  it('never writes RCM rows to localStorage', async () => {
    const store = useRCMStore()
    vi.mocked($fetch).mockResolvedValueOnce({ success: true, data: [row('a')] })
    await store.fetchRCMList()
    expect(localStorage.getItem('rcm_items_v2')).toBeNull()
  })
})
