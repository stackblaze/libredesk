<template>
  <div class="p-4 md:p-6 space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-lg font-semibold">{{ t('report.tickets') }}</h1>
      <div class="flex flex-wrap items-center gap-2">
        <ReportFilters v-model:team-id="teamId" v-model:priority-id="priorityId" />
        <ReportRangePicker v-model="days" />
      </div>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="text-sm text-muted-foreground">{{ t('report.created') }}</p>
        <p class="mt-2 text-3xl font-semibold tabular-nums">{{ report.created || 0 }}</p>
      </div>
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="text-sm text-muted-foreground">{{ t('report.resolved') }}</p>
        <p class="mt-2 text-3xl font-semibold tabular-nums">{{ report.resolved || 0 }}</p>
      </div>
    </div>

    <div class="rounded-lg border bg-card p-5 shadow-xs">
      <p class="mb-3 text-sm font-medium text-muted-foreground">{{ t('report.byChannel') }}</p>
      <div v-if="report.by_channel?.length" class="space-y-2">
        <div
          v-for="row in report.by_channel"
          :key="row.channel"
          class="flex items-center justify-between text-sm"
        >
          <span class="capitalize">{{ row.channel }}</span>
          <span class="tabular-nums text-muted-foreground">
            {{ row.created }} {{ t('report.created').toLowerCase() }} · {{ row.resolved }}
            {{ t('report.resolved').toLowerCase() }}
          </span>
        </div>
      </div>
      <p v-else class="text-sm text-muted-foreground">{{ t('report.empty') }}</p>
    </div>

    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="mb-3 text-sm font-medium text-muted-foreground">{{ t('report.byStatus') }}</p>
        <div v-if="report.by_status?.length" class="space-y-2">
          <div v-for="row in report.by_status" :key="row.status" class="flex items-center justify-between text-sm">
            <span>{{ row.status }}</span>
            <span class="tabular-nums text-muted-foreground">{{ row.created }}</span>
          </div>
        </div>
        <p v-else class="text-sm text-muted-foreground">{{ t('report.empty') }}</p>
      </div>
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="mb-3 text-sm font-medium text-muted-foreground">{{ t('report.byPriority') }}</p>
        <div v-if="report.by_priority?.length" class="space-y-2">
          <div v-for="row in report.by_priority" :key="row.priority" class="flex items-center justify-between text-sm">
            <span>{{ row.priority }}</span>
            <span class="tabular-nums text-muted-foreground">{{ row.created }}</span>
          </div>
        </div>
        <p v-else class="text-sm text-muted-foreground">{{ t('report.empty') }}</p>
      </div>
    </div>

    <div class="rounded-lg border bg-card p-5 shadow-xs">
      <p class="mb-4 text-sm font-medium text-muted-foreground">{{ t('report.chart.title') }}</p>
      <LineChart v-if="chartData.length" :data="chartData" />
      <p v-else class="py-10 text-center text-sm text-muted-foreground">{{ t('report.empty') }}</p>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api'
import LineChart from '@/features/reports/OverviewLineChart.vue'
import ReportRangePicker from '@/features/reports/ReportRangePicker.vue'
import ReportFilters from '@/features/reports/ReportFilters.vue'
import { reportParams } from '@/features/reports/reportParams.js'
import { useEmitter } from '@/composables/useEmitter'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'

const { t } = useI18n()
const emitter = useEmitter()
const days = ref(30)
const teamId = ref(0)
const priorityId = ref(0)
const report = ref({ created: 0, resolved: 0, by_channel: [], by_status: [], by_priority: [], new_conversations: [], resolved_conversations: [] })

const chartData = computed(() => {
  const created = report.value.new_conversations || []
  const resolved = report.value.resolved_conversations || []
  const dateMap = new Map()
  created.forEach((item) => {
    dateMap.set(item.date, {
      date: item.date,
      [t('report.chart.newConversations')]: item.count,
      [t('report.chart.resolvedConversations')]: 0
    })
  })
  resolved.forEach((item) => {
    const existing = dateMap.get(item.date)
    if (existing) existing[t('report.chart.resolvedConversations')] = item.count
    else {
      dateMap.set(item.date, {
        date: item.date,
        [t('report.chart.newConversations')]: 0,
        [t('report.chart.resolvedConversations')]: item.count
      })
    }
  })
  return Array.from(dateMap.values()).sort((a, b) => new Date(a.date) - new Date(b.date))
})

async function load() {
  try {
    const resp = await api.getTicketReports(reportParams(days.value, teamId.value, priorityId.value))
    report.value = resp.data.data || report.value
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  }
}

onMounted(load)
watch([days, teamId, priorityId], load)
</script>
