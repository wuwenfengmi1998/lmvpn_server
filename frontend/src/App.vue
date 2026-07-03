<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

function handleLogout() {
  authStore.logout()
  router.push('/')
}

const navLinks = [
  { to: '/', label: '首页' },
  { to: '/about', label: '关于' },
]
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50 dark:bg-gray-950">
    <header class="bg-sky-600 text-white shadow-md">
      <div class="max-w-6xl mx-auto px-4 py-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <RouterLink to="/" class="text-xl font-bold tracking-wide hover:text-sky-200 transition-colors">
          LmVPN
        </RouterLink>
        <nav class="flex items-center gap-1 flex-wrap">
          <RouterLink
            v-for="link in navLinks"
            :key="link.to"
            :to="link.to"
            class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                   hover:bg-sky-500/40"
            active-class="bg-sky-700/60"
          >
            {{ link.label }}
          </RouterLink>
          <template v-if="authStore.isLoggedIn">
            <RouterLink
              to="/profile"
              class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                     hover:bg-sky-500/40"
              active-class="bg-sky-700/60"
            >
              用户信息
            </RouterLink>
            <template v-if="authStore.user?.role === 'admin'">
              <RouterLink
                to="/admin"
                class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                       hover:bg-sky-500/40"
                active-class="bg-sky-700/60"
              >
                管理后台
              </RouterLink>
              <RouterLink
                to="/admin/users"
                class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                       hover:bg-sky-500/40"
                active-class="bg-sky-700/60"
              >
                用户管理
              </RouterLink>
            </template>
            <a
              href="#"
              class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                     hover:bg-red-500/40"
              @click.prevent="handleLogout"
            >
              退出
            </a>
          </template>
          <template v-else>
            <RouterLink
              to="/login"
              class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                     hover:bg-sky-500/40"
              active-class="bg-sky-700/60"
            >
              登录
            </RouterLink>
          </template>
        </nav>
      </div>
    </header>

    <main class="flex-1">
      <RouterView />
    </main>

    <footer class="bg-slate-800 text-slate-400 py-6 text-center text-sm">
      <div class="max-w-6xl mx-auto px-4">
        <p>&copy; {{ new Date().getFullYear() }} LmVPN. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>
