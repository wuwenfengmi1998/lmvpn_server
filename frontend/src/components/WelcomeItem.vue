<template>
  <div class="flex gap-4 p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700
              hover:shadow-md transition-shadow">
    <div class="flex-shrink-0 w-12 h-12 flex items-center justify-center rounded-lg
                bg-sky-100 dark:bg-sky-900/40 text-sky-600 dark:text-sky-400">
      <slot name="icon"></slot>
    </div>
    <div class="flex-1 min-w-0">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-1">
        <slot name="heading"></slot>
      </h3>
      <div class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
        <slot></slot>
      </div>
    </div>
  </div>
</template>
