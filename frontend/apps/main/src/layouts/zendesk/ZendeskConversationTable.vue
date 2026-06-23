<template>
  <div class="flex flex-col h-full bg-background overflow-hidden">
    <div class="flex items-center justify-between px-4 h-14 border-b shrink-0">
      <div class="min-w-0">
        <h1 class="zendesk-title truncate">{{ listTitle }}</h1>
        <p v-if="hasConversations" class="zendesk-meta">
          {{ t('zendesk.ticketCount', conversationStore.conversations.total, { count: conversationStore.conversations.total }) }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="zendesk-add-btn shrink-0" @click="openNewTicket">
          <Plus class="size-3.5 mr-1" />
          {{ t('zendesk.add') }}
        </button>
        <Button
          v-if="hasConversations"
          variant="default"
          size="sm"
          @click="startPlay"
        >
          <Play class="size-4 mr-1" />
          {{ t('zendesk.play') }}
        </Button>
        <DropdownMenu v-if="!route.params.viewID">
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm">
              {{ conversationStore.getListStatus || t('globals.messages.filter') }}
              <ChevronDown class="size-4 ml-1 opacity-50" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem
              v-for="status in conversationStore.statusOptions"
              :key="status.value"
              @click="conversationStore.setListStatus(status.label)"
            >
              {{ status.label }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <ConversationBulkActionToolbar v-if="hasSelection && canBulkAct" />
      </div>
    </div>

    <div class="flex-1 overflow-auto">
      <EmptyList
        v-if="showEmpty"
        class="px-4 py-8"
        :title="t('conversation.noConversationsFound')"
        :message="t('conversation.tryAdjustingFilters')"
        :icon="MessageCircleQuestion"
      />

      <table v-else-if="hasConversations" class="zendesk-table w-full">
        <thead class="sticky top-0 z-10">
          <tr>
            <th class="w-10">
              <Checkbox
                v-if="canBulkAct"
                :checked="conversationStore.allSelected"
                class="ml-1"
                @update:checked="onToggleSelectAll"
              />
            </th>
            <th class="w-28">{{ t('globals.terms.status') }}</th>
            <th class="w-20">{{ t('zendesk.id') }}</th>
            <th>{{ t('globals.terms.subject') }}</th>
            <th class="w-28">{{ t('zendesk.channel') }}</th>
            <th class="w-44">{{ t('zendesk.requester') }}</th>
            <th class="w-36">{{ t('zendesk.assignee') }}</th>
            <th class="w-28">
              <button type="button" class="zendesk-sort-th" @click="toggleRequestedSort">
                {{ t('zendesk.requested') }}
                <component
                  :is="requestedSortIcon"
                  v-if="requestedSortIcon"
                  class="size-3"
                />
              </button>
            </th>
            <th class="w-32">{{ t('zendesk.slaColumn') }}</th>
          </tr>
        </thead>
        <tbody>
          <ZendeskConversationTableRow
            v-for="conversation in conversationStore.conversationsList"
            :key="conversation.uuid"
            :conversation="conversation"
            :contact-full-name="conversationStore.getContactFullName(conversation.uuid)"
            :is-current="conversation.uuid === route.params.uuid"
          />
        </tbody>
      </table>

      <div v-if="conversationStore.conversations.loading" class="p-4 space-y-2">
        <ConversationListItemSkeleton v-for="i in 8" :key="i" :index="i - 1" />
      </div>
    </div>

    <div
      v-if="conversationStore.conversations.hasMore"
      class="flex justify-center p-3 border-t shrink-0"
    >
      <Button
        variant="outline"
        size="sm"
        @click="conversationStore.fetchNextConversations"
        :disabled="conversationStore.conversations.fetching"
      >
        <Loader2 v-if="conversationStore.conversations.fetching" class="mr-2 size-4 animate-spin" />
        {{ conversationStore.conversations.fetching ? t('globals.terms.loading') : t('globals.terms.loadMore') }}
      </Button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessageCircleQuestion, ChevronDown, Loader2, Play, ArrowDown, ArrowUp, Plus } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { Checkbox } from '@shared-ui/components/ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@shared-ui/components/ui/dropdown-menu'
import { useConversationStore } from '@/stores/conversation'
import { useBulkActionPermissions } from '@/composables/useBulkActionPermissions'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'
import EmptyList from '@/features/conversation/list/ConversationEmptyList.vue'
import ConversationBulkActionToolbar from '@/features/conversation/list/ConversationBulkActionToolbar.vue'
import ConversationListItemSkeleton from '@/features/conversation/list/ConversationListItemSkeleton.vue'
import ZendeskConversationTableRow from './ZendeskConversationTableRow.vue'
import { useZendeskPlayMode } from '@main/composables/useZendeskPlayMode'

const conversationStore = useConversationStore()
const { canBulkAct } = useBulkActionPermissions()
const route = useRoute()
const { t } = useI18n()
const { startPlay } = useZendeskPlayMode()
const emitter = useEmitter()

const openNewTicket = () => {
  emitter.emit(EMITTER_EVENTS.OPEN_CREATE_CONVERSATION)
}

const hasSelection = computed(() => conversationStore.selectedCount > 0)
const hasConversations = computed(() => conversationStore.conversationsList.length > 0)

const onToggleSelectAll = (checked) => {
  if (checked) conversationStore.selectAll()
  else conversationStore.clearSelection()
}

// "Requested" sorts by created_at; toggle between newest-first and oldest-first.
const toggleRequestedSort = () => {
  const next =
    conversationStore.conversations.sortField === 'started_last' ? 'started_first' : 'started_last'
  conversationStore.setListSortField(next)
}

const requestedSortIcon = computed(() => {
  const field = conversationStore.conversations.sortField
  if (field === 'started_last') return ArrowDown
  if (field === 'started_first') return ArrowUp
  return null
})
const showEmpty = computed(
  () =>
    !hasConversations.value &&
    !conversationStore.conversations.errorMessage &&
    !conversationStore.conversations.loading &&
    conversationStore.conversations.initialized
)

const listTitle = computed(() => {
  const typeKey = route.meta?.typeKey?.(route)
  if (typeKey) return t(typeKey)
  const key = route.meta?.titleKey
  if (!key) return t('globals.terms.inbox')
  return t(key, route.meta?.titleCount || 1)
})
</script>
