<template>
  <button
    v-if="collapsed"
    type="button"
    class="zendesk-ticket-context-collapsed shrink-0 border-l bg-background flex items-start justify-center pt-3 hover:bg-muted/50 min-h-0"
    :title="t('conversation.sidebar.previousConvo')"
    @click="toggle"
  >
    <PanelRightOpen class="size-4 text-muted-foreground" />
  </button>
  <aside v-else class="zendesk-ticket-context p-4 min-w-0 min-h-0 overflow-y-auto bg-background border-l">
    <ConversationSideBarContact />
    <Accordion type="multiple" collapsible class="mt-4">
      <AccordionItem value="previous">
        <AccordionTrigger class="text-sm py-2">{{ t('conversation.sidebar.previousConvo') }}</AccordionTrigger>
        <AccordionContent>
          <PreviousConversations />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  </aside>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useStorage } from '@vueuse/core'
import { PanelRightOpen } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger
} from '@shared-ui/components/ui/accordion'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'
import ConversationSideBarContact from '@/features/conversation/sidebar/ConversationSideBarContact.vue'
import PreviousConversations from '@/features/conversation/sidebar/PreviousConversations.vue'

const { t } = useI18n()
const emitter = useEmitter()

// Persisted so the pane stays collapsed/expanded across tickets and reloads.
const collapsed = useStorage('libredesk_zendesk_context_collapsed', false)

const toggle = () => {
  collapsed.value = !collapsed.value
}

onMounted(() => emitter.on(EMITTER_EVENTS.CONVERSATION_SIDEBAR_TOGGLE, toggle))
onUnmounted(() => emitter.off(EMITTER_EVENTS.CONVERSATION_SIDEBAR_TOGGLE, toggle))
</script>
