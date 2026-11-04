import { describe, it, expect } from 'vitest'
import {
  caretAfterFormat,
  formatDisplayNumber,
  formatUnitValue,
  formatNumberInput,
  isCurrencyUnit,
  parseNumberInput,
  separatorsFor
} from '~/utils/numberInput'

describe('separatorsFor', () => {
  it('uses "," for en and "." for id; unknown falls back to en', () => {
    expect(separatorsFor('en')).toEqual({ group: ',', decimal: '.' })
    expect(separatorsFor('id')).toEqual({ group: '.', decimal: ',' })
    expect(separatorsFor('fr')).toEqual({ group: ',', decimal: '.' })
  })
})

describe('formatNumberInput', () => {
  it('groups thousands per locale', () => {
    expect(formatNumberInput('1234567', 'en')).toBe('1,234,567')
    expect(formatNumberInput('1234567', 'id')).toBe('1.234.567')
    expect(formatNumberInput('999', 'en')).toBe('999')
    expect(formatNumberInput('1000', 'id')).toBe('1.000')
  })
  it('keeps decimals and intermediate states', () => {
    expect(formatNumberInput('1234567.5', 'en')).toBe('1,234,567.5')
    expect(formatNumberInput('1234567.5', 'id')).toBe('1.234.567,5')
    expect(formatNumberInput('12.', 'id')).toBe('12,')
    expect(formatNumberInput('0.05', 'en')).toBe('0.05')
  })
  it('handles negative, empty and non-numeric', () => {
    expect(formatNumberInput('-1234', 'en')).toBe('-1,234')
    expect(formatNumberInput('-', 'en')).toBe('-')
    expect(formatNumberInput('', 'en')).toBe('')
    expect(formatNumberInput(null, 'en')).toBe('')
    expect(formatNumberInput('n/a', 'en')).toBe('n/a')
  })
})

describe('parseNumberInput', () => {
  it('strips group separators', () => {
    expect(parseNumberInput('1,234,567', 'en')).toBe('1234567')
    expect(parseNumberInput('1.234.567', 'id')).toBe('1234567')
  })
  it('reads the locale decimal separator', () => {
    expect(parseNumberInput('1,234.50', 'en')).toBe('1234.50')
    expect(parseNumberInput('1.234,50', 'id')).toBe('1234.50')
    expect(parseNumberInput('12,', 'id')).toBe('12.')
    expect(parseNumberInput(',5', 'id')).toBe('0.5')
  })
  it('keeps only the first decimal separator and drops junk', () => {
    expect(parseNumberInput('1.2.3', 'en')).toBe('1.23')
    expect(parseNumberInput('Rp 1.000abc', 'id')).toBe('1000')
  })
  it('handles negatives, empty and leading zeros', () => {
    expect(parseNumberInput('-1,234', 'en')).toBe('-1234')
    expect(parseNumberInput('-', 'en')).toBe('-')
    expect(parseNumberInput('-5', 'en', { allowNegative: false })).toBe('5')
    expect(parseNumberInput('', 'en')).toBe('')
    expect(parseNumberInput(undefined, 'id')).toBe('')
    expect(parseNumberInput('abc', 'en')).toBe('')
    expect(parseNumberInput('007', 'en')).toBe('7')
    expect(parseNumberInput('0', 'en')).toBe('0')
  })
  it('round-trips with format', () => {
    for (const locale of ['en', 'id']) {
      for (const raw of ['0', '1234567', '-9876.54', '1000000.']) {
        expect(parseNumberInput(formatNumberInput(raw, locale), locale)).toBe(raw)
      }
    }
  })
})

describe('caretAfterFormat', () => {
  it('keeps the caret after the typed digit when a separator is inserted', () => {
    // "1,23" + typed "4" at the end -> "1,234" -> "1234" -> "1,234"
    expect(caretAfterFormat('1,234', 5, '1,234', 'en')).toBe(5)
    // typed "5" in the middle of "1234": "12534" caret 3 -> "12,534" caret after "5" (index 4)
    expect(caretAfterFormat('12534', 3, '12,534', 'en')).toBe(4)
    expect(caretAfterFormat('1000', 1, '1.000', 'id')).toBe(1)
  })
  it('falls back to the end', () => {
    expect(caretAfterFormat('12', 2, '12', 'en')).toBe(2)
  })
})

describe('formatDisplayNumber', () => {
  it('formats plain numbers and leaves other text alone', () => {
    expect(formatDisplayNumber('1500000', 'id')).toBe('1.500.000')
    expect(formatDisplayNumber('85.5', 'en')).toBe('85.5')
    expect(formatDisplayNumber('n/a', 'en')).toBe('n/a')
  })
})

describe('formatUnitValue', () => {
  it('adds the currency symbol', () => {
    expect(formatUnitValue('1234567', 'Rp', 'id')).toBe('Rp 1.234.567')
    expect(formatUnitValue('1234567', 'Rp', 'en')).toBe('Rp 1,234,567')
    expect(formatUnitValue('1234567.5', 'USD', 'en')).toBe('$1,234,567.5')
    expect(formatUnitValue('1234567.5', 'USD', 'id')).toBe('$1.234.567,5')
    expect(formatUnitValue('2500', 'GBP', 'en')).toBe('£2,500')
    expect(formatUnitValue('-2500', 'GBP', 'en')).toBe('-£2,500')
  })
  it('appends other units, and shows Amount/no unit as a bare number', () => {
    expect(formatUnitValue('85', '%', 'en')).toBe('85 %')
    expect(formatUnitValue('1200', 'Hour', 'id')).toBe('1.200 Hour')
    expect(formatUnitValue('1200', 'Amount', 'en')).toBe('1,200')
    expect(formatUnitValue('1200', '', 'en')).toBe('1,200')
    expect(formatUnitValue('87.5', 'Score', 'en')).toBe('87.5')
    expect(formatUnitValue('1200', 'Score', 'id')).toBe('1.200')
  })
  it('shows "-" for empty values and leaves non-numeric text', () => {
    expect(formatUnitValue('', 'Rp', 'en')).toBe('-')
    expect(formatUnitValue(undefined, 'Rp', 'en')).toBe('-')
    expect(formatUnitValue('-', 'Rp', 'en')).toBe('-')
    expect(formatUnitValue('n/a', 'Rp', 'en')).toBe('n/a')
  })
})

describe('isCurrencyUnit', () => {
  it('recognises the three currencies only', () => {
    expect(['Rp', 'USD', 'GBP'].every(isCurrencyUnit)).toBe(true)
    expect(isCurrencyUnit('%')).toBe(false)
    expect(isCurrencyUnit(undefined)).toBe(false)
  })
})
