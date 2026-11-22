import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const leaflet = vi.hoisted(() => {
  const layerGroup = {
    clearLayers: vi.fn(),
    addLayer: vi.fn(),
  }
  const mapInstance = {
    setView: vi.fn().mockReturnThis(),
    fitBounds: vi.fn(),
    invalidateSize: vi.fn(),
    remove: vi.fn(),
  }
  function makeMarker() {
    const handlers = {}
    let html = ''
    const popupDiv = document.createElement('div')
    return {
      bindPopup(value) {
        html = value
        return this
      },
      on(event, fn) {
        handlers[event] = fn
        return this
      },
      getPopup() {
        return { getElement: () => popupDiv }
      },
      // Simula a abertura do popup no DOM real (Leaflet injeta o HTML lá).
      openPopup() {
        document.body.appendChild(popupDiv)
        popupDiv.innerHTML = html
        if (handlers.popupopen) handlers.popupopen()
      },
      popupEl: popupDiv,
    }
  }
  return {
    map: vi.fn(() => mapInstance),
    tileLayer: vi.fn(() => ({ addTo: vi.fn() })),
    layerGroup: vi.fn(() => ({
      ...layerGroup,
      addTo: vi.fn().mockReturnValue(layerGroup),
    })),
    divIcon: vi.fn(() => ({})),
    marker: vi.fn(() => makeMarker()),
    Icon: { Default: { mergeOptions: vi.fn() } },
    __map: mapInstance,
    __layer: layerGroup,
  }
})

vi.mock('leaflet', () => ({ default: leaflet }))

import AdminMap from '../../src/components/AdminMap.vue'
import { useAdminStore } from '../../src/stores/admin'

function user(id, lat = -23.5505, lng = -46.6333, isActive = true) {
  return {
    id,
    name: `Usuário ${id}`,
    email: `${id}@radiopx.local`,
    isActive,
    isAdmin: false,
    lastLat: lat,
    lastLng: lng,
    lastLocationAt: '2026-11-21T12:34:56Z',
  }
}

describe('AdminMap', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('cria marcadores e ajusta limites só quando a quantidade muda', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const store = useAdminStore()
    store.users = [user('a'), user('b')]

    const wrapper = mount(AdminMap, { global: { plugins: [pinia] } })
    await flushPromises()

    expect(leaflet.marker).toHaveBeenCalledTimes(2)
    expect(leaflet.__map.fitBounds).toHaveBeenCalledTimes(1)
    expect(wrapper.find('#map-empty').attributes('hidden')).toBeDefined()

    store.users = [user('a'), user('b', -23.6, -46.7)]
    await flushPromises()
    expect(leaflet.marker).toHaveBeenCalledTimes(4)
    expect(leaflet.__map.fitBounds).toHaveBeenCalledTimes(1)

    store.users = [user('a'), user('b'), user('c')]
    await flushPromises()
    expect(leaflet.marker).toHaveBeenCalledTimes(7)
    expect(leaflet.__map.fitBounds).toHaveBeenCalledTimes(2)
  })

  it('expõe invalidate e esconde o mapa quando não há posições', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const store = useAdminStore()
    store.users = [user('a', null, null)]

    const wrapper = mount(AdminMap, { global: { plugins: [pinia] } })
    await flushPromises()

    expect(leaflet.marker).not.toHaveBeenCalled()
    expect(wrapper.find('#map-empty').attributes('hidden')).toBeUndefined()

    wrapper.vm.invalidate()
    expect(leaflet.__map.invalidateSize).toHaveBeenCalled()
  })

  it('preenche o endereço do popup no open: Nominatim ok e falha', async () => {
    vi.useFakeTimers()
    const fetchMock = vi.fn().mockImplementation((url) => {
      if (String(url).includes('lat=-23.5505')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({ display_name: 'Avenida Paulista, São Paulo' }),
        })
      }
      return Promise.reject(new Error('offline'))
    })
    vi.stubGlobal('fetch', fetchMock)

    try {
      const pinia = createPinia()
      setActivePinia(pinia)
      const store = useAdminStore()
      store.users = [user('a'), user('b', -23.6, -46.7)]

      mount(AdminMap, { global: { plugins: [pinia] } })
      await vi.advanceTimersByTimeAsync(0)

      const okMarker = leaflet.marker.mock.results[0].value
      okMarker.openPopup()
      await vi.advanceTimersByTimeAsync(1000)

      const okAddr = okMarker.popupEl.querySelector('.popup-addr')
      expect(okAddr.textContent).toBe('Avenida Paulista, São Paulo')
      expect(okAddr.dataset.pending).toBeUndefined()

      const failMarker = leaflet.marker.mock.results[1].value
      failMarker.openPopup()
      await vi.advanceTimersByTimeAsync(1000)

      const failAddr = failMarker.popupEl.querySelector('.popup-addr')
      expect(failAddr.textContent).toBe('endereço indisponível')
      expect(failAddr.dataset.pending).toBeUndefined()
    } finally {
      vi.unstubAllGlobals()
      vi.useRealTimers()
      document.body.innerHTML = ''
    }
  })
})
