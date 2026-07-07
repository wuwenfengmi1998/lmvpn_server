<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import logo from '@/assets/logo.svg'

const { t } = useI18n()
</script>

<template>
  <div class="py-16 sm:py-24 text-center">
    <div class="max-w-3xl mx-auto px-4">
      <img :src="logo" alt="LmVPN" class="h-20 w-20 rounded-2xl mx-auto mb-6 shadow-sm" />
      <h1 class="text-4xl sm:text-5xl font-extrabold text-sky-600 dark:text-sky-400 mb-4">
        LmVPN
      </h1>
      <p class="text-lg sm:text-xl text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
        {{ t('home.tagline') }}
      </p>
    </div>
  </div>
</template>
