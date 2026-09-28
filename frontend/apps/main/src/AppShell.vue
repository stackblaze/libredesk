<template>
  <ZendeskApp />
</template>

<script setup>
import { useZendeskViewport } from '@main/composables/useZendeskViewport'
import { useNotificationBadge } from '@main/composables/useNotificationBadge'
import ZendeskApp from '@main/layouts/zendesk/ZendeskApp.vue'

useZendeskViewport()
useNotificationBadge()
</script>
