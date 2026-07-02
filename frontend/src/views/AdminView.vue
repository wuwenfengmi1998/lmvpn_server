<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const stats = ref([
  { label: '运行时长', value: '--', unit: '', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' },
  { label: '活跃设备', value: '--', unit: '', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
  { label: '今日流量', value: '--', unit: 'GB', icon: 'M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z' },
  { label: '在线节点', value: '--', unit: '', icon: 'M5 12h14M12 5l7 7-7 7' },
])

onMounted(async () => {
  await authStore.fetchUser()
})

function handleLogout() {
  authStore.logout()
  router.push('/')
}
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 py-12">
    <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-8">管理后台</h2>

    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4 mb-8">
      <div
        v-for="stat in stats"
        :key="stat.label"
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5"
      >
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg bg-sky-100 dark:bg-sky-900/40 flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-sky-600 dark:text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" :d="stat.icon" />
            </svg>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ stat.label }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ stat.value }}<span class="text-sm font-normal text-gray-500">{{ stat.unit }}</span>
            </p>
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">用户信息</h3>
      <div class="space-y-2 text-sm text-gray-700 dark:text-gray-300">
        <p><span class="font-medium text-gray-900 dark:text-white">用户名：</span>{{ authStore.user?.username }}</p>
        <p><span class="font-medium text-gray-900 dark:text-white">角色：</span>{{ authStore.user?.role === 'admin' ? '管理员' : '普通用户' }}</p>
      </div>
    </div>

    <button
      class="px-6 py-2.5 rounded-lg font-medium text-white bg-red-500 hover:bg-red-600 transition-colors"
      @click="handleLogout"
    >
      退出登录
    </button>
  </div>
</template>
