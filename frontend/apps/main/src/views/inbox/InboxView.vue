<template>
  <ConversationPlaceholder v-if="['inbox', 'team-inbox', 'view-inbox'].includes(route.name)" />
  <router-view />
</template>

<script setup>
import { useRoute } from 'vue-router'
import { useInboxRouteSync } from '@/composables/useInboxRouteSync'
import ConversationPlaceholder from '@/features/conversation/ConversationPlaceholder.vue'

const route = useRoute()
useInboxRouteSync()
</script>
