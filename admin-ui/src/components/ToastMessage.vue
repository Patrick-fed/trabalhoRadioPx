<script setup>
import { computed } from 'vue'
import { useAdminStore } from '../stores/admin'

const store = useAdminStore()
const toast = computed(() => store.toast)
</script>

<template>
  <div v-if="toast.visible" class="toast" :class="toast.kind">
    {{ toast.text }}
  </div>
</template>
