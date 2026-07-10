<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const { t } = useI18n()

onMounted(async () => {
  await authStore.fetchUser()
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
