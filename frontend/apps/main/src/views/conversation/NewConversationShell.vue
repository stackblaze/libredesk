<template>
  <ZendeskNewTicketWorkspace :compose-id="composeId" />
</template>

<script setup>
import ZendeskNewTicketWorkspace from '@main/layouts/zendesk/ZendeskNewTicketWorkspace.vue'

defineProps({
  composeId: String
})
</script>
