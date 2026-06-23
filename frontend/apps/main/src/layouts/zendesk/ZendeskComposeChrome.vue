<template>
  <div class="zendesk-ticket-chrome shrink-0">
    <div class="zendesk-ticket-breadcrumb flex items-stretch justify-between shrink-0">
      <nav class="zendesk-breadcrumb-nav min-w-0" aria-label="breadcrumb">
        <router-link :to="listRoute" class="zendesk-breadcrumb-segment zendesk-breadcrumb-link">
          {{ viewLabel }}
        </router-link>
        <span class="zendesk-breadcrumb-segment zendesk-breadcrumb-current truncate max-w-[20rem]">
          {{ t('conversation.newConversation') }}
        </span>
      </nav>

      <div class="zendesk-breadcrumb-meta flex items-stretch shrink-0">
        <div class="zendesk-panel-toggle">
          <button
            type="button"
            class="zendesk-panel-toggle-btn active"
            disabled
          >
            {{ t('zendesk.user') }}
          </button>
          <button
            type="button"
            class="zendesk-panel-toggle-btn"
            disabled
          >
            {{ t('zendesk.apps') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { useInboxViewContext } from '@main/composables/useInboxViewContext'

defineProps({
  composeId: { type: String, default: '' }
})

const { t } = useI18n()
const { viewLabel, listRoute } = useInboxViewContext()
</script>
