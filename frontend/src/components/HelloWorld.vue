<script setup lang="ts">
defineProps<{
  msg: string
}>()
</script>

<template>
  <div class="py-16 sm:py-24 text-center">
    <div class="max-w-3xl mx-auto px-4">
      <h1 class="text-4xl sm:text-5xl font-extrabold text-sky-600 dark:text-sky-400 mb-4">
        {{ msg }}
      </h1>
      <p class="text-lg sm:text-xl text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
        安全、快速、可靠的 VPN 隧道管理系统
      </p>
    </div>
  </div>
</template>
