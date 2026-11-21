import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('leaflet', () => ({
  default: {
    map: vi.fn(() => ({ setView: vi.fn().mockReturnThis(), fitBounds: vi.fn(), invalidateSize: vi.fn(), remove: vi.fn() })),
    tileLayer: vi.fn(() => ({ addTo: vi.fn() })),
    layerGroup: vi.fn(() => ({ clearLayers: vi.fn(), addLayer: vi.fn(), addTo: vi.fn().mockReturnThis() })),
    divIcon: vi.fn(() => ({})),
    marker: vi.fn(() => ({ bindPopup: vi.fn().mockReturnThis() })),
    Icon: { Default: { mergeOptions: vi.fn() } },
  },
}))

vi.mock('../../src/services/adminApi', () => ({
  adminApi: {
    listUsers: vi.fn().mockResolvedValue([]),
    getStats: vi.fn().mockResolvedValue({
      usersTotal: 0,
      usersActive: 0,
      channelsTotal: 0,
      wsConnections: 0,
      liveRooms: 0,
      transmittingChannels: 0,
      locatedUsers: 0,
    }),
    patchUser: vi.fn(),
    deleteUser: vi.fn(),
  },
}))

import router from '../../src/router'
import AdminShell from '../../src/views/AdminShell.vue'
import { useAdminStore } from '../../src/stores/admin'

describe('AdminShell', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('troca de aba sem alterar /admin/ nem criar histórico', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const store = useAdminStore()
    store.status = 'authenticated'
    store.me = { id: 'me', email: 'admin@radiopx.local', isAdmin: true }

    await router.replace('/')
    const urlBefore = router.currentRoute.value.fullPath
    const push = vi.spyOn(router, 'push')
    const replace = vi.spyOn(router, 'replace')

    const wrapper = mount(AdminShell, { global: { plugins: [pinia, router] } })
    await flushPromises()

    expect(router.options.history.base).toBe('/admin')
    await wrapper.findAll('button.nav-btn')[1].trigger('click')
    await flushPromises()

    expect(store.activeTab).toBe('users')
    expect(router.currentRoute.value.fullPath).toBe(urlBefore)
    expect(push).not.toHaveBeenCalled()
    expect(replace).not.toHaveBeenCalled()

    wrapper.unmount()
  })
})
