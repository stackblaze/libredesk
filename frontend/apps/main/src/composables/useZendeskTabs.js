import { computed, watch } from 'vue'
import { useStorage } from '@vueuse/core'
import { useRoute, useRouter } from 'vue-router'
import { useConversationStore } from '@main/stores/conversation'
import { useInboxViewContext } from '@main/composables/useInboxViewContext'

const MAX_TABS = 15

const CONVERSATION_ROUTE_NAMES = new Set([
  'inbox-conversation',
  'team-inbox-conversation',
  'view-inbox-conversation'
])

function stripHtml (value) {
  return (value || '').replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
}

function contactNameFrom (contact) {
  if (!contact) return ''
  return `${contact.first_name || ''} ${contact.last_name || ''}`.trim()
}

function messagePreviewFrom (listItem, current) {
  const raw = current?.last_message || listItem?.last_message || ''
  return stripHtml(raw)
}

export function useZendeskTabs () {
  const tabs = useStorage('libredesk_zendesk_tabs', [])
  const route = useRoute()
  const router = useRouter()
  const conversationStore = useConversationStore()
  const { listRoute } = useInboxViewContext()

  const activeUuid = computed(() => route.params.uuid || '')

  const buildTabFromRoute = () => {
    const uuid = route.params.uuid
    if (!uuid) return null

    const listItem = conversationStore.conversationsList.find((c) => c.uuid === uuid)
    const current = conversationStore.current?.uuid === uuid ? conversationStore.current : null
    const contact = current?.contact || listItem?.contact

    return {
      uuid,
      subject: current?.subject || listItem?.subject || '',
      reference_number: current?.reference_number || listItem?.reference_number || '',
      inbox_channel: current?.inbox_channel || listItem?.inbox_channel || '',
      priority: current?.priority || listItem?.priority || '',
      status: current?.status || listItem?.status || '',
      contact_name: contactNameFrom(contact) || conversationStore.getContactFullName(uuid) || '',
      message_preview: messagePreviewFrom(listItem, current),
      routeName: route.name,
      routeParams: { ...route.params }
    }
  }

  const syncActiveTab = () => {
    if (!CONVERSATION_ROUTE_NAMES.has(route.name)) return
    const tabData = buildTabFromRoute()
    if (!tabData) return

    const idx = tabs.value.findIndex((t) => t.uuid === tabData.uuid)
    if (idx >= 0) {
      tabs.value[idx] = { ...tabs.value[idx], ...tabData }
    } else {
      tabs.value.push(tabData)
      if (tabs.value.length > MAX_TABS) {
        tabs.value.shift()
      }
    }
  }

  watch(
    () => [
      route.name,
      route.params.uuid,
      conversationStore.current?.subject,
      conversationStore.current?.reference_number,
      conversationStore.current?.status,
      conversationStore.current?.priority,
      conversationStore.current?.inbox_channel,
      conversationStore.current?.contact?.first_name,
      conversationStore.current?.contact?.last_name,
      conversationStore.current?.last_message,
      conversationStore.conversationsList.find((c) => c.uuid === route.params.uuid)?.last_message
    ],
    syncActiveTab,
    { immediate: true }
  )

  const selectTab = (tab) => {
    if (tab.uuid === activeUuid.value) return
    router.push({ name: tab.routeName, params: tab.routeParams })
  }

  const closeTab = (uuid) => {
    const idx = tabs.value.findIndex((t) => t.uuid === uuid)
    if (idx === -1) return

    const wasActive = activeUuid.value === uuid
    tabs.value.splice(idx, 1)

    if (!wasActive) return

    if (tabs.value.length) {
      const nextTab = tabs.value[Math.min(idx, tabs.value.length - 1)]
      router.push({ name: nextTab.routeName, params: nextTab.routeParams })
    } else {
      router.push(listRoute.value)
    }
  }

  const addTab = () => {
    router.push(listRoute.value)
  }

  return { tabs, activeUuid, selectTab, closeTab, addTab }
}
