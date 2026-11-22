<script setup>
import { computed } from 'vue'
import { useAdminStore } from '../stores/admin'
import { formatDelta } from '../utils/format'

const store = useAdminStore()

const current = computed(
  () =>
    store.stats || {
      usersTotal: 0,
      usersActive: 0,
      channelsTotal: 0,
      wsConnections: 0,
      liveRooms: 0,
      transmittingChannels: 0,
      locatedUsers: 0,
    },
)

const subs = computed(() => {
  const suspended = current.value.usersTotal - current.value.usersActive
  return {
    users: suspended > 0 ? `${suspended} suspenso${suspended > 1 ? 's' : ''}` : 'todos ativos',
    active:
      current.value.usersTotal > 0
        ? `${Math.round((current.value.usersActive / current.value.usersTotal) * 100)}% do total`
        : 'sem usuários',
    channels: current.value.channelsTotal > 0 ? 'canais criados' : 'nenhum canal',
    ws: current.value.wsConnections > 0 ? 'clientes conectados' : 'nenhuma conexão',
    rooms: current.value.liveRooms > 0 ? 'salas com membros' : 'nenhuma sala',
    tx: current.value.transmittingChannels > 0 ? 'canais no ar' : 'silêncio no ar',
    loc: 'com posição conhecida',
  }
})

const deltas = computed(() => ({
  usersTotal: formatDelta(current.value.usersTotal, store.prevStats?.usersTotal),
  usersActive: formatDelta(current.value.usersActive, store.prevStats?.usersActive),
  channelsTotal: formatDelta(current.value.channelsTotal, store.prevStats?.channelsTotal),
  locatedUsers: formatDelta(current.value.locatedUsers, store.prevStats?.locatedUsers),
}))
</script>

<template>
  <div class="dashboard">
    <div id="stats-cards" class="cards">
      <div class="card">
        <span class="card-label">Usuários</span>
        <span id="stat-users" class="card-value">{{ current.usersTotal }}</span>
        <span id="stat-users-sub" class="card-sub">{{ subs.users }}</span>
        <span
          v-if="deltas.usersTotal"
          id="stat-users-delta"
          class="delta"
          :class="deltas.usersTotal.kind"
        >
          {{ deltas.usersTotal.text }}
        </span>
      </div>

      <div class="card">
        <span class="card-label">Ativos</span>
        <span id="stat-active" class="card-value">{{ current.usersActive }}</span>
        <span id="stat-active-sub" class="card-sub">{{ subs.active }}</span>
        <span
          v-if="deltas.usersActive"
          id="stat-active-delta"
          class="delta"
          :class="deltas.usersActive.kind"
        >
          {{ deltas.usersActive.text }}
        </span>
      </div>

      <div class="card">
        <span class="card-label">Canais</span>
        <span id="stat-channels" class="card-value">{{ current.channelsTotal }}</span>
        <span id="stat-channels-sub" class="card-sub">{{ subs.channels }}</span>
        <span
          v-if="deltas.channelsTotal"
          id="stat-channels-delta"
          class="delta"
          :class="deltas.channelsTotal.kind"
        >
          {{ deltas.channelsTotal.text }}
        </span>
      </div>

      <div class="card">
        <span class="card-label">Conexões WS</span>
        <span id="stat-ws" class="card-value">{{ current.wsConnections }}</span>
        <span id="stat-ws-sub" class="card-sub">{{ subs.ws }}</span>
      </div>

      <div class="card">
        <span class="card-label">Salas ativas</span>
        <span id="stat-rooms" class="card-value">{{ current.liveRooms }}</span>
        <span id="stat-rooms-sub" class="card-sub">{{ subs.rooms }}</span>
      </div>

      <div class="card">
        <span class="card-label">Transmitindo</span>
        <span id="stat-tx" class="card-value">{{ current.transmittingChannels }}</span>
        <span id="stat-tx-sub" class="card-sub">{{ subs.tx }}</span>
      </div>

      <div class="card card-loc">
        <span class="card-label">Localização</span>
        <span id="stat-loc" class="card-value">{{ current.locatedUsers }}</span>
        <span id="stat-loc-sub" class="card-sub">{{ subs.loc }}</span>
        <span
          v-if="deltas.locatedUsers"
          id="stat-loc-delta"
          class="delta"
          :class="deltas.locatedUsers.kind"
        >
          {{ deltas.locatedUsers.text }}
        </span>
      </div>
    </div>

    <p class="hint">Atualização automática a cada 10s. Variação desde a leitura anterior.</p>
  </div>
</template>
