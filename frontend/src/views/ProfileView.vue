<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import TrafficChart from '@/components/TrafficChart.vue'

const authStore = useAuthStore()
const router = useRouter()
const { t } = useI18n()

interface VpnConnection {
  ip: string
  ip6?: string
  connected_at: string
}
const vpnConnections = ref<VpnConnection[]>([])
const maxConns = ref(30)

interface TrafficRecord {
  date: string
  rx_bytes: number
  tx_bytes: number
}
const myTraffic7d = ref<TrafficRecord[]>([])
const todayRx = ref(0)
const todayTx = ref(0)

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}

async function fetchMyTraffic() {
  try {
    const res = await fetch('/api/me/traffic?days=7', {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (!res.ok) return
    const data = await res.json()
    myTraffic7d.value = data.records || []
    todayRx.value = data.today_rx_bytes || 0
    todayTx.value = data.today_tx_bytes || 0
  } catch {}
}

async function fetchVpnConnections() {
  try {
    const res = await fetch('/api/me/vpn/connections', {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (!res.ok) return
    const data = await res.json()
    vpnConnections.value = data.connections || []
    maxConns.value = data.max_conns_per_user || 30
  } catch {}
}

onMounted(async () => {
  await authStore.fetchUser()
  fetchVpnConnections()
  fetchMyTraffic()
})

const showPasswordModal = ref(false)
const passwordForm = ref({ old_password: '', new_password: '', confirm_password: '' })
const passwordError = ref('')
const changingPassword = ref(false)

function openPasswordModal() {
  passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  passwordError.value = ''
  showPasswordModal.value = true
}

async function handleChangePassword() {
  passwordError.value = ''
  if (!passwordForm.value.old_password || !passwordForm.value.new_password) {
    passwordError.value = t('profile.enterOldAndNewPassword')
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    passwordError.value = t('profile.passwordsDoNotMatch')
    return
  }
  changingPassword.value = true
  try {
    const res = await fetch('/api/me/password', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({
        old_password: passwordForm.value.old_password,
        new_password: passwordForm.value.new_password,
      }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || t('profile.passwordChangeFailed'))
    }
    showPasswordModal.value = false
    authStore.logout()
    router.push({ name: 'login', query: { msg: 'password_changed' } })
  } catch (e: any) {
    passwordError.value = e.message || t('profile.passwordChangeFailed')
  } finally {
    changingPassword.value = false
  }
}
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 py-12">
    <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-8">{{ t('profile.title') }}</h2>

    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('profile.basicInfo') }}</h3>
      <div class="space-y-2 text-sm text-gray-700 dark:text-gray-300">
        <p><span class="font-medium text-gray-900 dark:text-white">{{ t('profile.usernameLabel') }}</span>{{ authStore.user?.username }}</p>
        <p><span class="font-medium text-gray-900 dark:text-white">{{ t('profile.roleLabel') }}</span>{{ authStore.user?.role === 'admin' ? t('common.admin') : t('common.normalUser') }}</p>
      </div>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('traffic.myTraffic') }}</h3>
      <div class="grid grid-cols-3 gap-4 mb-6">
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">{{ t('traffic.upload') }}</p>
          <p class="text-lg font-bold text-sky-600 dark:text-sky-400 tabular-nums">{{ formatBytes(todayRx) }}</p>
        </div>
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">{{ t('traffic.download') }}</p>
          <p class="text-lg font-bold text-green-600 dark:text-green-400 tabular-nums">{{ formatBytes(todayTx) }}</p>
        </div>
        <div class="text-center">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">{{ t('traffic.total') }}</p>
          <p class="text-lg font-bold text-gray-900 dark:text-white tabular-nums">{{ formatBytes(todayRx + todayTx) }}</p>
        </div>
      </div>
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">{{ t('traffic.trafficHistory7d') }}</h4>
      <TrafficChart :records="myTraffic7d" />
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden mb-6">
      <div class="flex items-center justify-between p-6 pb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('profile.myVpnConnections') }}</h3>
        <span class="px-3 py-1 text-sm font-medium rounded-full" :class="vpnConnections.length > 0 ? 'bg-sky-100 text-sky-700 dark:bg-sky-900/40 dark:text-sky-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'">
          {{ vpnConnections.length }} / {{ maxConns }}
        </span>
      </div>
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.ipv4') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.ipv6') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.connectTime') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!vpnConnections.length">
            <td colspan="3" class="px-6 py-6 text-center text-gray-400">{{ t('profile.noConnections') }}</td>
          </tr>
          <tr v-for="(c, i) in vpnConnections" :key="i" class="border-b border-gray-100 dark:border-gray-700/50">
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ c.ip }}</td>
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ c.ip6 || '-' }}</td>
            <td class="px-6 py-3 text-gray-500 dark:text-gray-400">{{ c.connected_at }}</td>
          </tr>
        </tbody>
      </table>
      <p class="text-xs text-gray-400 px-6 py-4">{{ t('profile.abnormalConnectionHint') }}</p>
    </div>

    <button
      class="px-6 py-2.5 rounded-lg font-medium text-white bg-sky-600 hover:bg-sky-700 transition-colors"
      @click="openPasswordModal"
    >
      {{ t('profile.changePassword') }}
    </button>

    <div v-if="showPasswordModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showPasswordModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-sm mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('profile.changePassword') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('profile.oldPassword') }}</label>
            <input
              v-model="passwordForm.old_password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('profile.newPassword') }}</label>
            <input
              v-model="passwordForm.new_password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('profile.confirmNewPassword') }}</label>
            <input
              v-model="passwordForm.confirm_password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            />
          </div>
          <p v-if="passwordError" class="text-sm text-red-500">{{ passwordError }}</p>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="showPasswordModal = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition-colors"
            :disabled="changingPassword"
            @click="handleChangePassword"
          >
            {{ changingPassword ? t('common.saving') : t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
