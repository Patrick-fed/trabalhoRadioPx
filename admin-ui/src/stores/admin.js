import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { adminApi } from '../services/adminApi'
import { setApiToken, setUnauthorizedHandler } from '../services/api'

function normalizeUser(user) {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    isActive: user.is_active,
    isAdmin: user.is_admin,
    lastLat: user.last_lat ?? null,
    lastLng: user.last_lng ?? null,
    lastLocationAt: user.last_location_at ?? null,
  }
}

function normalizeStats(stats, users) {
  return {
    usersTotal: stats.users_total,
    usersActive: stats.users_active,
    channelsTotal: stats.channels_total,
    wsConnections: stats.ws_connections,
    liveRooms: stats.live_rooms,
    transmittingChannels: stats.transmitting_channels,
    locatedUsers: users.filter((user) => user.lastLat != null && user.lastLng != null).length,
  }
}

function previousSnapshot(stats) {
  if (!stats) return null
  return {
    usersTotal: stats.usersTotal,
    usersActive: stats.usersActive,
    channelsTotal: stats.channelsTotal,
    locatedUsers: stats.locatedUsers,
  }
}

export const useAdminStore = defineStore('admin', () => {
  const token = ref(null)
  const me = ref(null)
  const status = ref('anonymous')
  const users = ref([])
  const usersLoaded = ref(false)
  const usersTabLoading = ref(false)
  const usersError = ref('')
  const stats = ref(null)
  const prevStats = ref(null)
  const activeTab = ref('dashboard')
  const toast = ref({ text: '', kind: 'ok', visible: false })
  const isAdmin = computed(() => Boolean(me.value?.isAdmin))

  let statsTimer = null
  let toastTimer = null
  let refreshingUsers = false
  let refreshingDashboard = false
  let usersTabLoaded = false

  function clearToast() {
    if (toastTimer) clearTimeout(toastTimer)
    toastTimer = null
    toast.value = { text: '', kind: 'ok', visible: false }
  }

  function showToast(text, kind = 'ok') {
    if (toastTimer) clearTimeout(toastTimer)
    toast.value = { text, kind, visible: true }
    toastTimer = setTimeout(() => {
      toast.value = { ...toast.value, visible: false }
    }, 3200)
  }

  function stopStats() {
    if (statsTimer) clearInterval(statsTimer)
    statsTimer = null
  }

  function clearSession() {
    token.value = null
    me.value = null
    status.value = 'anonymous'
    users.value = []
    usersLoaded.value = false
    usersTabLoading.value = false
    usersError.value = ''
    stats.value = null
    prevStats.value = null
    activeTab.value = 'dashboard'
    usersTabLoaded = false
    stopStats()
    clearToast()
    setApiToken(null)
  }

  function handleUnauthorized() {
    clearSession()
  }

  setUnauthorizedHandler(handleUnauthorized)

  async function login(email, password) {
    status.value = 'authenticating'
    try {
      const data = await adminApi.login(email, password)
      token.value = data?.token ?? null
      setApiToken(token.value)

      let list
      try {
        list = await adminApi.listUsers()
      } catch (error) {
        clearSession()
        if (error.status === 403) throw new Error('Acesso negado: apenas administradores')
        throw error
      }

      const normalizedEmail = String(email).toLowerCase().trim()
      const found = (list || []).map(normalizeUser).find((user) => user.email === normalizedEmail)
      if (!found || !found.isAdmin) {
        clearSession()
        throw new Error('Acesso negado: apenas administradores')
      }

      me.value = found
      status.value = 'authenticated'
      return found
    } catch (error) {
      token.value = null
      me.value = null
      status.value = 'anonymous'
      setApiToken(null)
      throw error
    }
  }

  function logout() {
    clearSession()
  }

  async function refreshUsers() {
    if (refreshingUsers) return
    refreshingUsers = true
    usersError.value = ''
    try {
      const list = await adminApi.listUsers()
      users.value = (list || []).map(normalizeUser)
      usersLoaded.value = true
    } catch (error) {
      if (error.status !== 401) usersError.value = error.message
    } finally {
      refreshingUsers = false
    }
  }

  async function refreshDashboard() {
    if (refreshingDashboard) return
    refreshingDashboard = true
    try {
      const [snapshot, list] = await Promise.all([adminApi.getStats(), adminApi.listUsers()])
      const normalizedUsers = (list || []).map(normalizeUser)
      users.value = normalizedUsers
      usersLoaded.value = true
      prevStats.value = previousSnapshot(stats.value)
      stats.value = normalizeStats(snapshot, normalizedUsers)
    } catch (error) {
      if (error.status !== 401) console.warn('dashboard:', error.message)
    } finally {
      refreshingDashboard = false
    }
  }

  function startStats() {
    stopStats()
    refreshDashboard()
    statsTimer = setInterval(refreshDashboard, 10000)
  }

  function setActiveTab(tab) {
    if (tab !== 'dashboard' && tab !== 'users') return
    activeTab.value = tab
    if (tab === 'users') {
      if (!usersTabLoaded) {
        usersTabLoaded = true
        usersTabLoading.value = true
        refreshUsers().finally(() => {
          usersTabLoading.value = false
        })
      } else {
        refreshUsers()
      }
    }
    if (tab === 'dashboard') refreshDashboard()
  }

  async function userAction(action, id) {
    if (me.value && id === me.value.id) return
    const target = users.value.find((user) => user.id === id)
    const name = target?.email || id

    try {
      if (action === 'suspend') {
        if (!window.confirm(`Suspender ${name}? Ele não poderá mais entrar.`)) return
        await adminApi.patchUser(id, { is_active: false })
        showToast('Usuário suspenso')
      } else if (action === 'reactivate') {
        await adminApi.patchUser(id, { is_active: true })
        showToast('Usuário reativado')
      } else if (action === 'promote') {
        if (!window.confirm(`Promover ${name} a administrador?`)) return
        await adminApi.patchUser(id, { is_admin: true })
        showToast('Usuário promovido a admin')
      } else if (action === 'demote') {
        if (!window.confirm(`Remover privilégios de admin de ${name}?`)) return
        await adminApi.patchUser(id, { is_admin: false })
        showToast('Privilégios de admin removidos')
      } else if (action === 'delete') {
        if (!window.confirm(`Excluir definitivamente ${name}? Essa ação não pode ser desfeita.`)) return
        await adminApi.deleteUser(id)
        showToast('Usuário excluído')
      } else {
        return
      }
      await refreshUsers()
      refreshDashboard()
    } catch (error) {
      if (error.status === 401) return
      showToast(error.message, 'err')
    }
  }

  return {
    token,
    me,
    status,
    users,
    usersLoaded,
    usersTabLoading,
    usersError,
    stats,
    prevStats,
    activeTab,
    toast,
    isAdmin,
    login,
    logout,
    refreshUsers,
    refreshDashboard,
    startStats,
    stopStats,
    setActiveTab,
    userAction,
    showToast,
    clearSession,
  }
})
