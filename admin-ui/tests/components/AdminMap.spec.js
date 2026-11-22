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
  return {
    map: vi.fn(() => mapInstance),
    tileLayer: vi.fn(() => ({ addTo: vi.fn() })),
    layerGroup: vi.fn(() => ({
      ...layerGroup,
      addTo: vi.fn().mockReturnValue(layerGroup),
    })),
    divIcon: vi.fn(() => ({})),
    marker: vi.fn(() => ({ bindPopup: vi.fn().mockReturnThis() })),
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
})
