<template>
  <div class="p-4 md:p-6 space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-lg font-semibold">{{ t('report.efficiency') }}</h1>
      <div class="flex flex-wrap items-center gap-2">
        <ReportFilters v-model:team-id="teamId" v-model:priority-id="priorityId" />
        <ReportRangePicker v-model="days" />
      </div>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="text-sm text-muted-foreground">{{ t('report.medianFirstReply') }}</p>
        <p class="mt-2 text-3xl font-semibold tabular-nums">{{ firstReply }}</p>
        <p class="mt-2 text-xs text-muted-foreground">
          {{ t('report.sampleCount', { count: report.first_reply_count || 0 }) }}
        </p>
      </div>
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="text-sm text-muted-foreground">{{ t('report.medianResolution') }}</p>
        <p class="mt-2 text-3xl font-semibold tabular-nums">{{ resolution }}</p>
        <p class="mt-2 text-xs text-muted-foreground">
          {{ t('report.sampleCount', { count: report.resolution_count || 0 }) }}
        </p>
      </div>
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="text-sm text-muted-foreground">{{ t('report.slaMet') }}</p>
        <p class="mt-2 text-3xl font-semibold tabular-nums">{{ report.sla_met || 0 }}</p>
      </div>
      <div class="rounded-lg border bg-card p-5 shadow-xs">
        <p class="text-sm text-muted-foreground">{{ t('report.slaBreached') }}</p>
        <p class="mt-2 text-3xl font-semibold tabular-nums">{{ report.sla_breached || 0 }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api'
import ReportRangePicker from '@/features/reports/ReportRangePicker.vue'
import ReportFilters from '@/features/reports/ReportFilters.vue'
import { reportParams } from '@/features/reports/reportParams.js'
import { formatDuration } from '@shared-ui/utils/datetime.js'
import { useEmitter } from '@/composables/useEmitter'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'

const { t } = useI18n()
const emitter = useEmitter()
const days = ref(30)
const teamId = ref(0)
const priorityId = ref(0)
const report = ref({
  median_first_reply_seconds: null,
  first_reply_count: 0,
  median_resolution_seconds: null,
  resolution_count: 0,
  sla_met: 0,
  sla_breached: 0
})

const firstReply = computed(() =>
  report.value.first_reply_count
    ? formatDuration(report.value.median_first_reply_seconds, false)
    : '—'
)
const resolution = computed(() =>
  report.value.resolution_count
    ? formatDuration(report.value.median_resolution_seconds, false)
    : '—'
)

async function load() {
  try {
    const resp = await api.getEfficiencyReports(reportParams(days.value, teamId.value, priorityId.value))
    report.value = { ...report.value, ...(resp.data.data || {}) }
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
