<template>
  <div class="login-page">
    <div class="login-box">
      <h1>SSO Demo</h1>
      <p>練習用 OAuth2 SSO 登入</p>
      <button @click="auth.login()">使用 Demo 帳號登入（SSO）</button>
    </div>
  </div>
</template>

<script setup>
import { useAuthStore } from '@/stores/auth'
const auth = useAuthStore()
</script>

<style scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: #0f172a;
}

.login-box {
  background: #1e293b;
  padding: 48px;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.4);
  text-align: center;
  width: 320px;
  border: 1px solid #334155;
}

h1 {
  margin-bottom: 8px;
  color: #f1f5f9;
}

p {
  color: #94a3b8;
  margin-bottom: 32px;
  font-size: 14px;
}

button {
  width: 100%;
  padding: 12px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  cursor: pointer;
  transition: background 0.2s;
}

button:hover {
  background: #2563eb;
}
</style>
