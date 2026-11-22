import { api } from './api'

export function login(email, password) {
  return api('/api/v1/auth/login', {
    method: 'POST',
    skipAuth: true,
    body: JSON.stringify({ email, password }),
  })
}

export function listUsers() {
  return api('/api/v1/admin/users')
}

export function getStats() {
  return api('/api/v1/admin/stats')
}

export function patchUser(id, payload) {
  return api(`/api/v1/admin/users/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function deleteUser(id) {
  return api(`/api/v1/admin/users/${id}`, { method: 'DELETE' })
}

export const adminApi = { login, listUsers, getStats, patchUser, deleteUser }
