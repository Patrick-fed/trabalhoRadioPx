<script setup>
import { computed, ref } from 'vue'
import { useAdminStore } from '../stores/admin'
import { formatCoord, formatWhen, osmUrl } from '../utils/format'

const store = useAdminStore()
const search = ref('')

const filteredUsers = computed(() => {
  const query = search.value.toLowerCase().trim()
  if (!query) return store.users
  return store.users.filter(
    (user) =>
      (user.name || '').toLowerCase().includes(query) ||
      (user.email || '').toLowerCase().includes(query),
  )
})
</script>

<template>
  <section class="tab">
    <div class="toolbar">
      <input id="user-search" v-model="search" type="search" placeholder="Buscar por nome ou e-mail…" />
      <button id="refresh-users" class="btn-secondary" type="button" @click="store.refreshUsers()">
        Atualizar
      </button>
    </div>

    <div class="table-wrap">
      <table id="users-table">
        <thead>
          <tr>
            <th>Nome</th>
            <th>E-mail</th>
            <th>Status</th>
            <th>Papel</th>
            <th>Localização</th>
            <th>Ações</th>
          </tr>
        </thead>
        <tbody id="users-tbody">
          <tr v-if="!store.usersLoaded || store.usersTabLoading">
            <td colspan="6" class="empty">Carregando…</td>
          </tr>
          <tr v-else-if="!filteredUsers.length">
            <td colspan="6" class="empty">Nenhum usuário encontrado</td>
          </tr>
          <template v-else>
            <tr v-for="user in filteredUsers" :key="user.id">
            <td>{{ user.name || '—' }}</td>
            <td>{{ user.email || '—' }}</td>
            <td>
              <span class="badge" :class="user.isActive ? 'badge-ok' : 'badge-off'">
                {{ user.isActive ? 'ativo' : 'suspenso' }}
              </span>
            </td>
            <td>
              <span class="badge" :class="user.isAdmin ? 'badge-admin' : 'badge-user'">
                {{ user.isAdmin ? 'admin' : 'usuário' }}
              </span>
            </td>
            <td>
              <span v-if="user.lastLat == null || user.lastLng == null" class="loc-cell">—</span>
              <span v-else class="loc-cell">
                {{ formatCoord(user.lastLat) }}, {{ formatCoord(user.lastLng) }}
                <a :href="osmUrl(user.lastLat, user.lastLng)" target="_blank" rel="noopener">
                  ver no mapa
                </a>
                <br />
                <small>{{ formatWhen(user.lastLocationAt) }}</small>
              </span>
            </td>
            <td>
              <div v-if="store.me && user.id === store.me.id" class="actions">
                <span class="loc-cell">você</span>
              </div>
              <div v-else class="actions">
                <button
                  v-if="user.isActive"
                  class="btn-tiny warn"
                  type="button"
                  @click="store.userAction('suspend', user.id)"
                >
                  Suspender
                </button>
                <button
                  v-else
                  class="btn-tiny"
                  type="button"
                  @click="store.userAction('reactivate', user.id)"
                >
                  Reativar
                </button>
                <button
                  v-if="user.isAdmin"
                  class="btn-tiny"
                  type="button"
                  @click="store.userAction('demote', user.id)"
                >
                  Remover admin
                </button>
                <button
                  v-else
                  class="btn-tiny"
                  type="button"
                  @click="store.userAction('promote', user.id)"
                >
                  Tornar admin
                </button>
                <button
                  class="btn-tiny danger"
                  type="button"
                  @click="store.userAction('delete', user.id)"
                >
                  Excluir
                </button>
              </div>
            </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <p v-if="store.usersError" id="users-error" class="error">{{ store.usersError }}</p>
  </section>
</template>
