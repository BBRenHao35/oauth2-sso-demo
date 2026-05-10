import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)   // null = 未登入，有值 = 已登入

  // 呼叫後端 /api/auth/me 確認目前登入狀態
  async function fetchUser() {
    try {
      const res = await fetch('/api/auth/me', { credentials: 'include' })
      user.value = res.ok ? await res.json() : null
    } catch {
      user.value = null
    }
  }

  // 跳轉到後端 /api/auth/login，後端會導去 Keycloak
  function login() {
    window.location.href = '/api/auth/login'
  }

  // 跳轉到後端 /api/auth/logout，後端清 session 後導回首頁
  function logout() {
    window.location.href = '/api/auth/logout'
  }

  return { user, fetchUser, login, logout }
})
