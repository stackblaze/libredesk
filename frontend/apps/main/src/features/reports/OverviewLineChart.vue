<template>
  <LineChart
    :data="data"
    index="date"
    :categories="chartCategories"
    :x-formatter="xFormatter"
    :y-formatter="yFormatter"
  />
</template>

<script setup>
import { computed } from 'vue'
import { LineChart } from '@shared-ui/components/ui/chart-line'
import { useI18n } from 'vue-i18n'
const props = defineProps({
  data: {
    type: Array,
    default: () => []
  },
  categories: {
    type: Array,
    default: () => []
  }
})
const { t } = useI18n()
const chartCategories = computed(() =>
  props.categories.length
    ? props.categories
    : [t('report.chart.newConversations'), t('report.chart.resolvedConversations')]
)

const xFormatter = (tick) => {
  return props.data[tick]?.date ?? ''
}

const yFormatter = (tick) => {
  return Number.isInteger(tick) ? tick : ''
}
</script>
