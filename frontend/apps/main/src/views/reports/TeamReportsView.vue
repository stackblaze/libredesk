<template>
  <div class="space-y-4 p-4 md:p-6">
    <div class="flex items-center justify-between gap-3">
      <h1 class="text-lg font-semibold">{{ t('report.teams') }}</h1>
      <ReportRangePicker v-model="days" />
    </div>
    <DataTable :columns="columns" :data="rows" :loading="loading" />
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api'
import DataTable from '@main/components/datatable/DataTable.vue'
import ReportRangePicker from '@/features/reports/ReportRangePicker.vue'

const { t } = useI18n()
const days = ref(30)
const rows = ref([])
const loading = ref(true)

function formatSeconds(sec) {
  if (!sec) return '—'
  const m = Math.round(sec / 60)
  return `${m}m`
}

const columns = [
  {
    accessorKey: 'name',
    header: () => t('globals.terms.team')
  },
  {
    accessorKey: 'tickets_assigned',
    header: () => t('report.ticketsAssigned')
  },
  {
    accessorKey: 'tickets_resolved',
    header: () => t('report.ticketsResolved')
  },
  {
    accessorKey: 'replies',
    header: () => t('report.replies')
  },
  {
    accessorKey: 'avg_first_reply_seconds',
    header: () => t('report.avgFirstReply'),
    cell: ({ row }) => formatSeconds(row.original.avg_first_reply_seconds)
  },
  {
    accessorKey: 'csat_avg',
    header: () => t('globals.terms.csatRating'),
    cell: ({ row }) => Number(row.original.csat_avg || 0).toFixed(1)
  }
]

async function load() {
  try {
    loading.value = true
    const resp = await api.getTeamReports({ days: days.value })
    rows.value = resp.data.data || []
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(days, load)
</script>
