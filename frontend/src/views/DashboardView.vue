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

      <button @click="auth.logout()">登出</button>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

onMounted(async () => {
  await auth.fetchUser()
})
</script>

<style scoped>
.dashboard {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: #0f172a;
}

.card {
  background: #1e293b;
  padding: 48px;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.4);
  width: 400px;
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
  margin-bottom: 32px;
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
