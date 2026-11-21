<script setup>
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '../stores/admin'
import AdminHeader from '../components/AdminHeader.vue'
import DashboardStats from '../components/DashboardStats.vue'
import AdminMap from '../components/AdminMap.vue'
import UsersTable from '../components/UsersTable.vue'
import ToastMessage from '../components/ToastMessage.vue'

const router = useRouter()
const store = useAdminStore()
const email = ref(store.me?.email || '')
const mapRef = ref(null)

function changeTab(tab) {
  store.setActiveTab(tab)
}

function logout() {
  store.logout()
}

onMounted(() => {
  email.value = store.me?.email || ''
  if (store.status === 'authenticated') store.startStats()
})

onUnmounted(() => store.stopStats())

watch(
  () => store.status,
  (status) => {
    if (status !== 'authenticated') router.replace({ name: 'login' })
  },
)

watch(
  () => store.activeTab,
  async (tab) => {
    if (tab === 'dashboard') {
      await nextTick()
      mapRef.value?.invalidate()
    }
  },
)
</script>

<template>
  <div id="app-view" class="view">
    <AdminHeader
      :active-tab="store.activeTab"
      :email="email"
      @change-tab="changeTab"
      @logout="logout"
    />
    <main>
      <section v-show="store.activeTab === 'dashboard'" id="dashboard-tab" class="tab">
        <DashboardStats />
        <AdminMap ref="mapRef" />
      </section>
      <section v-show="store.activeTab === 'users'" id="users-tab" class="tab">
        <UsersTable />
      </section>
    </main>
    <ToastMessage />
  </div>
</template>
