<template>
  <div class="zendesk-ticket-chrome shrink-0">
    <ZendeskTicketTabs />
    <ZendeskTicketBreadcrumb
      :requester-name="conversationStore.currentContactName"
      :status="conversationStore.current?.status"
      :ticket-number="conversationStore.current?.reference_number"
      :has-next="hasNext"
      @next="goToNext"
    />
  </div>
</template>

<script setup>
import { useConversationStore } from '@main/stores/conversation'
import { useZendeskTicketNav } from '@main/composables/useZendeskTicketNav'
import ZendeskTicketTabs from './ZendeskTicketTabs.vue'
import ZendeskTicketBreadcrumb from './ZendeskTicketBreadcrumb.vue'

const conversationStore = useConversationStore()
const { hasNext, goToNext } = useZendeskTicketNav()
</script>
