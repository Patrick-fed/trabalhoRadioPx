import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api, setApiToken, setUnauthorizedHandler } from '../../src/services/api'

function response(status, body) {
  return {
    status,
    ok: status >= 200 && status < 300,
    text: async () => (body === undefined ? '' : JSON.stringify(body)),
  }
}

describe('api', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    setApiToken(null)
    setUnauthorizedHandler(null)
  })

  it('envia JSON, omite Authorization no login e envia Bearer quando há token', async () => {
    const fetchMock = vi.fn().mockResolvedValue(response(200, { token: 'jwt' }))
    vi.stubGlobal('fetch', fetchMock)

    await api('/api/v1/auth/login', {
      method: 'POST',
      skipAuth: true,
      body: JSON.stringify({ email: 'admin@radiopx.local', password: 'senha' }),
    })

    expect(fetchMock.mock.calls[0][1].headers).toMatchObject({
      'Content-Type': 'application/json',
    })
    expect(fetchMock.mock.calls[0][1].headers).not.toHaveProperty('Authorization')

    setApiToken('jwt')
    await api('/api/v1/admin/users')

    expect(fetchMock.mock.calls[1][1].headers).toMatchObject({
      'Content-Type': 'application/json',
      Authorization: 'Bearer jwt',
    })
  })

  it('limpa a sessão e usa a mensagem exata em 401', async () => {
    const onUnauthorized = vi.fn()
    setUnauthorizedHandler(onUnauthorized)
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response(401, { error: 'unauthorized' })))

    await expect(api('/api/v1/admin/users')).rejects.toThrow('Sessão expirada. Entre novamente.')
    expect(onUnauthorized).toHaveBeenCalledOnce()
  })

  it('usa a mensagem exata em 429', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response(429, { error: 'too many' })))

    await expect(api('/api/v1/auth/login', { method: 'POST' })).rejects.toThrow(
      'Muitas tentativas. Aguarde um momento.',
    )
  })

  it('usa a mensagem exata em falha de rede', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(api('/api/v1/admin/stats')).rejects.toThrow('Não foi possível conectar ao servidor')
  })

  it('prefere error/message do corpo e cai para Erro <status>', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(400, { message: 'payload inválido' }))
      .mockResolvedValueOnce(response(500, { unexpected: true }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(api('/api/v1/admin/users')).rejects.toThrow('payload inválido')
    await expect(api('/api/v1/admin/users')).rejects.toThrow('Erro 500')
  })
})
