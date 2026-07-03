<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const authHeader = () => ({ Authorization: `Bearer ${authStore.token}` })

interface Settings {
  enabled: boolean
  subnet: string
  mtu: number
  interface_name: string
  allow_client_to_client: boolean
  do_local_ip_config: boolean
  do_remote_ip_config: boolean
}
interface ClientInfo {
  username: string
  ip: string
  connected_at: string
}
interface Status {
  enabled: boolean
  online: number
  used_ips: number
  capacity: number
  clients: ClientInfo[]
}
interface Reservation {
  id: number
  user_id: number
  username: string
  ip_address: string
  created_at: string
}
interface User {
  id: number
  username: string
}

const settings = ref<Settings | null>(null)
const status = ref<Status | null>(null)
const reservations = ref<Reservation[]>([])
const users = ref<User[]>([])
const loading = ref(false)
const error = ref('')
const saving = ref(false)
const saveMsg = ref('')

const form = ref<Settings>({
  enabled: false,
  subnet: '192.168.3.0/24',
  mtu: 1420,
  interface_name: '',
  allow_client_to_client: false,
  do_local_ip_config: true,
  do_remote_ip_config: true,
})

async function fetchSettings() {
  try {
    const res = await fetch('/api/admin/vpn/settings', { headers: authHeader() })
    if (!res.ok) throw new Error('加载失败')
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
    if (!res.ok) throw new Error('加载失败')
    status.value = await res.json()
  } catch (e: any) {
    error.value = e.message
  }
}

async function fetchReservations() {
  try {
    const res = await fetch('/api/admin/vpn/reservations', { headers: authHeader() })
    if (!res.ok) throw new Error('加载失败')
    const data = await res.json()
    reservations.value = data.reservations
  } catch (e: any) {
    error.value = e.message
  }
}

async function fetchUsers() {
  try {
    const res = await fetch('/api/admin/users', { headers: authHeader() })
    if (!res.ok) throw new Error('加载失败')
    const data = await res.json()
    users.value = data.users
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
    if (!res.ok) throw new Error(data.error || '保存失败')
    saveMsg.value = '保存成功'
    await Promise.all([fetchSettings(), fetchStatus()])
  } catch (e: any) {
    saveMsg.value = e.message
  } finally {
    saving.value = false
  }
}

const showAddResv = ref(false)
const resvForm = ref({ user_id: 0, ip_address: '' })
const resvError = ref('')

async function handleAddResv() {
  resvError.value = ''
  if (!resvForm.value.user_id || !resvForm.value.ip_address) {
    resvError.value = '请选择用户并填写 IP'
    return
  }
  try {
    const res = await fetch('/api/admin/vpn/reservations', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify(resvForm.value),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || '创建失败')
    showAddResv.value = false
    resvForm.value = { user_id: 0, ip_address: '' }
    await fetchReservations()
  } catch (e: any) {
    resvError.value = e.message
  }
}

async function handleDeleteResv(id: number) {
  if (!confirm('确认删除该预留？')) return
  try {
    const res = await fetch(`/api/admin/vpn/reservations/${id}`, {
      method: 'DELETE',
      headers: authHeader(),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || '删除失败')
    await fetchReservations()
  } catch (e: any) {
    error.value = e.message
  }
}

async function refreshAll() {
  loading.value = true
  error.value = ''
  await Promise.all([fetchSettings(), fetchStatus(), fetchReservations(), fetchUsers()])
  loading.value = false
}

onMounted(() => {
  refreshAll()
})
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-8 space-y-8">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white">VPN 管理</h2>
      <button
        class="px-4 py-2 rounded-lg font-medium text-sm text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
        @click="refreshAll"
      >
        刷新
      </button>
    </div>

    <p v-if="error" class="text-red-500">{{ error }}</p>

    <!-- 状态 -->
    <div class="grid gap-4 sm:grid-cols-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <p class="text-xs text-gray-500 dark:text-gray-400">服务状态</p>
        <p class="text-xl font-bold mt-1" :class="status?.enabled ? 'text-green-600' : 'text-gray-400'">
          {{ status?.enabled ? '运行中' : '已停止' }}
        </p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <p class="text-xs text-gray-500 dark:text-gray-400">在线客户端</p>
        <p class="text-xl font-bold text-gray-900 dark:text-white mt-1">{{ status?.online ?? '--' }}</p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <p class="text-xs text-gray-500 dark:text-gray-400">IP 用量</p>
        <p class="text-xl font-bold text-gray-900 dark:text-white mt-1">{{ status?.used_ips ?? '--' }} / {{ status?.capacity ?? '--' }}</p>
      </div>
    </div>

    <!-- 设置 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">隧道设置</h3>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">启用 VPN 服务</label>
          <select v-model="form.enabled" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
            <option :value="true">启用</option>
            <option :value="false">停止</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">子网 (CIDR)</label>
          <input v-model="form.subnet" type="text" placeholder="192.168.3.0/24" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">MTU</label>
          <input v-model.number="form.mtu" type="number" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">接口名 (留空自动)</label>
          <input v-model="form.interface_name" type="text" placeholder="tun0" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">允许客户端互通</label>
          <select v-model="form.allow_client_to_client" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
            <option :value="false">禁止</option>
            <option :value="true">允许</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">服务端配置 TUN IP</label>
          <select v-model="form.do_local_ip_config" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
            <option :value="true">自动配置</option>
            <option :value="false">手动</option>
          </select>
        </div>
      </div>
      <div class="flex items-center gap-4 mt-6">
        <button
          class="px-4 py-2 rounded-lg font-medium text-sm text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition-colors"
          :disabled="saving"
          @click="handleSave"
        >
          {{ saving ? '保存中...' : '保存设置' }}
        </button>
        <span v-if="saveMsg" :class="saveMsg === '保存成功' ? 'text-green-500' : 'text-red-500'" class="text-sm">{{ saveMsg }}</span>
      </div>
    </div>

    <!-- 在线客户端 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white p-6 pb-4">在线客户端</h3>
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">用户</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">分配 IP</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">连接时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!status?.clients?.length">
            <td colspan="3" class="px-6 py-6 text-center text-gray-400">暂无在线客户端</td>
          </tr>
          <tr v-for="(c, i) in status?.clients" :key="i" class="border-b border-gray-100 dark:border-gray-700/50">
            <td class="px-6 py-3 text-gray-900 dark:text-white font-medium">{{ c.username }}</td>
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ c.ip }}</td>
            <td class="px-6 py-3 text-gray-500 dark:text-gray-400">{{ c.connected_at }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 静态预留 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div class="flex items-center justify-between p-6 pb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">静态 IP 预留</h3>
        <button class="px-3 py-1.5 text-xs rounded-md font-medium text-white bg-sky-600 hover:bg-sky-700 transition-colors" @click="showAddResv = true">新增预留</button>
      </div>
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">用户</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">预留 IP</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">创建时间</th>
            <th class="px-6 py-3 text-left font-medium text-gray-500 dark:text-gray-400">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!reservations.length">
            <td colspan="4" class="px-6 py-6 text-center text-gray-400">暂无预留</td>
          </tr>
          <tr v-for="r in reservations" :key="r.id" class="border-b border-gray-100 dark:border-gray-700/50">
            <td class="px-6 py-3 text-gray-900 dark:text-white font-medium">{{ r.username }}</td>
            <td class="px-6 py-3 text-gray-700 dark:text-gray-300">{{ r.ip_address }}</td>
            <td class="px-6 py-3 text-gray-500 dark:text-gray-400">{{ r.created_at }}</td>
            <td class="px-6 py-3">
              <button class="px-3 py-1 text-xs rounded-md font-medium text-red-700 bg-red-50 hover:bg-red-100 dark:text-red-400 dark:bg-red-900/20 transition-colors" @click="handleDeleteResv(r.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增预留弹窗 -->
    <div v-if="showAddResv" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showAddResv = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">新增 IP 预留</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">用户</label>
            <select v-model.number="resvForm.user_id" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              <option :value="0" disabled>选择用户</option>
              <option v-for="u in users" :key="u.id" :value="u.id">{{ u.username }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">IP 地址</label>
            <input v-model="resvForm.ip_address" type="text" placeholder="192.168.3.10" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
          </div>
          <p v-if="resvError" class="text-sm text-red-500">{{ resvError }}</p>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors" @click="showAddResv = false">取消</button>
          <button class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 transition-colors" @click="handleAddResv">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>
