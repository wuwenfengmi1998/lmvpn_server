<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { toggleLocale } from '@/i18n'
import logo from '@/assets/logo.svg'

const router = useRouter()
const authStore = useAuthStore()
const { t, locale } = useI18n()

const versionInfo = ref<{ commit: string; commitTime: string } | null>(null)

onMounted(async () => {
  try {
    const res = await fetch('/api/version')
    if (res.ok) {
      const data = await res.json()
      versionInfo.value = { commit: data.commit, commitTime: data.commit_time }
    }
  } catch {
    // 版本信息获取失败时静默处理，footer 仅显示版权
  }
})

function handleLogout() {
  authStore.logout()
  router.push('/')
}

const navLinks = [
  { to: '/', label: 'nav.home' },
]
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50 dark:bg-gray-950">
    <header class="bg-sky-600 text-white shadow-md">
      <div class="max-w-6xl mx-auto px-4 py-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <RouterLink to="/" class="flex items-center gap-2 text-xl font-bold tracking-wide hover:text-sky-200 transition-colors">
          <img :src="logo" alt="LmVPN" class="h-8 w-8 rounded-lg" />
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
            {{ t(link.label) }}
          </RouterLink>
          <a
            href="https://github.com/wuwenfengmi1998/lmvpn_server"
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors hover:bg-sky-500/40"
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-4 h-4">
              <path d="M12 .5C5.37.5 0 5.78 0 12.292c0 5.211 3.438 9.63 8.205 11.188.6.111.82-.254.82-.567 0-.28-.01-1.022-.015-2.005-3.338.711-4.042-1.582-4.042-1.582-.546-1.361-1.335-1.725-1.335-1.725-1.087-.731.084-.716.084-.716 1.205.082 1.838 1.215 1.838 1.215 1.07 1.803 2.809 1.282 3.495.981.108-.763.417-1.282.76-1.577-2.665-.295-5.466-1.309-5.466-5.827 0-1.287.465-2.339 1.235-3.164-.135-.298-.54-1.497.105-3.121 0 0 1.005-.316 3.3 1.209.96-.262 1.98-.392 3-.398 1.02.006 2.04.136 3 .398 2.28-1.525 3.285-1.209 3.285-1.209.645 1.624.24 2.823.12 3.121.765.825 1.23 1.877 1.23 3.164 0 4.53-2.805 5.527-5.475 5.817.42.354.81 1.077.81 2.182 0 1.578-.015 2.846-.015 3.229 0 .315.21.687.825.57C20.565 21.917 24 17.495 24 12.292 24 5.78 18.627.5 12 .5z" />
            </svg>
            GitHub
          </a>
          <template v-if="authStore.isLoggedIn">
            <RouterLink
              to="/profile"
              class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                     hover:bg-sky-500/40"
              active-class="bg-sky-700/60"
            >
              {{ t('nav.profile') }}
            </RouterLink>
            <template v-if="authStore.user?.role === 'admin'">
              <RouterLink
                to="/admin"
                class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                       hover:bg-sky-500/40"
                active-class="bg-sky-700/60"
              >
                {{ t('nav.admin') }}
              </RouterLink>
            </template>
            <a
              href="#"
              class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                     hover:bg-red-500/40"
              @click.prevent="handleLogout"
            >
              {{ t('nav.logout') }}
            </a>
          </template>
          <template v-else>
            <RouterLink
              to="/login"
              class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                     hover:bg-sky-500/40"
              active-class="bg-sky-700/60"
            >
              {{ t('nav.login') }}
            </RouterLink>
          </template>
          <button
            class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors
                   hover:bg-sky-500/40 border border-white/20"
            @click="toggleLocale"
          >
            {{ locale === 'zh' ? 'EN' : '中' }}
          </button>
        </nav>
      </div>
    </header>

    <main class="flex-1">
      <RouterView />
    </main>

    <footer class="bg-slate-800 text-slate-400 py-6 text-center text-sm">
      <div class="max-w-6xl mx-auto px-4">
        <p>&copy; {{ new Date().getFullYear() }} LmVPN. All rights reserved.</p>
        <p v-if="versionInfo" class="mt-1 text-slate-500">
          {{ t('footer.lastCommit') }}: {{ versionInfo.commitTime }} · {{ versionInfo.commit }}
        </p>
      </div>
    </footer>
  </div>
</template>
