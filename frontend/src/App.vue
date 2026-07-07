<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { toggleLocale } from '@/i18n'
import logo from '@/assets/logo.svg'

const router = useRouter()
const authStore = useAuthStore()
const { t, locale } = useI18n()

function handleLogout() {
  authStore.logout()
  router.push('/')
}

const navLinks = [
  { to: '/', label: 'nav.home' },
  { to: '/about', label: 'nav.about' },
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
      </div>
    </footer>
  </div>
</template>
