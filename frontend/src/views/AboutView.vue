<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-16 text-center">
    <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-4">{{ t('about.title') }}</h1>
    <p class="text-gray-600 dark:text-gray-400 leading-relaxed max-w-xl mx-auto">
      {{ t('about.description') }}
    </p>
  </div>
</template>
