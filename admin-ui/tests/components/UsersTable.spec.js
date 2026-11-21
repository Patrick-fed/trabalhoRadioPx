import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import UsersTable from '../../src/components/UsersTable.vue'
import { useAdminStore } from '../../src/stores/admin'

function makeUser(id, name, email, overrides = {}) {
  return {
    id,
    name,
    email,
    isActive: true,
    isAdmin: false,
    lastLat: null,
    lastLng: null,
    lastLocationAt: null,
    ...overrides,
  }
}

describe('UsersTable', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renderiza colunas, filtro local e ações; oculta ações da própria linha', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const store = useAdminStore()
    store.me = { id: 'admin-1', email: 'admin@radiopx.local', isAdmin: true }
    store.users = [
      makeUser('admin-1', 'Admin', 'admin@radiopx.local'),
      makeUser('u2', 'Outro Usuário', 'outro@radiopx.local'),
    ]
    store.usersLoaded = true

    const wrapper = mount(UsersTable, { global: { plugins: [pinia] } })
    await flushPromises()

    expect(wrapper.findAll('th').map((th) => th.text())).toEqual([
      'Nome',
      'E-mail',
      'Status',
      'Papel',
      'Localização',
      'Ações',
    ])
    expect(wrapper.findAll('tbody tr')).toHaveLength(2)
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('você')
    expect(wrapper.findAll('tbody tr')[0].findAll('button')).toHaveLength(0)
    expect(wrapper.findAll('tbody tr')[1].findAll('button').map((b) => b.text())).toEqual([
      'Suspender',
      'Tornar admin',
      'Excluir',
    ])

    await wrapper.find('#user-search').setValue('OUTRO')

    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
    expect(wrapper.find('tbody tr').text()).toContain('outro@radiopx.local')
  })

  it('mostra Carregando… e Nenhum usuário encontrado nos estados vazios', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const store = useAdminStore()
    store.usersLoaded = false

    const wrapper = mount(UsersTable, { global: { plugins: [pinia] } })
    expect(wrapper.find('tbody').text()).toContain('Carregando…')

    store.usersLoaded = true
    await flushPromises()
    expect(wrapper.find('tbody').text()).toContain('Nenhum usuário encontrado')
  })

  it('mostra Carregando… na primeira abertura após o dashboard já ter carregado usuários', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const store = useAdminStore()
    store.status = 'authenticated'
    store.users = [makeUser('admin-1', 'Admin', 'admin@radiopx.local')]
    store.usersLoaded = true

    let resolveFetch
    vi.stubGlobal(
      'fetch',
      vi.fn(
        () =>
          new Promise((resolve) => {
            resolveFetch = resolve
          }),
      ),
    )

    const wrapper = mount(UsersTable, { global: { plugins: [pinia] } })
    expect(wrapper.find('tbody').text()).not.toContain('Carregando…')

    store.setActiveTab('users')
    await nextTick()

    expect(store.usersTabLoading).toBe(true)
    expect(wrapper.find('tbody').text()).toContain('Carregando…')

    resolveFetch({
      status: 200,
      ok: true,
      text: async () =>
        JSON.stringify([makeUser('admin-1', 'Admin', 'admin@radiopx.local')]),
    })
    await flushPromises()

    expect(store.usersTabLoading).toBe(false)
    expect(wrapper.find('tbody').text()).not.toContain('Carregando…')
  })
})
