<template>
  <div class="flex flex-col h-full bg-background">
    <!-- Ticket header -->
    <div class="flex items-center justify-between px-4 h-11 border-b shrink-0 text-sm">
      <div class="flex items-center gap-2 min-w-0">
        <Button variant="ghost" size="sm" class="shrink-0" @click="backToList">
          <ArrowLeft class="size-4 mr-1" />
          {{ t('zendesk.backToViews') }}
        </Button>
        <span class="text-muted-foreground">·</span>
        <span class="truncate font-medium">{{ conversationStore.currentContactName }}</span>
        <span
          v-if="conversationStore.current?.status"
          class="zendesk-status-badge shrink-0"
          :class="statusClass"
        >
          {{ conversationStore.current.status }}
        </span>
        <span v-if="conversationStore.current?.reference_number" class="text-muted-foreground shrink-0">
          #{{ conversationStore.current.reference_number }}
        </span>
      </div>
      <Button variant="outline" size="sm" @click="goToNext" :disabled="!hasNext">
        {{ t('zendesk.next') }}
        <ArrowRight class="size-4 ml-1" />
      </Button>
    </div>

    <div
      v-if="isLoading"
      class="h-0.5 shrink-0 bg-primary/40 animate-pulse"
    />

    <div v-if="showContent" class="flex flex-1 min-h-0">
      <ZendeskTicketProperties />
      <div class="flex flex-col flex-1 min-w-0 border-r">
        <div class="px-4 py-2 border-b shrink-0">
          <h2 class="text-base font-normal truncate">
            {{ conversationStore.current?.subject || t('zendesk.noSubject') }}
          </h2>
        </div>
        <div class="flex flex-col flex-1 min-h-0 overflow-hidden">
          <MessageList class="flex-1 overflow-y-auto" />
          <ReplyBox />
        </div>
      </div>
      <ZendeskCustomerContext />
    </div>

    <ZendeskWorkspaceFooter v-if="showContent" />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDocumentVisibility } from '@vueuse/core'
import { ArrowLeft, ArrowRight } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { useI18n } from 'vue-i18n'
import { useConversationStore } from '@main/stores/conversation'
import MessageList from '@/features/conversation/message/MessageList.vue'
import ReplyBox from '@/features/conversation/ReplyBox.vue'
import ZendeskTicketProperties from './ZendeskTicketProperties.vue'
import ZendeskCustomerContext from './ZendeskCustomerContext.vue'
import ZendeskWorkspaceFooter from './ZendeskWorkspaceFooter.vue'

const props = defineProps({
  uuid: String
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const conversationStore = useConversationStore()

const showContent = computed(
  () => conversationStore.current || conversationStore.conversation.loading
)

const isLoading = computed(
  () => conversationStore.conversation.loading || conversationStore.messages.loading
)

const statusClass = computed(() => {
  const s = (conversationStore.current?.status || '').toLowerCase()
  if (s === 'new') return 'status-new'
  if (s === 'open') return 'status-open'
  return 'status-default'
})

const listRoute = computed(() => {
  if (route.params.teamID) {
    return { name: 'team-inbox', params: { teamID: route.params.teamID } }
  }
  if (route.params.viewID) {
    return { name: 'view-inbox', params: { viewID: route.params.viewID } }
  }
  return { name: 'inbox', params: { type: route.params.type || 'assigned' } }
})

const currentIndex = computed(() =>
  conversationStore.conversationsList.findIndex((c) => c.uuid === props.uuid)
)

const hasNext = computed(() => {
  const idx = currentIndex.value
  return idx >= 0 && idx < conversationStore.conversationsList.length - 1
})

const backToList = () => {
  router.push(listRoute.value)
}

const goToNext = () => {
  const idx = currentIndex.value
  if (idx < 0 || idx >= conversationStore.conversationsList.length - 1) return
  const next = conversationStore.conversationsList[idx + 1]
  const baseRoute = route.params.teamID
    ? 'team-inbox-conversation'
    : route.params.viewID
      ? 'view-inbox-conversation'
      : 'inbox-conversation'
  router.push({
    name: baseRoute,
    params: {
      uuid: next.uuid,
      ...(route.params.teamID && { teamID: route.params.teamID }),
      ...(route.params.viewID && { viewID: route.params.viewID }),
      ...(baseRoute === 'inbox-conversation' && { type: route.params.type || 'assigned' })
    }
  })
}

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
