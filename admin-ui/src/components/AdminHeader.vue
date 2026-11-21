<script setup>
defineProps({
  activeTab: { type: String, required: true },
  email: { type: String, default: '' },
})

defineEmits(['change-tab', 'logout'])
</script>

<template>
  <header>
    <div class="brand">
      <span class="logo">RadioPX</span>
      <span class="tag">Admin</span>
    </div>
    <nav>
      <button
        class="nav-btn"
        :class="{ active: activeTab === 'dashboard' }"
        type="button"
        @click="$emit('change-tab', 'dashboard')"
      >
        Dashboard
      </button>
      <button
        class="nav-btn"
        :class="{ active: activeTab === 'users' }"
        type="button"
        @click="$emit('change-tab', 'users')"
      >
        Usuários
      </button>
    </nav>
    <div class="header-right">
      <span class="admin-email">{{ email }}</span>
      <button id="logout-btn" class="btn-ghost" type="button" @click="$emit('logout')">
        Sair
      </button>
    </div>
  </header>
</template>
