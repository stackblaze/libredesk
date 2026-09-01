import { nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useConversationStore } from '@main/stores/conversation'
import { useUserStore } from '@main/stores/user'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'
import { conversationRouteForContext } from '@main/composables/useZendeskTabs'
import { useInboxViewContext } from '@main/composables/useInboxViewContext'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '@main/api'

export function useTicketActions () {
  const route = useRoute()
  const router = useRouter()
  const { t } = useI18n()
  const conversationStore = useConversationStore()
  const userStore = useUserStore()
  const emitter = useEmitter()

  const { listRoute } = useInboxViewContext()

  const toastError = (err) => {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(err).message
    })
  }

  const toastOk = (description) => {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { description })
  }

  const openTicket = (uuid) => {
    if (!uuid) return
    router.push(conversationRouteForContext(route, uuid))
  }

  const statusTargets = (uuid) => {
    if (conversationStore.isSelected(uuid) && conversationStore.selectedCount > 1) {
      return [...conversationStore.selectedUUIDs]
    }
    return [uuid]
  }

  const setStatus = async (uuid, status) => {
    if (!uuid || !status) return
    const uuids = statusTargets(uuid)
    const results = await Promise.allSettled(
      uuids.map((id) => api.updateConversationStatus(id, { status }))
    )
    const ok = uuids.filter((_, i) => results[i].status === 'fulfilled')
    results.forEach((result) => {
      if (result.status === 'rejected') toastError(result.reason)
    })
    for (const id of ok) {
      conversationStore.mergeConversationUpdate({ uuid: id, status })
    }
    if (ok.length) {
      toastOk(
        uuids.length > 1
          ? t('conversation.bulkActions.statusSuccess', { count: ok.length, status })
          : t('conversation.ticketActions.statusSet', { status })
      )
    }
    if (uuids.length > 1) {
      conversationStore.clearSelection()
      conversationStore.fetchFirstPageConversations()
    }
  }

  const assignToMe = async (uuid) => {
    if (!uuid || !userStore.userID) return
    try {
      await api.updateAssignee(uuid, 'user', { assignee_id: userStore.userID })
      conversationStore.mergeConversationUpdate({ uuid, assigned_user_id: userStore.userID })
      toastOk(t('actions.assignToMe'))
    } catch (err) {
      toastError(err)
    }
  }

  const markUnread = (uuid) => {
    if (!uuid) return
    conversationStore.markAsUnread(uuid)
  }

  const toggleSelect = (uuid) => {
    if (!uuid) return
    conversationStore.toggleSelect(uuid, false)
  }

  const openMerge = (uuid) => {
    emitter.emit(EMITTER_EVENTS.OPEN_MERGE_DIALOG, { uuid })
  }

  const createLinked = async (uuid, origin) => {
    if (!uuid) return
    try {
      const req =
        origin === 'follow_up'
          ? api.createFollowUpConversation(uuid, {})
          : api.createChildConversation(uuid, {})
      const { data } = await req
      const created = data.data
      toastOk(
        origin === 'follow_up'
          ? t('conversation.related.followUpCreated')
          : t('conversation.related.childCreated')
      )
      if (created?.uuid) openTicket(created.uuid)
    } catch (err) {
      toastError(err)
    }
  }

  const focusComposer = async (uuid, type) => {
    if (!uuid) return
    if (route.params.uuid !== uuid) {
      await router.push(conversationRouteForContext(route, uuid))
      await nextTick()
      setTimeout(() => {
        emitter.emit(EMITTER_EVENTS.FOCUS_COMPOSER, { type })
      }, 200)
      return
    }
    emitter.emit(EMITTER_EVENTS.FOCUS_COMPOSER, { type })
  }

  const startSplit = async (uuid) => {
    if (!uuid) return
    if (route.params.uuid !== uuid) {
      await router.push(conversationRouteForContext(route, uuid))
      await nextTick()
      setTimeout(() => {
        emitter.emit(EMITTER_EVENTS.START_SPLIT_MODE)
      }, 200)
      return
    }
    emitter.emit(EMITTER_EVENTS.START_SPLIT_MODE)
  }

  const deleteTickets = async (uuid) => {
    if (!uuid) return
    const uuids = statusTargets(uuid)
    try {
      await conversationStore.deleteConversations(uuids)
      toastOk(
        uuids.length > 1
          ? t('conversation.spam.ticketsDeleted', uuids.length, { count: uuids.length })
          : t('conversation.spam.ticketDeleted')
      )
      if (uuids.includes(route.params.uuid)) {
        router.push(listRoute.value)
      }
    } catch (err) {
      toastError(err)
    }
  }

  return {
    openTicket,
    setStatus,
    assignToMe,
    markUnread,
    toggleSelect,
    openMerge,
    createLinked,
    focusComposer,
    startSplit,
    deleteTickets,
    statusTargets
  }
}
