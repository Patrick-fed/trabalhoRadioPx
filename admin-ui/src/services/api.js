export const API = ''

let token = null
let unauthorizedHandler = null

export class ApiError extends Error {
  constructor(message, status) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export function setApiToken(nextToken) {
  token = nextToken ?? null
}

export function setUnauthorizedHandler(handler) {
  unauthorizedHandler = handler
}

export async function api(path, options = {}) {
  const { skipAuth = false, headers: extraHeaders = {}, ...rest } = options
  const headers = { 'Content-Type': 'application/json', ...extraHeaders }
  if (token && !skipAuth) headers.Authorization = `Bearer ${token}`

  let response
  try {
    response = await fetch(`${API}${path}`, { ...rest, headers })
  } catch {
    throw new ApiError('Não foi possível conectar ao servidor')
  }

  if (response.status === 401) {
    if (unauthorizedHandler) unauthorizedHandler()
    throw new ApiError('Sessão expirada. Entre novamente.', 401)
  }
  if (response.status === 429) {
    throw new ApiError('Muitas tentativas. Aguarde um momento.', 429)
  }

  const text = await response.text()
  let body = null
  if (text) {
    try {
      body = JSON.parse(text)
    } catch {
      body = { error: text }
    }
  }

  if (!response.ok) {
    const message = (body && (body.error || body.message)) || `Erro ${response.status}`
    throw new ApiError(message, response.status)
  }

  return body
}
