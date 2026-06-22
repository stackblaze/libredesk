<template>
  <div class="h-full w-full min-h-0 min-w-0 flex flex-col">
    <ConversationPlaceholder v-if="['inbox', 'team-inbox', 'view-inbox'].includes(route.name)" />
    <router-view class="flex-1 min-h-0 min-w-0 w-full" />
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { useInboxRouteSync } from '@/composables/useInboxRouteSync'
import ConversationPlaceholder from '@/features/conversation/ConversationPlaceholder.vue'

const route = useRoute()
useInboxRouteSync()
</script>
