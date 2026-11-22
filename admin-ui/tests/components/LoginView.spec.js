import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import LoginView from '../../src/views/LoginView.vue'

const replace = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal()
  return { ...actual, useRouter: () => ({ replace }) }
})

function response(status, body) {
  return {
    status,
    ok: status >= 200 && status < 300,
    text: async () => (body === undefined ? '' : JSON.stringify(body)),
  }
}

const adminUser = {
  id: 'admin-1',
  name: 'Admin',
  email: 'admin@radiopx.local',
  is_active: true,
  is_admin: true,
}

function mountLogin() {
  return mount(LoginView, { global: { plugins: [createPinia()] } })
}

describe('LoginView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('entra com admin via replace e sem erro', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response(200, { token: 'jwt' }))
        .mockResolvedValueOnce(response(200, [adminUser])),
    )
    const wrapper = mountLogin()

    await wrapper.find('#email').setValue('admin@radiopx.local')
    await wrapper.find('#password').setValue('senha')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(replace).toHaveBeenCalledWith({ name: 'admin' })
    expect(wrapper.find('#login-error').exists()).toBe(false)
  })

  it('mostra a mensagem exata para usuário comum', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response(200, { token: 'jwt' }))
        .mockResolvedValueOnce(response(200, [{ ...adminUser, is_admin: false }])),
    )
    const wrapper = mountLogin()

    await wrapper.find('#email').setValue('admin@radiopx.local')
    await wrapper.find('#password').setValue('senha')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(replace).not.toHaveBeenCalled()
    expect(wrapper.find('#login-error').text()).toBe('Acesso negado: apenas administradores')
  })
})
