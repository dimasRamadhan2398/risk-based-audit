import { describe, it, expect } from 'vitest'
import { buildAuditableEntityOptions } from '~/utils/auditableEntityOptions'

describe('buildAuditableEntityOptions', () => {
  const corporate = [
    {
      id: 'corp-1',
      name: 'Finance',
      children: [
        { id: 'corp-2', name: 'Treasury', children: [] },
        { id: 'corp-3', name: 'Tax', children: [] }
      ]
    },
    { id: 'corp-4', name: 'IT', children: [] }
  ]

  it('lists yearly entities first with the corporate id and scored values', () => {
    const yearly = [
      {
        id: 'year-row-1',
        corporate_audit_universe_id: 'corp-4',
        corporate_audit_universe: { id: 'corp-4', name: 'IT' },
        risk_level: 'High',
        risk_index: '72.5',
        audit_priority: true
      }
    ]

    const options = buildAuditableEntityOptions(yearly, corporate)

    expect(options[0]).toEqual({
      id: 'corp-4',
      name: 'IT',
      label: 'IT',
      value: 'IT',
      riskLevel: 'High',
      riskIndex: 72.5,
      auditPriority: true,
      source: 'yearly'
    })
  })

  it('adds corporate nodes and sub-entities that are not already listed, with parent in the label', () => {
    const options = buildAuditableEntityOptions([], corporate)

    expect(options.map(o => o.name)).toEqual(['Finance', 'Treasury', 'Tax', 'IT'])
    expect(options.find(o => o.name === 'Treasury')).toMatchObject({
      id: 'corp-2',
      label: 'Treasury (Finance)',
      riskLevel: 'Not Scored',
      riskIndex: 0,
      auditPriority: false,
      source: 'corporate'
    })
  })

  it('de-duplicates names case-insensitively, keeping the yearly entry', () => {
    const yearly = [
      { corporate_audit_universe_id: 'corp-4', corporate_audit_universe: { name: 'it' }, risk_level: 'Low' }
    ]

    const options = buildAuditableEntityOptions(yearly, corporate)

    const itEntries = options.filter(o => o.name.toLowerCase() === 'it')
    expect(itEntries).toHaveLength(1)
    expect(itEntries[0]).toMatchObject({ source: 'yearly', riskLevel: 'Low', id: 'corp-4' })
  })

  it('returns an empty list when both sources are empty or missing', () => {
    expect(buildAuditableEntityOptions([], [])).toEqual([])
    expect(buildAuditableEntityOptions(null, undefined)).toEqual([])
  })

  it('skips yearly rows without a corporate name and defaults missing risk data', () => {
    const yearly = [
      { id: 'x', corporate_audit_universe_id: 'corp-9', corporate_audit_universe: null, risk_level: 'High' },
      { id: 'y', corporate_audit_universe_id: 'corp-5', corporate_audit_universe: { name: 'Ops' } }
    ]

    const options = buildAuditableEntityOptions(yearly, [])

    expect(options).toHaveLength(1)
    expect(options[0]).toMatchObject({ name: 'Ops', riskLevel: 'Not Scored', riskIndex: 0, auditPriority: false, id: 'corp-5' })
  })
})
