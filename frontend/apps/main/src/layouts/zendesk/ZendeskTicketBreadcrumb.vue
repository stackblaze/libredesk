<template>
  <div class="zendesk-ticket-breadcrumb flex items-stretch justify-between shrink-0">
    <nav class="flex items-stretch min-w-0 h-full text-sm" aria-label="breadcrumb">
      <router-link :to="listRoute" class="zendesk-breadcrumb-segment zendesk-breadcrumb-link">
        {{ viewLabel }}
      </router-link>
      <span class="zendesk-breadcrumb-segment truncate max-w-[14rem]">
        {{ requesterName }}
      </span>
      <span class="zendesk-breadcrumb-segment flex items-center gap-2 min-w-0">
        <span
          v-if="status"
          class="zendesk-status-badge shrink-0"
          :class="statusClass"
        >
          {{ status }}
        </span>
        <span v-if="ticketNumber" class="text-muted-foreground shrink-0 whitespace-nowrap">
          {{ t('zendesk.ticketNumber', { number: ticketNumber }) }}
        </span>
      </span>
    </nav>
    <div class="flex items-center px-3 shrink-0 border-l" style="border-color: hsl(var(--zendesk-border))">
      <Button variant="ghost" size="sm" class="h-7 text-xs" @click="emit('next')" :disabled="!hasNext">
        {{ t('zendesk.next') }}
        <ArrowRight class="size-3.5 ml-1" />
      </Button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ArrowRight } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { Button } from '@shared-ui/components/ui/button'
import { useInboxViewContext } from '@main/composables/useInboxViewContext'
import { useStatusCategory } from '@main/composables/useStatusCategory'

const props = defineProps({
  requesterName: { type: String, default: '' },
  status: { type: String, default: '' },
  ticketNumber: { type: [String, Number], default: '' },
  hasNext: { type: Boolean, default: false }
})

const emit = defineEmits(['next'])

const { t } = useI18n()
const { viewLabel, listRoute } = useInboxViewContext()
const { categoryClass } = useStatusCategory()

const statusClass = computed(() => categoryClass(props.status))
</script>
