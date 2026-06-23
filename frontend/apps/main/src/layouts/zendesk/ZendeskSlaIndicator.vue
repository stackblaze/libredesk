<template>
  <div v-if="hasDeadlines" class="flex items-center gap-1.5">
    <SlaBadge
      v-show="nrdStatus === 'overdue' || nrdStatus === 'remaining'"
      :dueAt="conv.next_response_deadline_at"
      :actualAt="conv.next_response_met_at"
      :label="t('zendesk.slaNextReply')"
      :showExtra="false"
      @status="nrdStatus = $event"
      :key="`nrd-${conv.uuid}-${conv.next_response_deadline_at}`"
    />
    <SlaBadge
      v-show="frdStatus === 'overdue' || frdStatus === 'remaining'"
      :dueAt="conv.first_response_deadline_at"
      :actualAt="conv.first_reply_at"
      :label="t('zendesk.slaFirstReply')"
      :showExtra="false"
      @status="frdStatus = $event"
      :key="`frd-${conv.uuid}-${conv.first_response_deadline_at}`"
    />
    <SlaBadge
      v-show="rdStatus === 'overdue' || rdStatus === 'remaining'"
      :dueAt="conv.resolution_deadline_at"
      :actualAt="conv.resolved_at"
      :label="t('zendesk.slaResolution')"
      :showExtra="false"
      @status="rdStatus = $event"
      :key="`rd-${conv.uuid}-${conv.resolution_deadline_at}`"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import SlaBadge from '@main/features/sla/SlaBadge.vue'

const props = defineProps({
  conversation: { type: Object, default: null }
})

const { t } = useI18n()
const conv = computed(() => props.conversation || {})
const frdStatus = ref('')
const rdStatus = ref('')
const nrdStatus = ref('')

const hasDeadlines = computed(() => {
  const c = conv.value
  return !!(c.first_response_deadline_at || c.resolution_deadline_at || c.next_response_deadline_at)
})
</script>
