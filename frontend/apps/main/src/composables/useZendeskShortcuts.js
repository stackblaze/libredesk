import { onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useConversationStore } from '@main/stores/conversation'

const CONVERSATION_ROUTES = new Set([
  'inbox-conversation',
  'team-inbox-conversation',
  'view-inbox-conversation'
])
const LIST_ROUTES = new Set(['inbox', 'team-inbox', 'view-inbox'])

function isTypingTarget (el) {
  if (!el) return false
  const tag = el.tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || el.isContentEditable
}

/**
 * Zendesk-style list/ticket keyboard shortcuts (active only in Zendesk layout):
 *  - j / k : open next / previous ticket in the current list
 *  - Enter : open the first ticket when viewing a list
 * Ctrl/Cmd+Enter to submit a reply is handled by the editor itself.
 */
export function useZendeskShortcuts () {
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

  const openIndex = (idx) => {
    const list = conversationStore.conversationsList
    if (idx < 0 || idx >= list.length) return
    router.push({ name: conversationRouteName(), params: paramsFor(list[idx].uuid) })
  }

  const currentIndex = () =>
    conversationStore.conversationsList.findIndex((c) => c.uuid === route.params.uuid)

  const onKeydown = (e) => {
    if (isTypingTarget(e.target)) return
    if (e.metaKey || e.ctrlKey || e.altKey) return

    const inConversation = CONVERSATION_ROUTES.has(route.name)
    const inList = LIST_ROUTES.has(route.name)
    if (!inConversation && !inList) return

    if (e.key === 'j' || e.key === 'k') {
      e.preventDefault()
      const idx = currentIndex()
      if (idx === -1) {
        if (e.key === 'j') openIndex(0)
        return
      }
      openIndex(e.key === 'j' ? idx + 1 : idx - 1)
    } else if (e.key === 'Enter' && inList) {
      e.preventDefault()
      openIndex(0)
    }
  }

  onMounted(() => window.addEventListener('keydown', onKeydown))
  onUnmounted(() => window.removeEventListener('keydown', onKeydown))
}
