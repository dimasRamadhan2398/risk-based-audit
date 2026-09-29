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

describe('RCM Store - Persistence and API handling', () => {
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
