import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

interface UserInfo {
  id: number
  username: string
  role: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  function setAuth(t: string, u: UserInfo) {
    token.value = t
    user.value = u
    localStorage.setItem('token', t)
  }

  function clearAuth() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  async function login(username: string, password: string) {
    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || '登录失败')
    }
    const data = await res.json()
    setAuth(data.token, data.user)
  }

  function logout() {
    clearAuth()
  }

  async function fetchUser() {
    if (!token.value) return
    try {
      const res = await fetch('/api/me', {
        headers: { Authorization: `Bearer ${token.value}` },
      })
      if (!res.ok) {
        clearAuth()
        return
      }
      user.value = await res.json()
    } catch {
      clearAuth()
    }
  }

  return { token, user, isLoggedIn, login, logout, fetchUser }
})
