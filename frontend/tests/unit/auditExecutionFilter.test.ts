import { describe, it, expect } from 'vitest'
import { normalizeAuditStatus } from '~/stores/audit-execution'

describe('Audit Execution Status Normalization for 6 Stages', () => {
  it('should normalize all 6 stages correctly', () => {
    expect(normalizeAuditStatus('planning')).toBe('planning')
    expect(normalizeAuditStatus('planned')).toBe('planning')
    expect(normalizeAuditStatus('Planning & Preparation')).toBe('planning')

    expect(normalizeAuditStatus('entry meeting')).toBe('entry meeting')
    expect(normalizeAuditStatus('entry_meeting')).toBe('entry meeting')
    expect(normalizeAuditStatus('Entry Meeting & Scope Alignment')).toBe('entry meeting')

    expect(normalizeAuditStatus('fieldwork')).toBe('fieldwork')
    expect(normalizeAuditStatus('Fieldwork & Control Testing')).toBe('fieldwork')

    expect(normalizeAuditStatus('draft findings')).toBe('draft findings')
    expect(normalizeAuditStatus('draft_findings')).toBe('draft findings')
    expect(normalizeAuditStatus('Draft Findings & Recommendations')).toBe('draft findings')

    expect(normalizeAuditStatus('reporting')).toBe('reporting')
    expect(normalizeAuditStatus('Reporting & Exit Meeting')).toBe('reporting')

    expect(normalizeAuditStatus('completed')).toBe('completed')
    expect(normalizeAuditStatus('Completed')).toBe('completed')
    expect(normalizeAuditStatus('Audit Completed')).toBe('completed')
  })

  it('should map legacy in_progress based on progress percentage', () => {
    expect(normalizeAuditStatus('in_progress', 15)).toBe('entry meeting')
    expect(normalizeAuditStatus('in_progress', 40)).toBe('fieldwork')
    expect(normalizeAuditStatus('in_progress', 60)).toBe('draft findings')
    expect(normalizeAuditStatus('in_progress', 80)).toBe('reporting')
  })

  it('should fallback to progress when status is missing', () => {
    expect(normalizeAuditStatus(undefined, 0)).toBe('planning')
    expect(normalizeAuditStatus(undefined, 20)).toBe('entry meeting')
    expect(normalizeAuditStatus(undefined, 35)).toBe('fieldwork')
    expect(normalizeAuditStatus(undefined, 65)).toBe('draft findings')
    expect(normalizeAuditStatus(undefined, 85)).toBe('reporting')
    expect(normalizeAuditStatus(undefined, 100)).toBe('completed')
  })
})

describe('Audit Execution 6-Stage Filter Logic', () => {
  const normalizeStatus = (val?: string | null): string => {
    if (!val) return ''
    const clean = String(val).toLowerCase().replace(/[\s_-]+/g, '')
    if (clean.includes('completed') || clean.includes('selesai') || clean.includes('done')) return 'completed'
    if (clean.includes('reporting') || clean.includes('report') || clean.includes('pelaporan')) return 'reporting'
    if (clean.includes('draftfinding') || clean.includes('findings') || clean.includes('temuan')) return 'draftfindings'
    if (clean.includes('fieldwork')) return 'fieldwork'
    if (clean.includes('entrymeeting') || clean.includes('entry')) return 'entrymeeting'
    if (clean.includes('planning') || clean.includes('planned') || clean.includes('perencanaan')) return 'planning'
    if (clean === 'inprogress') return 'inprogress'
    if (clean === 'canceled') return 'cancelled'
    return clean
  }

  const isStatusMatch = (auditStatusVal?: string | null, filterStatusVal?: string, progress?: number): boolean => {
    if (!filterStatusVal) return true
    const fNorm = normalizeStatus(filterStatusVal)
    let aNorm = normalizeStatus(auditStatusVal)

    if (!aNorm && typeof progress === 'number') {
      if (progress >= 100) aNorm = 'completed'
      else if (progress >= 76) aNorm = 'reporting'
      else if (progress >= 51) aNorm = 'draftfindings'
      else if (progress >= 26) aNorm = 'fieldwork'
      else if (progress >= 1) aNorm = 'entrymeeting'
      else aNorm = 'planning'
    }

    if (aNorm === fNorm) return true

    if (fNorm === 'inprogress' && (aNorm === 'entrymeeting' || aNorm === 'fieldwork' || aNorm === 'draftfindings' || aNorm === 'reporting')) {
      return true
    }

    if (aNorm === 'inprogress' && typeof progress === 'number') {
      if (progress >= 76 && fNorm === 'reporting') return true
      if (progress >= 51 && progress < 76 && fNorm === 'draftfindings') return true
      if (progress >= 26 && progress < 51 && fNorm === 'fieldwork') return true
      if (progress >= 1 && progress < 26 && fNorm === 'entrymeeting') return true
    }

    return false
  }

  // Items matching the backend API response
  const items = [
    { id: '1', ref: 'ST-003/SKAI/2026', status: 'planning', progress: 0 },
    { id: '2', ref: 'ST-004/SKAI/2026', status: 'entry meeting', progress: 15 },
    { id: '3', ref: 'ST-002/SKAI/2026', status: 'fieldwork', progress: 40 },
    { id: '4', ref: 'ST-005/SKAI/2026', status: 'draft findings', progress: 60 },
    { id: '5', ref: 'ST-001/SKAI/2026', status: 'reporting', progress: 80 },
    { id: '6', ref: 'ST-001/SKAI/2025', status: 'completed', progress: 100 },
    { id: '7', ref: 'ST-002/SKAI/2025', status: 'completed', progress: 100 },
  ]

  it('filters 1. Planning correctly', () => {
    const res = items.filter(i => isStatusMatch(i.status, 'planning', i.progress))
    expect(res.length).toBe(1)
    expect(res[0].id).toBe('1')
  })

  it('filters 2. Entry Meeting correctly', () => {
    const res = items.filter(i => isStatusMatch(i.status, 'entry meeting', i.progress))
    expect(res.length).toBe(1)
    expect(res[0].id).toBe('2')
  })

  it('filters 3. Fieldwork correctly', () => {
    const res = items.filter(i => isStatusMatch(i.status, 'fieldwork', i.progress))
    expect(res.length).toBe(1)
    expect(res[0].id).toBe('3')
  })

  it('filters 4. Draft Findings correctly', () => {
    const res = items.filter(i => isStatusMatch(i.status, 'draft findings', i.progress))
    expect(res.length).toBe(1)
    expect(res[0].id).toBe('4')
  })

  it('filters 5. Reporting correctly', () => {
    const res = items.filter(i => isStatusMatch(i.status, 'reporting', i.progress))
    expect(res.length).toBe(1)
    expect(res[0].id).toBe('5')
  })

  it('filters 6. Completed correctly', () => {
    const res = items.filter(i => isStatusMatch(i.status, 'completed', i.progress))
    expect(res.length).toBe(2)
  })

  it('returns all items when filter is undefined', () => {
    const all = items.filter(i => isStatusMatch(i.status, undefined, i.progress))
    expect(all.length).toBe(7)
  })
})
