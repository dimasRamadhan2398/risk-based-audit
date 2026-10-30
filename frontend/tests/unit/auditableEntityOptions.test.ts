import { describe, it, expect } from 'vitest'
import {
  buildAuditableEntityOptions,
  buildCorporateUniverseNameMap,
  findAuditableEntityForActivity,
  resolveAuditableEntitySelection
} from '~/utils/auditableEntityOptions'

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
      value: 'corp-4',
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

  it('de-duplicates by corporate id, keeping the yearly entry', () => {
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

  describe('entities sharing a name', () => {
    const dupCorporate = [
      { id: 'jkt', name: 'Jakarta Branch', children: [{ id: 'ops-jkt', name: 'Operations', children: [] }] },
      { id: 'sby', name: 'Surabaya Branch', children: [{ id: 'ops-sby', name: 'Operations', children: [] }] }
    ]

    it('keeps both entities, uses the corporate id as the value and the parent name in the label', () => {
      const yearly = [
        { id: 'y1', corporate_audit_universe_id: 'ops-jkt', corporate_audit_universe: { id: 'ops-jkt', name: 'Operations', parent_id: 'jkt' }, risk_level: 'High', risk_index: 85, audit_priority: true, scored_at: '2026-10-01T00:00:00Z' },
        { id: 'y2', corporate_audit_universe_id: 'ops-sby', corporate_audit_universe: { id: 'ops-sby', name: 'Operations', parent_id: 'sby' }, risk_level: 'Low', risk_index: 10, audit_priority: false, scored_at: '2026-10-01T00:00:00Z' }
      ]

      const options = buildAuditableEntityOptions(yearly, dupCorporate)
      const ops = options.filter(o => o.name === 'Operations')

      expect(ops).toHaveLength(2)
      expect(ops.map(o => o.value)).toEqual(['ops-jkt', 'ops-sby'])
      expect(ops.map(o => o.label)).toEqual(['Operations (Jakarta Branch)', 'Operations (Surabaya Branch)'])
      expect(new Set(options.map(o => o.value)).size).toBe(options.length)
    })

    it('resolves the picked option by id, not by name', () => {
      const yearly = [
        { corporate_audit_universe_id: 'ops-jkt', corporate_audit_universe: { name: 'Operations' }, risk_level: 'High', risk_index: 85, audit_priority: true },
        { corporate_audit_universe_id: 'ops-sby', corporate_audit_universe: { name: 'Operations' }, risk_level: 'Low', risk_index: 10, audit_priority: false }
      ]
      const options = buildAuditableEntityOptions(yearly, dupCorporate)
      const surabaya = options.find(o => o.id === 'ops-sby')!

      // option object as emitted by USelectMenu
      expect(resolveAuditableEntitySelection(options, surabaya)).toMatchObject({ id: 'ops-sby', riskLevel: 'Low' })
      // a copy of the object (different reference) still resolves by id
      expect(resolveAuditableEntitySelection(options, { ...surabaya })).toBe(surabaya)
      // a bare value
      expect(resolveAuditableEntitySelection(options, 'ops-sby')).toBe(surabaya)
      // a bare name never matches
      expect(resolveAuditableEntitySelection(options, 'Operations')).toBeNull()
      expect(resolveAuditableEntitySelection(options, null)).toBeNull()
    })
  })

  describe('pre-selecting a saved PKAT row', () => {
    const options = buildAuditableEntityOptions([], [
      { id: 'jkt', name: 'Jakarta Branch', children: [{ id: 'ops-jkt', name: 'Operations', children: [] }] },
      { id: 'sby', name: 'Surabaya Branch', children: [{ id: 'ops-sby', name: 'Operations', children: [] }] },
      { id: 'it', name: 'IT', children: [] }
    ])

    it('selects by auditUniverseId even when the name is shared', () => {
      const selected = findAuditableEntityForActivity(options, { auditUniverseId: 'ops-sby', riskName: 'Operations', riskLevel: 'Low' })
      expect(selected?.id).toBe('ops-sby')
      expect(selected?.label).toBe('Operations (Surabaya Branch)')
    })

    it('prefers the id over a different saved name (renamed entity)', () => {
      expect(findAuditableEntityForActivity(options, { auditUniverseId: 'it', riskName: 'Information Technology' })?.id).toBe('it')
    })

    it('falls back to the name only when the row has no id, and only for an unambiguous name', () => {
      expect(findAuditableEntityForActivity(options, { auditUniverseId: '', riskName: 'it' })?.id).toBe('it')
      const ambiguous = findAuditableEntityForActivity(options, { auditUniverseId: '', riskName: 'Operations', riskLevel: 'High' })
      expect(ambiguous?.id).toBeUndefined()
      expect(ambiguous).toMatchObject({ name: 'Operations', riskLevel: 'High' })
    })

    it('keeps showing a saved entity that is no longer in the lists, and nothing for an empty row', () => {
      expect(findAuditableEntityForActivity(options, { auditUniverseId: 'gone', riskName: 'Old Unit', riskLevel: 'Medium' }))
        .toMatchObject({ id: 'gone', name: 'Old Unit', value: 'gone', riskLevel: 'Medium', auditPriority: false })
      expect(findAuditableEntityForActivity(options, { auditUniverseId: '', riskName: '' })).toBeUndefined()
    })
  })

  describe('unscored entities', () => {
    it('are never presented as prioritized and show as Not Scored', () => {
      const yearly = [
        { corporate_audit_universe_id: 'corp-4', corporate_audit_universe: { name: 'IT' }, scored_at: null, risk_level: '', risk_index: 0, audit_priority: true }
      ]
      const [it0] = buildAuditableEntityOptions(yearly, [])
      expect(it0).toMatchObject({ auditPriority: false, riskLevel: 'Not Scored', riskIndex: 0 })
    })

    it('treats rows from an older backend without scored_at as scored', () => {
      const yearly = [
        { corporate_audit_universe_id: 'corp-4', corporate_audit_universe: { name: 'IT' }, risk_level: 'High', risk_index: 81, audit_priority: true }
      ]
      expect(buildAuditableEntityOptions(yearly, [])[0]).toMatchObject({ auditPriority: true, riskLevel: 'High', riskIndex: 81 })
    })
  })
})

describe('buildCorporateUniverseNameMap', () => {
  it('maps every node of the tree by lower-case id and skips nodes without id or name', () => {
    const names = buildCorporateUniverseNameMap([
      { id: 'Corp-1', name: 'Finance', children: [{ id: 'corp-2', name: 'Treasury', children: [{ id: 'corp-4', name: 'Cash' }] }, { name: 'No id' }] },
      { id: 'corp-3', children: [] }
    ])
    expect(names.get('corp-1')).toBe('Finance')
    expect(names.get('corp-2')).toBe('Treasury')
    expect(names.get('corp-4')).toBe('Cash')
    expect(names.has('corp-3')).toBe(false)
    expect(names.size).toBe(3)
  })

  it('returns an empty map for no data', () => {
    expect(buildCorporateUniverseNameMap(null).size).toBe(0)
    expect(buildCorporateUniverseNameMap(undefined).size).toBe(0)
  })
})
