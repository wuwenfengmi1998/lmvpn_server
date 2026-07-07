<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const { t } = useI18n()
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
      throw new Error(data.error || t('common.loadFailed'))
    }
    const data = await res.json()
    users.value = data.users
  } catch (e: any) {
    error.value = e.message || t('common.loadFailed')
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
    createError.value = t('userManage.fillUsernamePassword')
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
      throw new Error(data.error || t('common.createFailed'))
    }
    showCreateModal.value = false
    createForm.value = { username: '', password: '', role: 'user', status: 1 }
    await fetchUsers()
  } catch (e: any) {
    createError.value = e.message || t('common.createFailed')
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
      throw new Error(data.error || t('common.saveFailed'))
    }
    showEditModal.value = false
    await fetchUsers()
  } catch (e: any) {
    editError.value = e.message || t('common.saveFailed')
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
      throw new Error(data.error || t('common.deleteFailed'))
    }
    showDeleteConfirm.value = false
    await fetchUsers()
  } catch (e: any) {
    deleteError.value = e.message || t('common.deleteFailed')
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
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('userManage.title') }}</h2>
      <button
        class="px-4 py-2 rounded-lg font-medium text-white bg-sky-600 hover:bg-sky-700 transition-colors"
        @click="showCreateModal = true"
      >
        {{ t('userManage.addUser') }}
      </button>
    </div>

    <p v-if="error" class="text-red-500 mb-4">{{ error }}</p>

    <div v-if="loading" class="text-gray-500 dark:text-gray-400 py-8 text-center">{{ t('common.loading') }}</div>

    <div v-else class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.id') }}</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.username') }}</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.role') }}</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.status') }}</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.createdAt') }}</th>
            <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.actions') }}</th>
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
                {{ user.role === 'admin' ? t('common.admin') : t('common.normalUser') }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span
                class="inline-block px-2 py-0.5 rounded-full text-xs font-medium"
                :class="user.status === 1 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'"
              >
                {{ user.status === 1 ? t('common.enabled') : t('common.disabled') }}
              </span>
            </td>
            <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ user.created_at }}</td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button
                  class="px-3 py-1 text-xs rounded-md font-medium text-sky-700 bg-sky-50 hover:bg-sky-100 dark:text-sky-400 dark:bg-sky-900/20 dark:hover:bg-sky-900/40 transition-colors"
                  @click="openEditModal(user)"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  class="px-3 py-1 text-xs rounded-md font-medium text-red-700 bg-red-50 hover:bg-red-100 dark:text-red-400 dark:bg-red-900/20 dark:hover:bg-red-900/40 transition-colors"
                  @click="confirmDelete(user)"
                >
                  {{ t('common.delete') }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showCreateModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('userManage.addUser') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('common.username') }}</label>
            <input
              v-model="createForm.username"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500 placeholder-gray-400 dark:placeholder-gray-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('common.password') }}</label>
            <input
              v-model="createForm.password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500 placeholder-gray-400 dark:placeholder-gray-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('common.role') }}</label>
            <select
              v-model="createForm.role"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option value="user">{{ t('common.normalUser') }}</option>
              <option value="admin">{{ t('common.admin') }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('common.status') }}</label>
            <select
              v-model.number="createForm.status"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option :value="1">{{ t('common.enabled') }}</option>
              <option :value="0">{{ t('common.disabled') }}</option>
            </select>
          </div>
          <p v-if="createError" class="text-sm text-red-500">{{ createError }}</p>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="showCreateModal = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition-colors"
            :disabled="creating"
            @click="handleCreate"
          >
            {{ creating ? t('userManage.creating') : t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showEditModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showEditModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ t('userManage.editUser', { username: editingUser?.username }) }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('common.role') }}</label>
            <select
              v-model="editForm.role"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option value="user">{{ t('common.normalUser') }}</option>
              <option value="admin">{{ t('common.admin') }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('common.status') }}</label>
            <select
              v-model.number="editForm.status"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option :value="1">{{ t('common.enabled') }}</option>
              <option :value="0">{{ t('common.disabled') }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('userManage.newPasswordOptional') }}</label>
            <input
              v-model="editForm.password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-sky-500 placeholder-gray-400 dark:placeholder-gray-500"
              :placeholder="t('userManage.newPasswordPlaceholder')"
            />
          </div>
          <p v-if="editError" class="text-sm text-red-500">{{ editError }}</p>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="showEditModal = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition-colors"
            :disabled="saving"
            @click="handleUpdate"
          >
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showDeleteConfirm = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-sm mx-4 p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">{{ t('userManage.confirmDelete') }}</h3>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
          {{ t('userManage.confirmDeleteMessage', { username: deletingUser?.username }) }}
        </p>
        <p v-if="deleteError" class="text-sm text-red-500 mb-4">{{ deleteError }}</p>
        <div class="flex justify-end gap-3">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="showDeleteConfirm = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-500 hover:bg-red-600 disabled:opacity-50 transition-colors"
            :disabled="deleting"
            @click="handleDelete"
          >
            {{ deleting ? t('common.deleting') : t('userManage.confirmDeleteButton') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
