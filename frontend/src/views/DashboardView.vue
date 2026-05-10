<template>
  <div class="dashboard">
    <div class="card">
      <h1>Dashboard</h1>
      <p class="subtitle">SSO 登入成功</p>

      <div v-if="auth.user" class="user-info">
        <div class="field">
          <span class="label">姓名</span>
          <span>{{ auth.user.name || '-' }}</span>
        </div>
        <div class="field">
          <span class="label">帳號</span>
          <span>{{ auth.user.username }}</span>
        </div>
        <div class="field">
          <span class="label">Email</span>
          <span>{{ auth.user.email || '-' }}</span>
        </div>
        <div class="field">
          <span class="label">角色</span>
          <span>{{ auth.user.roles?.join(', ') || '無' }}</span>
        </div>
      </div>

      <!-- 只有 Admin / Advanced 才看得到這個區塊 -->
      <div v-if="isAdmin" class="admin-section">
        <p class="admin-label">管理員功能</p>
        <button class="admin-btn" @click="fetchAdminData">呼叫管理員 API</button>
        <p v-if="adminResult" class="admin-result">{{ adminResult }}</p>
      </div>

      <!-- 非管理員看到這個 -->
      <div v-else class="no-permission">
        無管理員權限（可在 Keycloak 指派 Admin role 後重新登入測試）
      </div>

      <button @click="auth.logout()">登出</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const adminResult = ref('')

// 判斷是否為管理員角色
const isAdmin = computed(() =>
  auth.user?.roles?.some(r => ['admin', 'advanced'].includes(r.toLowerCase()))
)

async function fetchAdminData() {
  const res = await fetch('/api/admin/data', { credentials: 'include' })
  if (res.ok) {
    const data = await res.json()
    adminResult.value = JSON.stringify(data.data)
  } else {
    adminResult.value = `錯誤 ${res.status}：權限不足`
  }
}

onMounted(async () => {
  await auth.fetchUser()
})
</script>

<style scoped>
.dashboard {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #0f172a;
  padding: 24px;
}

.card {
  background: #1e293b;
  padding: 48px;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.4);
  width: 420px;
  border: 1px solid #334155;
}

h1 {
  margin-bottom: 4px;
  color: #f1f5f9;
}

.subtitle {
  color: #4ade80;
  margin-bottom: 32px;
  font-size: 14px;
}

.user-info {
  margin-bottom: 24px;
}

.field {
  display: flex;
  padding: 12px 0;
  border-bottom: 1px solid #334155;
  gap: 16px;
  color: #e2e8f0;
}

.label {
  color: #64748b;
  width: 60px;
  flex-shrink: 0;
}

.admin-section {
  background: #1a2f1a;
  border: 1px solid #2d5a2d;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.admin-label {
  color: #4ade80;
  font-size: 13px;
  margin-bottom: 12px;
}

.admin-btn {
  width: 100%;
  padding: 10px;
  background: #166534;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 8px;
}

.admin-btn:hover {
  background: #15803d;
}

.admin-result {
  color: #86efac;
  font-size: 13px;
  word-break: break-all;
}

.no-permission {
  color: #475569;
  font-size: 13px;
  text-align: center;
  padding: 12px;
  margin-bottom: 24px;
  border: 1px dashed #334155;
  border-radius: 8px;
}

button {
  width: 100%;
  padding: 12px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  cursor: pointer;
  transition: background 0.2s;
}

button:hover {
  background: #dc2626;
}
</style>
