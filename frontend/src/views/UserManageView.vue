<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const users = ref<any[]>([])
const loading = ref(false)
const error = ref('')

interface User {
  id: number
  username: string
  role: string
  status: number
  created_at: string
  updated_at: string
}

async function fetchUsers() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch('/api/admin/users', {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || '加载失败')
    }
    const data = await res.json()
    users.value = data.users
  } catch (e: any) {
    error.value = e.message || '加载失败'
  } finally {
    loading.value = false
  }
}

const showCreateModal = ref(false)
const createForm = ref({ username: '', password: '', role: 'user', status: 1 })
const createError = ref('')
const creating = ref(false)

async function handleCreate() {
  createError.value = ''
  if (!createForm.value.username || !createForm.value.password) {
    createError.value = '请填写用户名和密码'
    return
  }
  creating.value = true
  try {
    const res = await fetch('/api/admin/users', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({
        username: createForm.value.username,
        password: createForm.value.password,
        role: createForm.value.role,
        status: createForm.value.status,
      }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || '创建失败')
    }
    showCreateModal.value = false
    createForm.value = { username: '', password: '', role: 'user', status: 1 }
    await fetchUsers()
  } catch (e: any) {
    createError.value = e.message || '创建失败'
  } finally {
    creating.value = false
  }
}

const showEditModal = ref(false)
const editingUser = ref<User | null>(null)
const editForm = ref({ role: 'user', status: 1, password: '' })
const editError = ref('')
const saving = ref(false)

function openEditModal(user: User) {
  editingUser.value = user
  editForm.value = { role: user.role, status: user.status, password: '' }
  editError.value = ''
  showEditModal.value = true
}

async function handleUpdate() {
  if (!editingUser.value) return
  editError.value = ''
  saving.value = true
  try {
    const body: any = { role: editForm.value.role, status: editForm.value.status }
    if (editForm.value.password) {
      body.password = editForm.value.password
    }
    const res = await fetch(`/api/admin/users/${editingUser.value.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify(body),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || '保存失败')
    }
    showEditModal.value = false
    await fetchUsers()
  } catch (e: any) {
    editError.value = e.message || '保存失败'
  } finally {
    saving.value = false
  }
}

const showDeleteConfirm = ref(false)
const deletingUser = ref<User | null>(null)
const deleteError = ref('')
const deleting = ref(false)

function confirmDelete(user: User) {
  deletingUser.value = user
  deleteError.value = ''
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (!deletingUser.value) return
  deleteError.value = ''
  deleting.value = true
  try {
    const res = await fetch(`/api/admin/users/${deletingUser.value.id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || '删除失败')
    }
    showDeleteConfirm.value = false
    await fetchUsers()
  } catch (e: any) {
    deleteError.value = e.message || '删除失败'
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-8">
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white">用户管理</h2>
      <button
        class="px-4 py-2 rounded-lg font-medium text-white bg-sky-600 hover:bg-sky-700 transition-colors"
        @click="showCreateModal = true"
      >
        新增用户
      </button>
    </div>

    <p v-if="error" class="text-red-500 mb-4">{{ error }}</p>

    <div v-if="loading" class="text-gray-500 dark:text-gray-400 py-8 text-center">加载中...</div>

    <div v-else class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">ID</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">用户名</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">角色</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">状态</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">创建时间</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="user in users"
            :key="user.id"
            class="border-b border-gray-100 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
          >
            <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ user.id }}</td>
            <td class="px-4 py-3 text-gray-900 dark:text-white font-medium">{{ user.username }}</td>
            <td class="px-4 py-3">
              <span
                class="inline-block px-2 py-0.5 rounded-full text-xs font-medium"
                :class="user.role === 'admin' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'"
              >
                {{ user.role === 'admin' ? '管理员' : '普通用户' }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span
                class="inline-block px-2 py-0.5 rounded-full text-xs font-medium"
                :class="user.status === 1 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'"
              >
                {{ user.status === 1 ? '启用' : '禁用' }}
              </span>
            </td>
            <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ user.created_at }}</td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button
                  class="px-3 py-1 text-xs rounded-md font-medium text-sky-700 bg-sky-50 hover:bg-sky-100 dark:text-sky-400 dark:bg-sky-900/20 dark:hover:bg-sky-900/40 transition-colors"
                  @click="openEditModal(user)"
                >
                  编辑
                </button>
                <button
                  class="px-3 py-1 text-xs rounded-md font-medium text-red-700 bg-red-50 hover:bg-red-100 dark:text-red-400 dark:bg-red-900/20 dark:hover:bg-red-900/40 transition-colors"
                  @click="confirmDelete(user)"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showCreateModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">新增用户</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">用户名</label>
            <input
              v-model="createForm.username"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500 placeholder-gray-400 dark:placeholder-gray-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">密码</label>
            <input
              v-model="createForm.password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500 placeholder-gray-400 dark:placeholder-gray-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">角色</label>
            <select
              v-model="createForm.role"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">状态</label>
            <select
              v-model.number="createForm.status"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option :value="1">启用</option>
              <option :value="0">禁用</option>
            </select>
          </div>
          <p v-if="createError" class="text-sm text-red-500">{{ createError }}</p>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="showCreateModal = false"
          >
            取消
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition-colors"
            :disabled="creating"
            @click="handleCreate"
          >
            {{ creating ? '创建中...' : '确定' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showEditModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showEditModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">编辑用户 - {{ editingUser?.username }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">角色</label>
            <select
              v-model="editForm.role"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">状态</label>
            <select
              v-model.number="editForm.status"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option :value="1">启用</option>
              <option :value="0">禁用</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">新密码（留空不修改）</label>
            <input
              v-model="editForm.password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500 placeholder-gray-400 dark:placeholder-gray-500"
              placeholder="留空则不修改密码"
            />
          </div>
          <p v-if="editError" class="text-sm text-red-500">{{ editError }}</p>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="showEditModal = false"
          >
            取消
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition-colors"
            :disabled="saving"
            @click="handleUpdate"
          >
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showDeleteConfirm = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-sm mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">确认删除</h3>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
          确定要删除用户 <span class="font-medium text-gray-900 dark:text-white">{{ deletingUser?.username }}</span> 吗？此操作不可撤销。
        </p>
        <p v-if="deleteError" class="text-sm text-red-500 mb-4">{{ deleteError }}</p>
        <div class="flex justify-end gap-3">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="showDeleteConfirm = false"
          >
            取消
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-500 hover:bg-red-600 disabled:opacity-50 transition-colors"
            :disabled="deleting"
            @click="handleDelete"
          >
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
