<template>
  <aside class="w-72 shrink-0 border-l overflow-y-auto bg-background p-4">
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
import { useI18n } from 'vue-i18n'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger
} from '@shared-ui/components/ui/accordion'
import ConversationSideBarContact from '@/features/conversation/sidebar/ConversationSideBarContact.vue'
import PreviousConversations from '@/features/conversation/sidebar/PreviousConversations.vue'

const { t } = useI18n()
</script>
