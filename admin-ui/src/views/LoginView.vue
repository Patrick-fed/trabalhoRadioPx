<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '../stores/admin'

const router = useRouter()
const store = useAdminStore()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await store.login(email.value, password.value)
    await router.replace({ name: 'admin' })
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

watch(
  () => store.status,
  (status) => {
    if (status === 'anonymous') password.value = ''
  },
)
</script>

<template>
  <div id="login-view" class="view">
    <div class="login-card">
      <h1>RadioPX</h1>
      <p class="subtitle">Painel Administrativo</p>
      <form @submit.prevent="submit">
        <label for="email">E-mail</label>
        <input id="email" v-model="email" type="email" autocomplete="username" required />

        <label for="password">Senha</label>
        <input
          id="password"
          v-model="password"
          type="password"
          autocomplete="current-password"
          required
        />

        <button id="login-btn" type="submit" :disabled="loading">Entrar</button>
        <p v-if="error" id="login-error" class="error">{{ error }}</p>
      </form>
    </div>
  </div>
</template>
