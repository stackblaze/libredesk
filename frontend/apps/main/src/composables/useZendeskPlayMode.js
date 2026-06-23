import { useStorage } from '@vueuse/core'
import { useRoute, useRouter } from 'vue-router'
import { useConversationStore } from '@main/stores/conversation'

// Module-scoped so all components share the same play state / preference.
const openNext = useStorage('libredesk_zendesk_open_next', false)
const playActive = useStorage('libredesk_zendesk_play_active', false)

/** Zendesk-style "Play" mode: work through a view's tickets one after another. */
export function useZendeskPlayMode () {
  const route = useRoute()
  const router = useRouter()
  const conversationStore = useConversationStore()

  const conversationRouteName = () => {
    if (route.params.teamID) return 'team-inbox-conversation'
    if (route.params.viewID) return 'view-inbox-conversation'
    return 'inbox-conversation'
  }

  const paramsFor = (uuid) => ({
    uuid,
    ...(route.params.teamID && { teamID: route.params.teamID }),
    ...(route.params.viewID && { viewID: route.params.viewID }),
    ...(!route.params.teamID && !route.params.viewID && { type: route.params.type || 'assigned' })
  })

  const listRoute = () => {
    if (route.params.teamID) return { name: 'team-inbox', params: { teamID: route.params.teamID } }
    if (route.params.viewID) return { name: 'view-inbox', params: { viewID: route.params.viewID } }
    return { name: 'inbox', params: { type: route.params.type || 'assigned' } }
  }

  const startPlay = () => {
    const first = conversationStore.conversationsList[0]
    if (!first) return
    playActive.value = true
    openNext.value = true
    router.push({ name: conversationRouteName(), params: paramsFor(first.uuid) })
  }

  const stopPlay = () => {
    playActive.value = false
  }

  const exitToList = () => {
    playActive.value = false
    router.push(listRoute())
  }

  return { openNext, playActive, startPlay, stopPlay, exitToList }
}
