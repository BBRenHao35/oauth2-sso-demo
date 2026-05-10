import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', name: 'login', component: LoginView },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
      meta: { requiresAuth: true },
    },
  ],
})

// 每次換頁前執行
router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return  // 不需要登入的頁面直接放行

  const auth = useAuthStore()

  // 還沒抓過 user 的話，先打 /api/auth/me 確認
  if (!auth.user) {
    await auth.fetchUser()
  }

  // 還是沒有 user → 未登入，導回 login
  if (!auth.user) {
    return { name: 'login' }
  }
})

export default router
