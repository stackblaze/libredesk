<template>
  <div class="flex flex-col h-full w-full min-h-0 min-w-0 bg-background">
    <div
      v-if="isLoading"
      class="h-0.5 shrink-0 bg-primary/40 animate-pulse"
    />

    <div v-if="showContent" class="flex flex-1 min-h-0 min-w-0 w-full">
      <ZendeskTicketProperties />
      <div class="flex flex-col flex-1 min-w-0 w-full border-r">
        <div class="px-4 py-2.5 border-b shrink-0 bg-background">
          <h2 class="text-base font-semibold truncate leading-snug">
            {{ conversationStore.current?.subject || t('zendesk.noSubject') }}
          </h2>
          <p
            v-if="conversationStore.current?.inbox_name"
            class="text-xs text-muted-foreground mt-0.5"
          >
            {{ t('zendesk.viaInbox', { inbox: conversationStore.current.inbox_name }) }}
          </p>
        </div>
        <div class="flex flex-col flex-1 min-h-0 overflow-hidden">
          <MessageList class="flex-1 overflow-y-auto" />
          <ReplyBox />
        </div>
      </div>
      <ZendeskCustomerContext />
    </div>
  </div>
</template>

<script setup>
import { computed, watch, onMounted } from 'vue'
import { useDocumentVisibility } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useConversationStore } from '@main/stores/conversation'
import MessageList from '@/features/conversation/message/MessageList.vue'
import ReplyBox from '@/features/conversation/ReplyBox.vue'
import ZendeskTicketProperties from './ZendeskTicketProperties.vue'
import ZendeskCustomerContext from './ZendeskCustomerContext.vue'

const props = defineProps({
  uuid: String
})

const { t } = useI18n()
const conversationStore = useConversationStore()

const showContent = computed(
  () => conversationStore.current || conversationStore.conversation.loading
)

const isLoading = computed(
  () => conversationStore.conversation.loading || conversationStore.messages.loading
)

const fetchConversation = async (uuid) => {
  await Promise.all([
    conversationStore.fetchConversation(uuid),
    conversationStore.fetchMessages(uuid)
  ])
  await conversationStore.updateAssigneeLastSeen(uuid)
}

onMounted(() => {
  if (props.uuid) fetchConversation(props.uuid)
})

watch(
  () => props.uuid,
  (newUUID, oldUUID) => {
    if (!newUUID || newUUID === oldUUID) return
    fetchConversation(newUUID)
  }
)

const visibility = useDocumentVisibility()
watch(visibility, (state) => {
  if (state === 'visible' && props.uuid) {
    conversationStore.updateAssigneeLastSeen(props.uuid)
  }
})
</script>
