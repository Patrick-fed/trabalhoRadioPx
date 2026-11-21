import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAdminStore } from '../../src/stores/admin'

const adminApi = vi.hoisted(() => ({
  login: vi.fn(),
  listUsers: vi.fn(),
  getStats: vi.fn(),
  patchUser: vi.fn(),
  deleteUser: vi.fn(),
}))

vi.mock('../../src/services/adminApi', () => ({ adminApi }))

const target = {
  id: 'u2',
  name: 'Outro',
  email: 'outro@radiopx.local',
  isActive: true,
  isAdmin: false,
  lastLat: null,
  lastLng: null,
  lastLocationAt: null,
}

function makeStore() {
  const pinia = createPinia()
  setActivePinia(pinia)
  const store = useAdminStore()
  store.me = { id: 'admin-1', email: 'admin@radiopx.local', isAdmin: true }
  store.users = [{ ...target }]
  store.status = 'authenticated'
  adminApi.listUsers.mockResolvedValue([{ ...target }])
  adminApi.getStats.mockResolvedValue({
    users_total: 1,
    users_active: 1,
    channels_total: 0,
    ws_connections: 0,
    live_rooms: 0,
    transmitting_channels: 0,
  })
  return store
}

describe('ações administrativas', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.stubGlobal('confirm', vi.fn(() => true))
  })

  it('suspende com confirmação, corpo, toast e recarga', async () => {
    const store = makeStore()
    await store.userAction('suspend', 'u2')

    expect(confirm).toHaveBeenCalledWith('Suspender outro@radiopx.local? Ele não poderá mais entrar.')
    expect(adminApi.patchUser).toHaveBeenCalledWith('u2', { is_active: false })
    expect(store.toast.text).toBe('Usuário suspenso')
    expect(adminApi.listUsers).toHaveBeenCalled()
    expect(adminApi.getStats).toHaveBeenCalledOnce()
  })

  it('reativa sem confirmação', async () => {
    const store = makeStore()
    store.users = [{ ...target, isActive: false }]
    await store.userAction('reactivate', 'u2')

    expect(confirm).not.toHaveBeenCalled()
    expect(adminApi.patchUser).toHaveBeenCalledWith('u2', { is_active: true })
    expect(store.toast.text).toBe('Usuário reativado')
  })

  it('promove e remove admin com as mensagens exatas', async () => {
    const store = makeStore()
    await store.userAction('promote', 'u2')
    expect(confirm).toHaveBeenCalledWith('Promover outro@radiopx.local a administrador?')
    expect(adminApi.patchUser).toHaveBeenCalledWith('u2', { is_admin: true })
    expect(store.toast.text).toBe('Usuário promovido a admin')

    store.users = [{ ...target, isAdmin: true }]
    await store.userAction('demote', 'u2')
    expect(confirm).toHaveBeenCalledWith('Remover privilégios de admin de outro@radiopx.local?')
    expect(adminApi.patchUser).toHaveBeenCalledWith('u2', { is_admin: false })
    expect(store.toast.text).toBe('Privilégios de admin removidos')
  })

  it('exclui com confirmação e corpo DELETE', async () => {
    const store = makeStore()
    await store.userAction('delete', 'u2')

    expect(confirm).toHaveBeenCalledWith(
      'Excluir definitivamente outro@radiopx.local? Essa ação não pode ser desfeita.',
    )
    expect(adminApi.deleteUser).toHaveBeenCalledWith('u2')
    expect(store.toast.text).toBe('Usuário excluído')
  })

  it('não executa ação cancelada ou sobre si mesmo', async () => {
    vi.stubGlobal('confirm', vi.fn(() => false))
    const store = makeStore()
    await store.userAction('suspend', 'u2')
    await store.userAction('delete', 'admin-1')

    expect(adminApi.patchUser).not.toHaveBeenCalled()
    expect(adminApi.deleteUser).not.toHaveBeenCalled()
  })

  it('mostra mensagem de erro sem recarregar em falha', async () => {
    adminApi.patchUser.mockRejectedValueOnce(new Error('Erro 403'))
    const store = makeStore()
    await store.userAction('suspend', 'u2')

    expect(store.toast.kind).toBe('err')
    expect(store.toast.text).toBe('Erro 403')
    expect(adminApi.listUsers).not.toHaveBeenCalled()
  })
})
