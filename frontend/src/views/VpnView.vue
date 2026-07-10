<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const { t } = useI18n()
const authHeader = () => ({ Authorization: `Bearer ${authStore.token}` })

interface Settings {
  enabled: boolean
  subnet: string
  subnet6: string
  mtu: number
  interface_name: string
  allow_client_to_client: boolean
  do_local_ip_config: boolean
  do_remote_ip_config: boolean
}
interface ClientInfo {
  user_id: number
  username: string
  ip: string
  ip6?: string
  connected_at: string
}
interface Status {
  enabled: boolean
  online: number
  used_ips: number
  capacity: number
  used_ips6?: number
  capacity6?: number
  clients: ClientInfo[]
}
interface Reservation {
  id: number
  user_id: number
  username: string
  ip_address: string
  ip_address6?: string
  created_at: string
}
interface User {
  id: number
  username: string
}
interface Diag {
  platform: string
  is_root: boolean
  has_cap_net_admin: boolean | null
  cap_net_admin_note?: string
  ip_forward: boolean | null
  ip_forward_note?: string
  masquerade: boolean | null
  masquerade_note?: string
  ip6_forward: boolean | null
  ip6_forward_note?: string
  masquerade6: boolean | null
  masquerade6_note?: string
  tun_create: string
  tun_running: boolean
  tun_name?: string
}

const settings = ref<Settings | null>(null)
const status = ref<Status | null>(null)
const reservations = ref<Reservation[]>([])
const users = ref<User[]>([])
const diag = ref<Diag | null>(null)
const loading = ref(false)
const error = ref('')
const saving = ref(false)
const saveMsg = ref('')
const saveOk = ref(false)

const form = ref<Settings>({
  enabled: false,
  subnet: '192.168.77.0/24',
  subnet6: '',
  mtu: 1420,
  interface_name: '',
  allow_client_to_client: false,
  do_local_ip_config: true,
  do_remote_ip_config: true,
})

async function fetchSettings() {
  try {
    const res = await fetch('/api/admin/vpn/settings', { headers: authHeader() })
    if (!res.ok) throw new Error(t('common.loadFailed'))
    const data = await res.json()
    form.value = { ...data }
    settings.value = { ...data }
  } catch (e: any) {
    error.value = e.message
  }
}

async function fetchStatus() {
  try {
    const res = await fetch('/api/admin/vpn/status', { headers: authHeader() })
    if (!res.ok) throw new Error(t('common.loadFailed'))
    status.value = await res.json()
  } catch (e: any) {
    error.value = e.message
  }
}

async function fetchReservations() {
  try {
    const res = await fetch('/api/admin/vpn/reservations', { headers: authHeader() })
    if (!res.ok) throw new Error(t('common.loadFailed'))
    const data = await res.json()
    reservations.value = data.reservations
  } catch (e: any) {
    error.value = e.message
  }
}

async function fetchUsers() {
  try {
    const res = await fetch('/api/admin/users', { headers: authHeader() })
    if (!res.ok) throw new Error(t('common.loadFailed'))
    const data = await res.json()
    users.value = data.users
  } catch (e: any) {
    error.value = e.message
  }
}

async function fetchDiag() {
  try {
    const res = await fetch('/api/admin/vpn/diag', { headers: authHeader() })
    if (!res.ok) throw new Error(t('common.loadFailed'))
    diag.value = await res.json()
  } catch (e: any) {
    error.value = e.message
  }
}

async function handleSave() {
  saving.value = true
  saveMsg.value = ''
  try {
    const res = await fetch('/api/admin/vpn/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify(form.value),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || t('common.saveFailed'))
    saveMsg.value = t('vpn.saveSuccess')
    saveOk.value = true
    await Promise.all([fetchSettings(), fetchStatus(), fetchDiag()])
  } catch (e: any) {
    saveMsg.value = e.message
    saveOk.value = false
  } finally {
    saving.value = false
  }
}

const showAddResv = ref(false)
const resvForm = ref({ user_id: 0, ip_address: '', ip_address6: '' })
const resvError = ref('')

async function handleAddResv() {
  resvError.value = ''
  if (!resvForm.value.user_id || (!resvForm.value.ip_address && !resvForm.value.ip_address6)) {
    resvError.value = t('vpn.selectUserAndIp')
    return
  }
  try {
    const res = await fetch('/api/admin/vpn/reservations', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify(resvForm.value),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || t('common.createFailed'))
    showAddResv.value = false
    resvForm.value = { user_id: 0, ip_address: '', ip_address6: '' }
    await fetchReservations()
  } catch (e: any) {
    resvError.value = e.message
  }
}

async function handleDeleteResv(id: number) {
  if (!confirm(t('vpn.confirmDeleteReservation'))) return
  try {
    const res = await fetch(`/api/admin/vpn/reservations/${id}`, {
      method: 'DELETE',
      headers: authHeader(),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || t('common.deleteFailed'))
    await fetchReservations()
  } catch (e: any) {
    error.value = e.message
  }
}

async function handleKick(userId: number, username: string) {
  if (!confirm(t('vpn.confirmKick', { username }))) return
  try {
    const res = await fetch(`/api/admin/vpn/clients/${userId}`, {
      method: 'DELETE',
      headers: authHeader(),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || t('vpn.kickFailed'))
    await fetchStatus()
  } catch (e: any) {
    error.value = e.message
  }
}

function checkTunCreate(): boolean | null {
  if (!diag.value) return null
  return diag.value.tun_create.startsWith('ok')
}

async function refreshAll() {
  loading.value = true
  error.value = ''
  await Promise.all([fetchSettings(), fetchStatus(), fetchReservations(), fetchUsers(), fetchDiag()])
  loading.value = false
}

onMounted(() => {
  refreshAll()
})
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-8 space-y-8">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('vpn.title') }}</h2>
      <button
        class="px-4 py-2 rounded-lg font-medium text-sm text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
        @click="refreshAll"
      >
        {{ t('vpn.refresh') }}
      </button>
    </div>

    <p v-if="error" class="text-red-500">{{ error }}</p>

    <!-- 状态 -->
    <div class="grid gap-4 sm:grid-cols-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('vpn.serviceStatus') }}</p>
        <p class="text-xl font-bold mt-1" :class="status?.enabled ? 'text-green-600' : 'text-gray-400'">
          {{ status?.enabled ? t('vpn.running') : t('vpn.stopped') }}
        </p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('vpn.onlineClients') }}</p>
        <p class="text-xl font-bold text-gray-900 dark:text-white mt-1">{{ status?.online ?? '--' }}</p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('vpn.ipUsage') }}</p>
        <p class="text-xl font-bold text-gray-900 dark:text-white mt-1">{{ status?.used_ips ?? '--' }} / {{ status?.capacity ?? '--' }}</p>
      </div>
    </div>

    <!-- 系统环境检测 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('vpn.systemCheck') }}</h3>
      <div class="grid gap-3 sm:grid-cols-2">
        <!-- TUN 创建状态 -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span v-if="diag?.tun_running" class="text-green-500 mt-0.5">✓</span>
          <span v-else-if="checkTunCreate()" class="text-green-500 mt-0.5">✓</span>
          <span v-else class="text-red-500 mt-0.5">✗</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.tunDevice') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              <template v-if="diag?.tun_running">{{ t('vpn.tunRunning') }} ({{ diag.tun_name }})</template>
              <template v-else>{{ diag?.tun_create }}</template>
            </p>
          </div>
        </div>

        <!-- Root -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span :class="diag?.is_root ? 'text-green-500' : 'text-red-500'" class="mt-0.5">{{ diag?.is_root ? '✓' : '✗' }}</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.rootPermission') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ diag?.is_root ? t('vpn.runningAsRoot') : t('vpn.notRunningAsRoot') }}</p>
          </div>
        </div>

        <!-- CAP_NET_ADMIN -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span v-if="diag?.has_cap_net_admin === null" class="text-gray-400 mt-0.5">—</span>
          <span v-else-if="diag?.has_cap_net_admin" class="text-green-500 mt-0.5">✓</span>
          <span v-else class="text-red-500 mt-0.5">✗</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.capNetAdmin') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              <template v-if="diag?.has_cap_net_admin === null">{{ diag?.cap_net_admin_note }}</template>
              <template v-else>{{ diag?.has_cap_net_admin ? t('vpn.authorized') : diag?.cap_net_admin_note || t('vpn.unauthorized') }}</template>
            </p>
          </div>
        </div>

        <!-- IP Forward -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span v-if="diag?.ip_forward === null" class="text-gray-400 mt-0.5">—</span>
          <span v-else-if="diag?.ip_forward" class="text-green-500 mt-0.5">✓</span>
          <span v-else class="text-red-500 mt-0.5">✗</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.ipForward') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              <template v-if="diag?.ip_forward === null">{{ diag?.ip_forward_note }}</template>
              <template v-else>{{ diag?.ip_forward ? t('common.enabled') : diag?.ip_forward_note || t('vpn.notEnabled') }}</template>
            </p>
          </div>
        </div>

        <!-- MASQUERADE -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span v-if="diag?.masquerade === null" class="text-gray-400 mt-0.5">—</span>
          <span v-else-if="diag?.masquerade" class="text-green-500 mt-0.5">✓</span>
          <span v-else class="text-red-500 mt-0.5">✗</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.natMasqueradeV4') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              <template v-if="diag?.masquerade === null">{{ diag?.masquerade_note }}</template>
              <template v-else>{{ diag?.masquerade ? t('vpn.configured') : diag?.masquerade_note || t('vpn.notConfigured') }}</template>
            </p>
          </div>
        </div>

        <!-- IPv6 Forward -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span v-if="diag?.ip6_forward === null" class="text-gray-400 mt-0.5">—</span>
          <span v-else-if="diag?.ip6_forward" class="text-green-500 mt-0.5">✓</span>
          <span v-else class="text-red-500 mt-0.5">✗</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.ipv6Forward') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              <template v-if="diag?.ip6_forward === null">{{ diag?.ip6_forward_note }}</template>
              <template v-else>{{ diag?.ip6_forward ? t('common.enabled') : diag?.ip6_forward_note || t('vpn.notEnabled') }}</template>
            </p>
          </div>
        </div>

        <!-- IPv6 MASQUERADE -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span v-if="diag?.masquerade6 === null" class="text-gray-400 mt-0.5">—</span>
          <span v-else-if="diag?.masquerade6" class="text-green-500 mt-0.5">✓</span>
          <span v-else class="text-red-500 mt-0.5">✗</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.natMasqueradeV6') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              <template v-if="diag?.masquerade6 === null">{{ diag?.masquerade6_note }}</template>
              <template v-else>{{ diag?.masquerade6 ? t('vpn.configured') : diag?.masquerade6_note || t('vpn.notConfigured') }}</template>
            </p>
          </div>
        </div>

        <!-- 平台 -->
        <div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/30">
          <span class="text-gray-400 mt-0.5">ℹ</span>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('vpn.platform') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ diag?.platform }}</p>
          </div>
        </div>
      </div>
      <p v-if="diag && diag.platform !== 'linux'" class="text-xs text-gray-400 mt-3">
        {{ t('vpn.nonLinuxNote') }}
      </p>
    </div>

    <!-- 设置 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('vpn.tunnelSettings') }}</h3>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.enableVpn') }}</label>
          <select v-model="form.enabled" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
            <option :value="true">{{ t('vpn.enable') }}</option>
            <option :value="false">{{ t('vpn.stop') }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.subnet') }}</label>
          <input v-model="form.subnet" type="text" placeholder="192.168.77.0/24" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.ipv6Subnet') }}</label>
          <input v-model="form.subnet6" type="text" placeholder="fd00:dead:beef::/112" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.mtu') }}</label>
          <input v-model.number="form.mtu" type="number" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.interfaceName') }}</label>
          <input v-model="form.interface_name" type="text" placeholder="tun0" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.allowClientToClient') }}</label>
          <select v-model="form.allow_client_to_client" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
            <option :value="false">{{ t('vpn.deny') }}</option>
            <option :value="true">{{ t('vpn.allow') }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.serverConfigTunIp') }}</label>
          <select v-model="form.do_local_ip_config" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
            <option :value="true">{{ t('vpn.autoConfig') }}</option>
            <option :value="false">{{ t('vpn.manual') }}</option>
          </select>
        </div>
      </div>
      <div class="flex items-center gap-4 mt-6">
        <button
          class="px-4 py-2 rounded-lg font-medium text-sm text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition-colors"
          :disabled="saving"
          @click="handleSave"
        >
          {{ saving ? t('common.saving') : t('vpn.saveSettings') }}
        </button>
        <span v-if="saveMsg" :class="saveOk ? 'text-green-500' : 'text-red-500'" class="text-sm">{{ saveMsg }}</span>
      </div>
    </div>

    <!-- 在线客户端 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white p-6 pb-4">{{ t('vpn.onlineClients') }}</h3>
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.user') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.ipv4') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.ipv6') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.connectTime') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!status?.clients?.length">
            <td colspan="5" class="px-6 py-6 text-center text-gray-400">{{ t('vpn.noOnlineClients') }}</td>
          </tr>
          <tr v-for="(c, i) in status?.clients" :key="i" class="border-b border-gray-100 dark:border-gray-700/50">
            <td class="px-6 py-3 text-gray-900 dark:text-white font-medium">{{ c.username }}</td>
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ c.ip }}</td>
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ c.ip6 || '—' }}</td>
            <td class="px-6 py-3 text-gray-500 dark:text-gray-400">{{ c.connected_at }}</td>
            <td class="px-6 py-3">
              <button
                class="px-3 py-1 text-xs rounded-md font-medium text-red-700 bg-red-50 hover:bg-red-100 dark:text-red-400 dark:bg-red-900/20 transition-colors"
                @click="handleKick(c.user_id, c.username)"
              >
                {{ t('vpn.kick') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 静态预留 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div class="flex items-center justify-between p-6 pb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('vpn.staticIpReservation') }}</h3>
        <button class="px-3 py-1.5 text-xs rounded-md font-medium text-white bg-sky-600 hover:bg-sky-700 transition-colors" @click="showAddResv = true">{{ t('vpn.addReservation') }}</button>
      </div>
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.user') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.ipv4') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('vpn.ipv6') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.createdAt') }}</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!reservations.length">
            <td colspan="5" class="px-6 py-6 text-center text-gray-400">{{ t('vpn.noReservations') }}</td>
          </tr>
          <tr v-for="r in reservations" :key="r.id" class="border-b border-gray-100 dark:border-gray-700/50">
            <td class="px-6 py-3 text-gray-900 dark:text-white font-medium">{{ r.username }}</td>
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ r.ip_address || '—' }}</td>
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ r.ip_address6 || '—' }}</td>
            <td class="px-6 py-3 text-gray-500 dark:text-gray-400">{{ r.created_at }}</td>
            <td class="px-6 py-3">
              <button class="px-3 py-1 text-xs rounded-md font-medium text-red-700 bg-red-50 hover:bg-red-100 dark:text-red-400 dark:bg-red-900/20 transition-colors" @click="handleDeleteResv(r.id)">{{ t('common.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增预留弹窗 -->
    <div v-if="showAddResv" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showAddResv = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('vpn.addIpReservation') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.user') }}</label>
            <select v-model.number="resvForm.user_id" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              <option :value="0" disabled>{{ t('vpn.selectUser') }}</option>
              <option v-for="u in users" :key="u.id" :value="u.id">{{ u.username }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.ipv4Address') }}</label>
            <input v-model="resvForm.ip_address" type="text" placeholder="192.168.77.10" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('vpn.ipv6Address') }}</label>
            <input v-model="resvForm.ip_address6" type="text" placeholder="fd00:dead:beef::10" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
          </div>
          <p v-if="resvError" class="text-sm text-red-500">{{ resvError }}</p>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors" @click="showAddResv = false">{{ t('common.cancel') }}</button>
          <button class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 transition-colors" @click="handleAddResv">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
