<template>
  <Teleport to="body">
    <div
      class="zendesk-row-preview"
      :style="positionStyle"
      role="dialog"
      aria-hidden="true"
    >
      <div class="flex items-center justify-between gap-2 mb-2">
        <span v-if="status" class="zendesk-status-badge shrink-0" :class="statusClass">
          {{ status }}
        </span>
        <span v-if="ticketNumber" class="text-xs text-muted-foreground tabular-nums">
          {{ t('zendesk.ticketNumber', { number: ticketNumber }) }}
        </span>
      </div>

      <p class="text-sm font-semibold text-foreground leading-snug mb-1 line-clamp-2">
        {{ conversation.subject || t('zendesk.noSubject') }}
      </p>

      <p v-if="requester" class="text-xs text-muted-foreground mb-2">
        {{ requester }}
      </p>

      <div class="border-t pt-2" style="border-color: hsl(var(--zendesk-border))">
        <p class="zendesk-eyebrow mb-1">{{ t('zendesk.latestComment') }}</p>
        <p class="text-sm text-foreground/90 leading-snug line-clamp-4">
          {{ latestComment || t('zendesk.noMessagePreview') }}
        </p>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useStatusCategory } from '@main/composables/useStatusCategory'

const props = defineProps({
  conversation: { type: Object, required: true },
  requester: { type: String, default: '' },
  // DOMRect of the hovered row, used to anchor the floating card.
  anchor: { type: Object, default: null }
})

const { t } = useI18n()
const { categoryClass } = useStatusCategory()

const status = computed(() => props.conversation.status)
const statusClass = computed(() => categoryClass(props.conversation.status))
const ticketNumber = computed(() => props.conversation.reference_number)
const latestComment = computed(() => {
  const raw = props.conversation.last_message || ''
  return raw.replace(/<[^>]*>/g, '').trim()
})

const PREVIEW_WIDTH = 360
const PREVIEW_MAX_HEIGHT = 220

const positionStyle = computed(() => {
  const rect = props.anchor
  if (!rect) return { display: 'none' }
  const margin = 12
  let left = rect.left + 48
  if (left + PREVIEW_WIDTH > window.innerWidth - margin) {
    left = Math.max(margin, window.innerWidth - PREVIEW_WIDTH - margin)
  }
  let top = rect.bottom + 4
  if (top + PREVIEW_MAX_HEIGHT > window.innerHeight - margin) {
    top = Math.max(margin, rect.top - PREVIEW_MAX_HEIGHT - 4)
  }
  return {
    left: `${left}px`,
    top: `${top}px`,
    width: `${PREVIEW_WIDTH}px`
  }
})
</script>
