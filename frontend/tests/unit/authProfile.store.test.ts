// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '~/stores/auth'

global.$fetch = vi.fn()

const AVATAR = 'data:image/webp;base64,AAAA'

describe('Auth store profile updates', () => {
  let cookies
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    cookies = {}
    global.useCookie = vi.fn((name) => ({
      get value() { return cookies[name] ?? null },
      set value(v) { cookies[name] = v }
    }))
  })

  const loggedIn = () => {
    const store = useAuthStore()
    store.token = 'tok'
    store.user = { id: 'u1', username: 'a', email: 'a@x.id', fullName: 'Old', roles: ['admin'], avatarUrl: 'data:image/png;base64,OLD' }
    return store
  }
  const base = { fullName: 'New', phone: '1', department: 'Finance', position: 'Mgr' }
  const sentBody = () => vi.mocked($fetch).mock.calls[0][1].body

  it('omits avatar_url when the avatar was not changed', async () => {
    const store = loggedIn()
    vi.mocked($fetch).mockResolvedValueOnce({ data: { id: 'u1', full_name: 'New', phone: '1', department: 'Finance', position: 'Mgr', avatar_url: 'data:image/png;base64,OLD' } })
    await store.updateProfile(base)
    expect(sentBody()).toEqual({ full_name: 'New', phone: '1', department: 'Finance', position: 'Mgr' })
    expect('avatar_url' in sentBody()).toBe(false)
    expect(store.user.avatarUrl).toBe('data:image/png;base64,OLD')
  })

  it('sends the data URL and applies the avatar returned by the server', async () => {
    const store = loggedIn()
    vi.mocked($fetch).mockResolvedValueOnce({ data: { full_name: 'New', avatar_url: AVATAR } })
    await store.updateProfile({ ...base, avatarUrl: AVATAR })
    expect(sentBody().avatar_url).toBe(AVATAR)
    expect(vi.mocked($fetch).mock.calls[0][1].method).toBe('PUT')
    expect(store.user.avatarUrl).toBe(AVATAR)
    expect(store.user.fullName).toBe('New')
  })

  it('sends an empty string to remove the avatar', async () => {
    const store = loggedIn()
    vi.mocked($fetch).mockResolvedValueOnce({ data: { avatar_url: null } })
    await store.updateProfile({ ...base, avatarUrl: '' })
    expect(sentBody().avatar_url).toBe('')
    expect(store.user.avatarUrl).toBe('')
  })

  it('never writes the avatar to the auth-user cookie', async () => {
    const store = loggedIn()
    vi.mocked($fetch).mockResolvedValueOnce({ data: { avatar_url: AVATAR } })
    await store.updateProfile({ ...base, avatarUrl: AVATAR })
    expect(cookies['auth-user']).toBeTruthy()
    expect(cookies['auth-user']).not.toContain('avatar')
    expect(cookies['auth-user']).not.toContain('base64')
    expect(JSON.parse(cookies['auth-user']).fullName).toBe('New')
  })

  it('keeps status and data on failure and leaves the store untouched', async () => {
    const store = loggedIn()
    vi.mocked($fetch).mockRejectedValueOnce({ status: 400, data: { error: { message: 'Photo is too large' } } })
    await expect(store.updateProfile({ ...base, avatarUrl: AVATAR })).rejects.toMatchObject({ status: 400 })
    expect(store.user.fullName).toBe('Old')
    expect(store.user.avatarUrl).toBe('data:image/png;base64,OLD')
  })

  it('maps avatar_url on login and keeps it out of the cookie', async () => {
    const store = useAuthStore()
    vi.mocked($fetch).mockResolvedValueOnce({ data: { token: 't', user: { id: 'u1', email: 'e', full_name: 'F', avatar_url: AVATAR, roles: [] } } })
    vi.mocked($fetch).mockResolvedValueOnce({ data: { has_accepted: true } })
    await store.login({ username: 'a', password: 'b' })
    expect(store.user.avatarUrl).toBe(AVATAR)
    expect('avatar_url' in store.user).toBe(false)
    expect(cookies['auth-user']).not.toContain('base64')
  })

  it('hydrates the avatar once after a cookie restore', async () => {
    const store = loggedIn()
    store.user.avatarUrl = undefined
    vi.mocked($fetch).mockResolvedValue({ data: { id: 'u1', avatar_url: AVATAR } })
    await store.hydrateProfile()
    await store.hydrateProfile()
    expect($fetch).toHaveBeenCalledTimes(1)
    expect(store.user.avatarUrl).toBe(AVATAR)
  })
})
