import { describe, expect, it } from 'vitest'
import { formatCoord, formatDelta, formatWhen, osmUrl } from '../../src/utils/format'

describe('format', () => {
  it('formata data/hora em pt-BR', () => {
    const iso = '2026-11-21T12:34:56Z'
    const expected = new Date(iso).toLocaleString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      year: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })

    expect(formatWhen(iso)).toBe(expected)
    expect(formatWhen(null)).toBe('—')
  })

  it('formata coordenadas com cinco casas', () => {
    expect(formatCoord(-23.5505)).toBe('-23.55050')
    expect(formatCoord(-46.6333)).toBe('-46.63330')
  })

  it('monta a URL do OpenStreetMap com as coordenadas formatadas', () => {
    expect(osmUrl(-23.5505, -46.6333)).toBe(
      'https://www.openstreetmap.org/?mlat=-23.55050&mlon=-46.63330#map=15/-23.55050/-46.63330',
    )
  })

  it('calcula deltas com sinal e classe', () => {
    expect(formatDelta(12, 10)).toEqual({ text: '+2', kind: 'up' })
    expect(formatDelta(8, 10)).toEqual({ text: '−2', kind: 'down' })
    expect(formatDelta(10, 10)).toBeNull()
    expect(formatDelta(10, null)).toBeNull()
  })
})
