import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// Cache e fila 1 req/s são estado de módulo: cada teste importa uma instância
// fresca (resetModules) e usa coordenadas próprias, para nenhum teste herdar
// o estado do anterior. Com estado zerado a 1ª chamada nunca espera a fila.
describe('reverseGeocode', () => {
  let fetchMock
  let reverseGeocode

  beforeEach(async () => {
    vi.useFakeTimers()
    vi.resetModules()
    ;({ reverseGeocode } = await import('../../src/utils/geocode'))
    fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ display_name: 'Rua das Flores, São Paulo' }),
    })
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.useRealTimers()
  })

  it('retorna display_name em pt-BR com a URL do Nominatim', async () => {
    const p = reverseGeocode(-23.5505, -46.6333)
    await vi.advanceTimersByTimeAsync(0)

    await expect(p).resolves.toBe('Rua das Flores, São Paulo')
    const [url, options] = fetchMock.mock.calls[0]
    expect(url).toContain('https://nominatim.openstreetmap.org/reverse?format=jsonv2')
    expect(url).toContain('lat=-23.5505')
    expect(url).toContain('lon=-46.6333')
    expect(url).toContain('accept-language=pt-BR')
    expect(url).toContain('zoom=18')
    expect(options).toEqual({ headers: { Accept: 'application/json' } })
  })

  it('cacheia por coordenada (4 casas): segunda chamada não refaz fetch', async () => {
    const p1 = reverseGeocode(-10.1234, -20.54321)
    await vi.advanceTimersByTimeAsync(0)
    await p1

    const p2 = reverseGeocode(-10.12341, -20.5432) // mesma chave em 4dp
    await vi.advanceTimersByTimeAsync(0)

    await expect(p2).resolves.toBe('Rua das Flores, São Paulo')
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('resolve null em HTTP não-OK e cacheia a falha', async () => {
    fetchMock.mockResolvedValue({ ok: false, status: 503 })

    const p1 = reverseGeocode(1.1111, 2.2222)
    await vi.advanceTimersByTimeAsync(0)
    await expect(p1).resolves.toBeNull()

    const p2 = reverseGeocode(1.1111, 2.2222)
    await vi.advanceTimersByTimeAsync(0)
    await expect(p2).resolves.toBeNull()
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('resolve null em erro de rede (popups nunca rejeitam)', async () => {
    fetchMock.mockRejectedValue(new Error('offline'))

    const p = reverseGeocode(3.3333, 4.4444)
    await vi.advanceTimersByTimeAsync(0)

    await expect(p).resolves.toBeNull()
  })

  it('respeta o intervalo de 1 req/s entre coordenadas diferentes', async () => {
    const p1 = reverseGeocode(5.5555, 6.6666)
    await vi.advanceTimersByTimeAsync(0)
    await p1
    expect(fetchMock).toHaveBeenCalledTimes(1)

    const p2 = reverseGeocode(7.7777, 8.8888)
    await vi.advanceTimersByTimeAsync(0)
    expect(fetchMock).toHaveBeenCalledTimes(1) // ainda aguardando a fila

    await vi.advanceTimersByTimeAsync(1000)
    await p2
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
})
