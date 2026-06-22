<template>
  <div class="zendesk-app flex w-full min-w-0 h-screen min-h-0 overflow-hidden bg-[#f8f9f9]">
    <ZendeskNavRail />

    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <div
        v-if="!isInboxRoute"
        class="flex-1 min-h-0 m-2 rounded-lg overflow-hidden bg-background border"
      >
        <AdminBanner v-if="route.path.startsWith('/admin')" />
        <PageHeader v-if="!route.meta?.hidePageHeader" />
        <RouterView class="flex-grow" />
      </div>

      <div v-else class="flex-1 min-h-0 min-w-0 w-full flex flex-col">
        <RouterView class="h-full w-full" />
      </div>
    </div>
  </div>

  <Command />
  <CreateConversation v-model="openCreateConversationDialog" v-if="openCreateConversationDialog" />
  <ViewForm v-model:openDialog="openCreateViewForm" v-model:view="view" />
</template>

<script setup>
import { onMounted, ref, watch, computed } from 'vue'
import { useStorage } from '@vueuse/core'
import { RouterView, useRoute } from 'vue-router'
import { useUserStore } from '@main/stores/user'
import { initWS } from '@main/websocket.js'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'
import { useEmitter } from '@main/composables/useEmitter'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useConversationStore } from '@main/stores/conversation'
import { CONVERSATION_LIST_TYPE } from '@main/constants/conversation'
import { useInboxStore } from '@main/stores/inbox'
import { useUsersStore } from '@main/stores/users'
import { useTeamStore } from '@main/stores/team'
import { useSlaStore } from '@main/stores/sla'
import { useMacroStore } from '@main/stores/macro'
import { useSharedViewStore } from '@main/stores/sharedView'
import { useTagStore } from '@main/stores/tag'
import { useCustomAttributeStore } from '@main/stores/customAttributes'
import { useIdleDetection } from '@main/composables/useIdleDetection'
import { useNotificationStore } from '@main/stores/notification'
import { initAudioContext } from '@shared-ui/composables/useNotificationSound'
import PageHeader from '@main/components/layout/PageHeader.vue'
import ViewForm from '@/features/view/ViewForm.vue'
import AdminBanner from '@/components/banner/AdminBanner.vue'
import { toast as sooner } from 'vue-sonner'
import Command from '@/features/command/CommandBox.vue'
import CreateConversation from '@/features/conversation/CreateConversation.vue'
import ZendeskNavRail from './ZendeskNavRail.vue'
import { useI18n } from 'vue-i18n'
import api from '@main/api'

const route = useRoute()
const emitter = useEmitter()
const userStore = useUserStore()
const conversationStore = useConversationStore()
const usersStore = useUsersStore()
const teamStore = useTeamStore()
const inboxStore = useInboxStore()
const slaStore = useSlaStore()
const macroStore = useMacroStore()
const sharedViewStore = useSharedViewStore()
const tagStore = useTagStore()
const customAttributeStore = useCustomAttributeStore()
const notificationStore = useNotificationStore()
const { t } = useI18n()

const openCreateViewForm = ref(false)
const openCreateConversationDialog = ref(false)
const view = ref({})

const lastInboxPath = useStorage('lastInboxPath', '')
watch(
  () => route.path,
  (path) => {
    if (path.startsWith('/inboxes') && path !== '/inboxes/search') {
      lastInboxPath.value = path
    }
  },
  { immediate: true }
)

const isInboxRoute = computed(() => route.path.startsWith('/inboxes'))

initWS()
useIdleDetection()

const unlockAudio = () => {
  initAudioContext()
  document.removeEventListener('click', unlockAudio)
  document.removeEventListener('touchstart', unlockAudio)
}
document.addEventListener('click', unlockAudio)
document.addEventListener('touchstart', unlockAudio)

watch([() => notificationStore.unreadCount, () => route.fullPath], ([count]) => {
  const base = document.title.replace(/^\(\d+\)\s*/, '')
  document.title = count > 0 ? `(${count}) ${base}` : base
})

onMounted(() => {
  initToaster()
  listenViewRefresh()
  listenViewFormOpen()
  initStores()
})

const initStores = async () => {
  if (!userStore.userID) {
    await userStore.getCurrentUser()
  }
  await Promise.allSettled([
    sharedViewStore.loadSharedViews(),
    conversationStore.fetchStatuses(),
    conversationStore.fetchPriorities(),
    conversationStore.fetchAllDrafts(),
    usersStore.fetchUsers(),
    teamStore.fetchTeams(),
    inboxStore.fetchInboxes(),
    slaStore.fetchSlas(),
    macroStore.loadMacros(),
    tagStore.fetchTags(),
    customAttributeStore.fetchCustomAttributes()
  ])
}

const listenViewFormOpen = () => {
  emitter.on(EMITTER_EVENTS.OPEN_VIEW_FORM, () => {
    view.value = {}
    openCreateViewForm.value = true
  })
}

const initToaster = () => {
  emitter.on(EMITTER_EVENTS.SHOW_TOAST, (message) => {
    if (!message.description) return
    if (message.variant === 'destructive') {
      sooner.error(message.description)
    } else if (message.variant === 'warning') {
      sooner.warning(message.description)
    } else {
      sooner.success(message.description)
    }
  })
}

const listenViewRefresh = () => {
  emitter.on(EMITTER_EVENTS.REFRESH_LIST, refreshViews)
}

const refreshViews = async (data) => {
  openCreateViewForm.value = false
  if (data?.model === 'view') {
    const openID = route.params.viewID
    if (openID) {
      conversationStore.resetConversations()
      conversationStore.fetchConversationsList(true, CONVERSATION_LIST_TYPE.VIEW, 0, [], openID)
    }
  }
}
</script>
