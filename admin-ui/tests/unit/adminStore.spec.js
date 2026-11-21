import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAdminStore } from '../../src/stores/admin'

function response(status, body) {
  return {
    status,
    ok: status >= 200 && status < 300,
    text: async () => (body === undefined ? '' : JSON.stringify(body)),
  }
}

const adminUser = {
  id: 'admin-1',
  name: 'Admin',
  email: 'admin@radiopx.local',
  is_active: true,
  is_admin: true,
  last_lat: null,
  last_lng: null,
  last_location_at: null,
}

describe('admin store', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    setActivePinia(createPinia())
  })

  it('faz login admin e mantém o token apenas em memória', async () => {
    const setItem = vi.spyOn(Storage.prototype, 'setItem')
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response(200, { token: 'jwt' }))
        .mockResolvedValueOnce(response(200, [adminUser])),
    )
    const store = useAdminStore()

    await store.login('  ADMIN@radiopx.local  ', 'senha')

    expect(store.status).toBe('authenticated')
    expect(store.token).toBe('jwt')
    expect(store.me.isAdmin).toBe(true)
    expect(store.me.email).toBe('admin@radiopx.local')
    expect(setItem).not.toHaveBeenCalled()
  })

  it('descarta o token e nega usuário sem is_admin', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response(200, { token: 'jwt' }))
        .mockResolvedValueOnce(
          response(200, [{ ...adminUser, is_admin: false }]),
        ),
    )
    const store = useAdminStore()

    await expect(store.login('admin@radiopx.local', 'senha')).rejects.toThrow(
      'Acesso negado: apenas administradores',
    )
    expect(store.status).toBe('anonymous')
    expect(store.token).toBeNull()
    expect(store.me).toBeNull()
  })

  it('limpa todo o estado em 401', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response(200, { token: 'jwt' }))
        .mockResolvedValueOnce(response(401, { error: 'unauthorized' })),
    )
    const store = useAdminStore()

    await expect(store.login('admin@radiopx.local', 'senha')).rejects.toThrow(
      'Sessão expirada. Entre novamente.',
    )

    expect(store.status).toBe('anonymous')
    expect(store.token).toBeNull()
    expect(store.users).toEqual([])
    expect(store.stats).toBeNull()
    expect(store.activeTab).toBe('dashboard')
  })

  it('logout limpa cache, aba, toast e timer', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response(200, { token: 'jwt' }))
        .mockResolvedValueOnce(response(200, [adminUser])),
    )
    const store = useAdminStore()
    await store.login('admin@radiopx.local', 'senha')
    store.users.push({ ...adminUser, id: 'outro' })
    store.activeTab = 'users'

    store.logout()

    expect(store.status).toBe('anonymous')
    expect(store.users).toEqual([])
    expect(store.activeTab).toBe('dashboard')
    expect(store.toast.visible).toBe(false)
  })
})
