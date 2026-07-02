<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

onMounted(async () => {
  await authStore.fetchUser()
})

function handleLogout() {
  authStore.logout()
  router.push('/')
}
</script>

<template>
  <div class="admin-container">
    <h2>管理后台</h2>
    <div class="user-card">
      <p><strong>用户名：</strong>{{ authStore.user?.username }}</p>
      <p><strong>角色：</strong>{{ authStore.user?.role === 'admin' ? '管理员' : '普通用户' }}</p>
    </div>
    <button class="logout-btn" @click="handleLogout">退出登录</button>
  </div>
</template>

<style scoped>
.admin-container {
  max-width: 600px;
  margin: 2rem auto;
  padding: 2rem;
}

.admin-container h2 {
  margin-bottom: 1.5rem;
  color: var(--color-heading);
}

.user-card {
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-background-soft);
}

.user-card p {
  margin: 0.5rem 0;
}

.logout-btn {
  margin-top: 1.5rem;
  padding: 0.6rem 1.5rem;
  border: none;
  border-radius: 4px;
  background: #e74c3c;
  color: #fff;
  font-size: 0.95rem;
  cursor: pointer;
}

.logout-btn:hover {
  opacity: 0.9;
}
</style>
