<template>
  <div class="p-4 md:p-6 space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-lg font-semibold">{{ t('report.backlog') }}</h1>
      <div class="flex flex-wrap items-center gap-2">
        <ReportFilters v-model:team-id="teamId" v-model:priority-id="priorityId" />
        <ReportRangePicker v-model="days" />
      </div>
    </div>

    <div class="rounded-lg border bg-card p-5 shadow-xs">
      <p class="text-sm text-muted-foreground">{{ t('report.openNow') }}</p>
      <p class="mt-2 text-3xl font-semibold tabular-nums">{{ report.open || 0 }}</p>
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
          <span class="tabular-nums">{{ row.open }}</span>
        </div>
      </div>
      <p v-else class="text-sm text-muted-foreground">{{ t('report.empty') }}</p>
    </div>

    <div class="rounded-lg border bg-card p-5 shadow-xs">
      <p class="mb-4 text-sm font-medium text-muted-foreground">{{ t('report.chart.open') }}</p>
      <LineChart
        v-if="hasSeries"
        :data="chartData"
        :categories="[t('report.chart.open')]"
      />
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
const report = ref({ open: 0, by_channel: [], series: [] })

const hasSeries = computed(() => (report.value.series || []).some((item) => item.count > 0))
const chartData = computed(() =>
  (report.value.series || []).map((item) => ({
    date: item.date,
    [t('report.chart.open')]: item.count
  }))
)

async function load() {
  try {
    const resp = await api.getBacklogReports(reportParams(days.value, teamId.value, priorityId.value))
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
