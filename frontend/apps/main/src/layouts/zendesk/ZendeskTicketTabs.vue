<template>
  <div class="zendesk-tab-bar flex items-end shrink-0 min-w-0">
    <div class="flex items-end min-w-0 overflow-x-auto">
      <div
        v-for="tab in tabs"
        :key="tab.uuid"
        class="zendesk-tab group"
        :class="{ active: tab.uuid === activeUuid }"
      >
        <button type="button" class="flex items-center gap-2 min-w-0 flex-1 text-left" @click="selectTab(tab)">
          <span class="relative shrink-0">
            <component
              :is="tab.inbox_channel === 'livechat' ? MessageSquare : Mail"
              class="size-4 text-muted-foreground"
              aria-hidden="true"
            />
            <span
              v-if="tab.priority"
              class="zendesk-priority-dot zendesk-tab-priority"
              :class="priorityDotClass(tab.priority)"
            />
          </span>
          <span class="flex flex-col min-w-0 flex-1 leading-tight">
            <span class="truncate text-xs font-medium">
              {{ tab.subject || t('zendesk.noSubject') }}
            </span>
            <span v-if="tab.reference_number" class="text-[11px] text-muted-foreground truncate">
              #{{ tab.reference_number }}
            </span>
          </span>
        </button>
        <button
          type="button"
          class="zendesk-tab-close shrink-0 opacity-0 group-hover:opacity-100 focus:opacity-100"
          :aria-label="t('zendesk.closeTab')"
          @click="closeTab(tab.uuid)"
        >
          <X class="size-3.5" />
        </button>
      </div>
    </div>
    <button type="button" class="zendesk-tab-add shrink-0" @click="addTab">
      <Plus class="size-3.5 mr-1" />
      {{ t('zendesk.addTab') }}
    </button>
  </div>
</template>

<script setup>
import { Mail, MessageSquare, X, Plus } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { useZendeskTabs } from '@main/composables/useZendeskTabs'
import { priorityDotClass } from '@main/composables/useConversationPriority'

const { t } = useI18n()
const { tabs, activeUuid, selectTab, closeTab, addTab } = useZendeskTabs()
</script>
