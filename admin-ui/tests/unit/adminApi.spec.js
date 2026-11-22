import { beforeEach, describe, expect, it, vi } from 'vitest'
import { deleteUser, getStats, listUsers, login, patchUser } from '../../src/services/adminApi'
import { setApiToken } from '../../src/services/api'

function response(body = {}) {
  return {
    status: 200,
    ok: true,
    text: async () => JSON.stringify(body),
  }
}

describe('adminApi', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    setApiToken(null)
  })

  it('POST /api/v1/auth/login com body de credenciais e sem Authorization', async () => {
    const fetchMock = vi.fn().mockResolvedValue(response({ token: 'jwt' }))
    vi.stubGlobal('fetch', fetchMock)

    await login('admin@radiopx.local', 'senha')

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/auth/login',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'admin@radiopx.local', password: 'senha' }),
      }),
    )
    expect(fetchMock.mock.calls[0][1].headers).not.toHaveProperty('Authorization')
  })

  it('GET /api/v1/admin/users com token', async () => {
    setApiToken('jwt')
    const fetchMock = vi.fn().mockResolvedValue(response([]))
    vi.stubGlobal('fetch', fetchMock)

    await listUsers()

    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/admin/users')
    expect(fetchMock.mock.calls[0][1].method).toBeUndefined()
    expect(fetchMock.mock.calls[0][1].headers.Authorization).toBe('Bearer jwt')
  })

  it('GET /api/v1/admin/stats', async () => {
    setApiToken('jwt')
    const fetchMock = vi.fn().mockResolvedValue(response({ users_total: 1 }))
    vi.stubGlobal('fetch', fetchMock)

    await getStats()

    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/admin/stats')
  })

  it('PATCH /api/v1/admin/users/{id} com corpo de is_active', async () => {
    setApiToken('jwt')
    const fetchMock = vi.fn().mockResolvedValue(response({}))
    vi.stubGlobal('fetch', fetchMock)

    await patchUser('u1', { is_active: false })

    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/admin/users/u1')
    expect(fetchMock.mock.calls[0][1]).toMatchObject({
      method: 'PATCH',
      body: JSON.stringify({ is_active: false }),
    })
  })

  it('PATCH /api/v1/admin/users/{id} com corpo de is_admin', async () => {
    setApiToken('jwt')
    const fetchMock = vi.fn().mockResolvedValue(response({}))
    vi.stubGlobal('fetch', fetchMock)

    await patchUser('u1', { is_admin: true })

    expect(fetchMock.mock.calls[0][1].body).toBe(JSON.stringify({ is_admin: true }))
  })

  it('DELETE /api/v1/admin/users/{id} sem body', async () => {
    setApiToken('jwt')
    const fetchMock = vi.fn().mockResolvedValue({ status: 204, ok: true, text: async () => '' })
    vi.stubGlobal('fetch', fetchMock)

    await deleteUser('u1')

    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/admin/users/u1')
    expect(fetchMock.mock.calls[0][1].method).toBe('DELETE')
    expect(fetchMock.mock.calls[0][1].body).toBeUndefined()
  })
})
