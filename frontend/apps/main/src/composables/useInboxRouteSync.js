import { computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useIntervalFn, useDocumentVisibility } from '@vueuse/core'
import { useConversationStore } from '@/stores/conversation'
import { CONVERSATION_LIST_TYPE, CONVERSATION_DEFAULT_STATUSES } from '@/constants/conversation'

/** Keeps conversation list in sync with the current inbox / view / team route. */
export function useInboxRouteSync () {
  const route = useRoute()
  const conversationStore = useConversationStore()

  const type = computed(() => route.params.type)
  const teamID = computed(() => route.params.teamID)
  const viewID = computed(() => route.params.viewID)

  let lastFetchedKey = ''

  const storeHasCurrentList = () => {
    const c = conversationStore.conversations
    if (!c.initialized) return false
    if (viewID.value) {
      return c.listType === CONVERSATION_LIST_TYPE.VIEW && String(c.viewID) === String(viewID.value)
    }
    if (type.value) return c.listType === type.value
    if (teamID.value) {
      return c.listType === CONVERSATION_LIST_TYPE.TEAM_UNASSIGNED && String(c.teamID) === String(teamID.value)
    }
    return false
  }

  const fetchForCurrentRoute = () => {
    if (!type.value && !teamID.value && !viewID.value) return

    const key = `${type.value || ''}|${teamID.value || ''}|${viewID.value || ''}`
    if (key === lastFetchedKey) return
    lastFetchedKey = key

    if (storeHasCurrentList()) {
      conversationStore.refreshConversationList()
      return
    }

    if (viewID.value) {
      conversationStore.setListStatus('', false)
      conversationStore.fetchConversationsList(true, CONVERSATION_LIST_TYPE.VIEW, 0, [], viewID.value)
      return
    }

    if (!conversationStore.getListStatus) {
      conversationStore.setListStatus(CONVERSATION_DEFAULT_STATUSES.OPEN, false)
    }
    if (type.value) {
      conversationStore.fetchConversationsList(true, type.value)
    } else {
      conversationStore.fetchConversationsList(true, CONVERSATION_LIST_TYPE.TEAM_UNASSIGNED, teamID.value)
    }
  }

  onMounted(fetchForCurrentRoute)

  const visibility = useDocumentVisibility()
  const { pause, resume } = useIntervalFn(
    () => conversationStore.refreshConversationList(),
    120000
  )

  watch(visibility, (v) => {
    if (v === 'visible') {
      conversationStore.refreshConversationList()
      resume()
    } else {
      pause()
    }
  })

  watch([type, teamID, viewID], fetchForCurrentRoute)
}
