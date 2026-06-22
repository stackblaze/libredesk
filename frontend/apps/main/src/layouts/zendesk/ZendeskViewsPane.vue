<template>
  <aside class="zendesk-views-pane w-56 shrink-0 flex flex-col h-full overflow-hidden">
    <div class="flex items-center justify-between px-3 py-3 border-b shrink-0">
      <span class="font-semibold text-sm">{{ t('zendesk.views') }}</span>
      <Button variant="ghost" size="icon" class="size-7" @click="refreshList">
        <RefreshCw class="size-4" :class="{ 'animate-spin': refreshing }" />
      </Button>
    </div>

    <div class="flex-1 overflow-y-auto py-1">
      <button
        v-for="item in inboxItems"
        :key="item.key"
        type="button"
        class="zendesk-view-item w-full text-left"
        :class="{ active: item.isActive }"
        @click="item.onClick"
      >
        <span class="truncate pr-2">{{ item.label }}</span>
        <span
          v-if="item.showCount && item.isActive"
          class="text-xs tabular-nums text-muted-foreground shrink-0"
        >
          {{ conversationStore.conversations.total }}
        </span>
      </button>

      <div v-if="userTeams.length" class="mt-2 pt-2 border-t">
        <p class="px-3 py-1 text-xs font-medium text-muted-foreground uppercase">
          {{ t('globals.terms.teamInbox', 2) }}
        </p>
        <button
          v-for="team in userTeams"
          :key="team.id"
          type="button"
          class="zendesk-view-item w-full text-left"
          :class="{ active: String(route.params.teamID) === String(team.id) }"
          @click="navigateToTeamInbox(team.id)"
        >
          <span class="truncate">{{ team.emoji }} {{ team.name }}</span>
        </button>
      </div>

      <div v-if="userViews.length && userStore.can(permissions.VIEW_MANAGE)" class="mt-2 pt-2 border-t">
        <div class="flex items-center justify-between px-3 py-1">
          <p class="text-xs font-medium text-muted-foreground uppercase">
            {{ t('globals.terms.view', 2) }}
          </p>
          <Button variant="ghost" size="icon" class="size-6" @click="openCreateView">
            <Plus class="size-3.5" />
          </Button>
        </div>
        <button
          v-for="view in userViews"
          :key="view.id"
          type="button"
          class="zendesk-view-item w-full text-left group"
          :class="{ active: String(route.params.viewID) === String(view.id) }"
          @click="navigateToViewInbox(view.id)"
        >
          <span class="truncate flex-1">{{ view.name }}</span>
        </button>
      </div>

      <div v-if="sharedViews.length" class="mt-2 pt-2 border-t">
        <p class="px-3 py-1 text-xs font-medium text-muted-foreground uppercase">
          {{ t('globals.terms.sharedView', 2) }}
        </p>
        <button
          v-for="view in sharedViews"
          :key="view.id"
          type="button"
          class="zendesk-view-item w-full text-left"
          :class="{ active: String(route.params.viewID) === String(view.id) }"
          @click="navigateToViewInbox(view.id)"
        >
          <span class="truncate">{{ view.name }}</span>
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { RefreshCw, Plus } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { useConversationStore } from '@main/stores/conversation'
import { useUserStore } from '@main/stores/user'
import { permissions } from '@main/constants/permissions'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'

defineProps({
  userTeams: { type: Array, default: () => [] },
  userViews: { type: Array, default: () => [] },
  sharedViews: { type: Array, default: () => [] }
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const conversationStore = useConversationStore()
const userStore = useUserStore()
const refreshing = ref(false)
const emitter = useEmitter()

const openCreateView = () => {
  emitter.emit(EMITTER_EVENTS.OPEN_VIEW_FORM)
}

const inboxItems = computed(() => [
  {
    key: 'assigned',
    label: t('globals.terms.myInbox'),
    isActive: route.path === '/inboxes/assigned' || route.path.startsWith('/inboxes/assigned/'),
    showCount: true,
    onClick: () => navigateToInbox('assigned')
  },
  {
    key: 'mentioned',
    label: t('globals.terms.mention', 2),
    isActive: route.path.startsWith('/inboxes/mentioned'),
    showCount: true,
    onClick: () => navigateToInbox('mentioned')
  },
  {
    key: 'unassigned',
    label: t('globals.terms.unassigned'),
    isActive: route.path.startsWith('/inboxes/unassigned'),
    showCount: true,
    onClick: () => navigateToInbox('unassigned')
  },
  {
    key: 'all',
    label: t('globals.messages.all'),
    isActive: route.path.startsWith('/inboxes/all'),
    showCount: true,
    onClick: () => navigateToInbox('all')
  }
])

const navigateToInbox = (type) => {
  router.push({ name: 'inbox', params: { type } })
}

const navigateToTeamInbox = (teamID) => {
  router.push({ name: 'team-inbox', params: { teamID } })
}

const navigateToViewInbox = (viewID) => {
  router.push({ name: 'view-inbox', params: { viewID } })
}

const refreshList = async () => {
  refreshing.value = true
  try {
    await conversationStore.refreshConversationList()
  } finally {
    refreshing.value = false
  }
}
</script>
