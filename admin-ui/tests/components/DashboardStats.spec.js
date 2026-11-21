import { describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import DashboardStats from '../../src/components/DashboardStats.vue'
import { useAdminStore } from '../../src/stores/admin'

const stats = {
  usersTotal: 12,
  usersActive: 10,
  channelsTotal: 4,
  wsConnections: 8,
  liveRooms: 3,
  transmittingChannels: 1,
  locatedUsers: 5,
}

async function mountDashboard(previous = null, current = stats) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const store = useAdminStore()
  store.stats = { ...current }
  store.prevStats = previous
  const wrapper = mount(DashboardStats, { global: { plugins: [pinia] } })
  await flushPromises()
  return wrapper
}

describe('DashboardStats', () => {
  it('renderiza os sete cards, subtítulos e o hint exato', async () => {
    const wrapper = await mountDashboard()

    expect(wrapper.findAll('.card')).toHaveLength(7)
    expect(wrapper.findAll('.card-label').map((el) => el.text())).toEqual([
      'Usuários',
      'Ativos',
      'Canais',
      'Conexões WS',
      'Salas ativas',
      'Transmitindo',
      'Localização',
    ])
    expect(wrapper.findAll('.card-sub').map((el) => el.text())).toEqual([
      '2 suspensos',
      '83% do total',
      'canais criados',
      'clientes conectados',
      'salas com membros',
      'canais no ar',
      'com posição conhecida',
    ])
    expect(wrapper.find('.hint').text()).toBe(
      'Atualização automática a cada 10s. Variação desde a leitura anterior.',
    )
  })

  it('mostra deltas apenas para users, ativos, canais e localização', async () => {
    const wrapper = await mountDashboard({
      usersTotal: 10,
      usersActive: 8,
      channelsTotal: 2,
      locatedUsers: 3,
    })

    const deltas = wrapper.findAll('.delta')
    expect(deltas).toHaveLength(4)
    expect(deltas.map((el) => el.text())).toEqual(['+2', '+2', '+2', '+2'])
  })

  it('usa os textos de estado vazio quando não há usuários', async () => {
    const wrapper = await mountDashboard(null, {
      ...stats,
      usersTotal: 0,
      usersActive: 0,
      channelsTotal: 0,
      wsConnections: 0,
      liveRooms: 0,
      transmittingChannels: 0,
      locatedUsers: 0,
    })

    expect(wrapper.findAll('.card-sub').map((el) => el.text())).toEqual([
      'todos ativos',
      'sem usuários',
      'nenhum canal',
      'nenhuma conexão',
      'nenhuma sala',
      'silêncio no ar',
      'com posição conhecida',
    ])
  })
})
