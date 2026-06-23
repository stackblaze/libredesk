<template>
  <div class="flex h-full w-full min-h-0 min-w-0">
    <template v-if="!hasConversation">
      <ZendeskViewsPane
        :user-teams="userTeams"
        :user-views="userViews"
        :shared-views="sharedViews"
      />
      <div class="flex-1 min-w-0 relative">
        <ZendeskConversationTable class="absolute inset-0 z-10" />
        <router-view class="sr-only" aria-hidden="true" />
      </div>
    </template>
    <div v-else class="flex flex-col flex-1 min-w-0 w-full h-full min-h-0">
      <router-view class="flex-1 min-w-0 w-full min-h-0" />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@main/stores/user'
import { useSharedViewStore } from '@main/stores/sharedView'
import { handleHTTPError } from '@shared-ui/utils/http'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'
import api from '@main/api'
import ZendeskViewsPane from './ZendeskViewsPane.vue'
import ZendeskConversationTable from './ZendeskConversationTable.vue'

import { TICKET_ROUTE_NAMES } from '@main/composables/useZendeskTabs'

const route = useRoute()
const userStore = useUserStore()
const sharedViewStore = useSharedViewStore()
const emitter = useEmitter()

const userViews = ref([])
const userTeams = computed(() => userStore.teams || [])
const sharedViews = computed(() => sharedViewStore.sharedViewList || [])

const hasConversation = computed(() => TICKET_ROUTE_NAMES.has(route.name))

onMounted(async () => {
  await loadUserViews()
})

const loadUserViews = async () => {
  try {
    const response = await api.getCurrentUserViews()
    userViews.value = response.data.data
  } catch (err) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(err).message
    })
  }
}
</script>
