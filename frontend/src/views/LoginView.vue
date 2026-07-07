<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  if (!username.value || !password.value) {
    error.value = t('login.enterUsernamePassword')
    return
  }
  loading.value = true
  try {
    await authStore.login(username.value, password.value)
    router.push('/admin')
  } catch (e: any) {
    error.value = e.message || t('login.loginFailed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-[calc(100vh-160px)] flex items-center justify-center px-4 py-12">
    <div class="w-full max-w-sm bg-white dark:bg-gray-800 rounded-xl shadow-lg p-8">
      <h2 class="text-2xl font-bold text-center text-gray-900 dark:text-white mb-6">
        {{ t('login.title') }}
      </h2>
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            {{ t('common.username') }}
          </label>
          <input
            id="username"
            v-model="username"
            type="text"
            :placeholder="t('login.enterUsername')"
            autocomplete="username"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                   bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                   focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent
                   placeholder-gray-400 dark:placeholder-gray-500 transition-colors"
          />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            {{ t('common.password') }}
          </label>
          <input
            id="password"
            v-model="password"
            type="password"
            :placeholder="t('login.enterPassword')"
            autocomplete="current-password"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                   bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                   focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent
                   placeholder-gray-400 dark:placeholder-gray-500 transition-colors"
          />
        </div>
        <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
        <button
          type="submit"
          :disabled="loading"
          class="w-full py-2.5 rounded-lg font-medium text-white transition-colors
                 bg-sky-600 hover:bg-sky-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ loading ? t('login.loggingIn') : t('login.loginButton') }}
        </button>
      </form>
    </div>
  </div>
</template>
