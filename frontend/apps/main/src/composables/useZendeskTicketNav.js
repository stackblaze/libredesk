import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useConversationStore } from '@main/stores/conversation'

export function useZendeskTicketNav () {
  const route = useRoute()
  const router = useRouter()
  const conversationStore = useConversationStore()

  const conversationRouteName = computed(() => {
    if (route.params.teamID) return 'team-inbox-conversation'
    if (route.params.viewID) return 'view-inbox-conversation'
    return 'inbox-conversation'
  })

  const currentIndex = computed(() =>
    conversationStore.conversationsList.findIndex((c) => c.uuid === route.params.uuid)
  )

  const hasNext = computed(() => {
    const idx = currentIndex.value
    return idx >= 0 && idx < conversationStore.conversationsList.length - 1
  })

  const goToNext = () => {
    const idx = currentIndex.value
    if (idx < 0 || idx >= conversationStore.conversationsList.length - 1) return
    const next = conversationStore.conversationsList[idx + 1]
    router.push({
      name: conversationRouteName.value,
      params: {
        uuid: next.uuid,
        ...(route.params.teamID && { teamID: route.params.teamID }),
        ...(route.params.viewID && { viewID: route.params.viewID }),
        ...(conversationRouteName.value === 'inbox-conversation' && { type: route.params.type || 'assigned' })
      }
    })
  }

  return { hasNext, goToNext }
}
