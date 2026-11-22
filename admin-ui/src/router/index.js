import { createRouter, createWebHistory } from 'vue-router'
import { useAdminStore } from '../stores/admin'
import LoginView from '../views/LoginView.vue'
import AdminShell from '../views/AdminShell.vue'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/login', name: 'login', component: LoginView },
    { path: '/', name: 'admin', component: AdminShell },
  ],
})

router.beforeEach((to) => {
  const store = useAdminStore()
  if (to.name !== 'login' && store.status !== 'authenticated') {
    return { name: 'login' }
  }
  if (to.name === 'login' && store.status === 'authenticated') {
    return { name: 'admin' }
  }
  return true
})

export default router
