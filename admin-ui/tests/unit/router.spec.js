import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import router from '../../src/router'
import { useAdminStore } from '../../src/stores/admin'

vi.mock('leaflet', () => ({
  default: {
    map: vi.fn(),
    tileLayer: vi.fn(),
    layerGroup: vi.fn(),
    divIcon: vi.fn(),
    marker: vi.fn(),
    Icon: { Default: { mergeOptions: vi.fn() } },
  },
}))

describe('router', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('registra apenas login e casca autenticada', () => {
    expect(router.getRoutes().map((route) => route.name).sort()).toEqual(['admin', 'login'])
  })

  it('redireciona para login quando não há sessão', async () => {
    await router.replace('/')

    expect(router.currentRoute.value.name).toBe('login')
  })

  it('abre a casca autenticada com sessão em memória', async () => {
    const store = useAdminStore()
    store.status = 'authenticated'

    await router.replace('/')

    expect(router.currentRoute.value.name).toBe('admin')
  })
})
