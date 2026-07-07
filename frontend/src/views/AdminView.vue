<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()
const stats = ref([
  { label: 'admin.uptime', value: '--', unit: '', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z', route: '' },
  { label: 'admin.activeDevices', value: '--', unit: '', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', route: '' },
  { label: 'admin.todayTraffic', value: '--', unit: 'GB', icon: 'M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z', route: '' },
  { label: 'admin.onlineNodes', value: '--', unit: '', icon: 'M5 12h14M12 5l7 7-7 7', route: '' },
  { label: 'admin.totalUsers', value: '--', unit: '', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z', route: '/admin/users' },
  { label: 'admin.vpnManage', value: 'admin.config', unit: '', icon: 'M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 4.05A12.884 12.884 0 0015 11a4 4 0 10-8 0c0 1.017.07 2.019.203 3M3 3l18 18', route: '/admin/vpn' },
])

const userCount = ref<number | null>(null)
let statsTimer: ReturnType<typeof setInterval> | null = null

function formatUptime(seconds: number): string {
  if (seconds <= 0) return '0m'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}d ${h}h`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

async function fetchUserCount() {
  try {
    const res = await fetch('/api/admin/users/count', {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (res.ok) {
      const data = await res.json()
      userCount.value = data.count
      const stat = stats.value.find(s => s.label === 'admin.totalUsers')
      if (stat) stat.value = String(data.count)
    }
  } catch {}
}

async function fetchStats() {
  try {
    const res = await fetch('/api/admin/stats', {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (!res.ok) return
    const data = await res.json()
    const set = (label: string, value: string) => {
      const stat = stats.value.find(s => s.label === label)
      if (stat) stat.value = value
    }
    set('admin.uptime', formatUptime(data.uptime_seconds))
    set('admin.activeDevices', String(data.active_devices))
    set('admin.todayTraffic', (data.today_traffic_bytes / 1e9).toFixed(2))
    set('admin.onlineNodes', String(data.online_nodes))
  } catch {}
}

onMounted(async () => {
  await authStore.fetchUser()
  fetchUserCount()
  fetchStats()
  statsTimer = setInterval(fetchStats, 30000)
})

onUnmounted(() => {
  if (statsTimer) clearInterval(statsTimer)
})

function handleStatClick(route: string) {
  if (route) router.push(route)
}

function handleLogout() {
  authStore.logout()
  router.push('/')
}
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 py-12">
    <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-8">{{ t('admin.title') }}</h2>

    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-5 mb-8">
      <div
        v-for="stat in stats"
        :key="stat.label"
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5 transition-colors"
        :class="stat.route ? 'cursor-pointer hover:bg-sky-50 dark:hover:bg-sky-900/20' : ''"
        @click="handleStatClick(stat.route)"
      >
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg bg-sky-100 dark:bg-sky-900/40 flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-sky-600 dark:text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" :d="stat.icon" />
            </svg>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t(stat.label) }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ stat.label === 'admin.vpnManage' ? t(stat.value) : stat.value }}<span class="text-sm font-normal text-gray-500">{{ stat.unit }}</span>
            </p>
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('admin.userInfo') }}</h3>
      <div class="space-y-2 text-sm text-gray-700 dark:text-gray-300">
        <p><span class="font-medium text-gray-900 dark:text-white">{{ t('admin.usernameLabel') }}</span>{{ authStore.user?.username }}</p>
        <p><span class="font-medium text-gray-900 dark:text-white">{{ t('admin.roleLabel') }}</span>{{ authStore.user?.role === 'admin' ? t('common.admin') : t('common.normalUser') }}</p>
      </div>
    </div>

    <button
      class="px-6 py-2.5 rounded-lg font-medium text-white bg-red-500 hover:bg-red-600 transition-colors"
      @click="handleLogout"
    >
      {{ t('admin.logoutButton') }}
    </button>
  </div>
</template>
