<template>
  <ZendeskConversationWorkspace v-if="isZendesk" :uuid="uuid" />
  <ConversationDetailView v-else :uuid="uuid" />
</template>

<script setup>
import { computed } from 'vue'
import { useUiLayout, UI_LAYOUT_ZENDESK } from '@main/composables/useUiLayout'
import ConversationDetailView from '@main/views/conversation/ConversationDetailView.vue'
import ZendeskConversationWorkspace from '@main/layouts/zendesk/ZendeskConversationWorkspace.vue'

defineProps({
  uuid: String
})

const { layout } = useUiLayout()
const isZendesk = computed(() => layout.value === UI_LAYOUT_ZENDESK)
</script>
