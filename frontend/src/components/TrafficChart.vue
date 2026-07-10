<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const { t } = useI18n()

interface TrafficRecord {
  date: string
  rx_bytes: number
  tx_bytes: number
}

const props = defineProps<{
  records: TrafficRecord[]
}>()

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}

const chartData = computed(() => ({
  labels: props.records.map(r => r.date.slice(5)),
  datasets: [
    {
      label: t('traffic.upload'),
      data: props.records.map(r => r.rx_bytes),
      backgroundColor: 'rgba(14, 165, 233, 0.6)',
      borderColor: 'rgba(14, 165, 233, 1)',
      borderWidth: 1,
    },
    {
      label: t('traffic.download'),
      data: props.records.map(r => r.tx_bytes),
      backgroundColor: 'rgba(34, 197, 94, 0.6)',
      borderColor: 'rgba(34, 197, 94, 1)',
      borderWidth: 1,
    },
  ],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top' as const,
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const label = context.dataset.label || ''
          return `${label}: ${formatBytes(context.parsed.y)}`
        },
      },
    },
  },
  scales: {
    y: {
      beginAtZero: true,
      ticks: {
        callback: (value: any) => formatBytes(Number(value)),
      },
    },
  },
}
</script>

<template>
  <div class="h-64">
    <Bar :data="chartData" :options="chartOptions" />
  </div>
</template>
