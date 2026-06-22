<template>
  <div class="flex flex-col h-full bg-background overflow-hidden">
    <div class="flex items-center justify-between px-4 h-12 border-b shrink-0">
      <h1 class="text-lg font-normal">{{ listTitle }}</h1>
      <div class="flex items-center gap-2">
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
            <th class="w-10" />
            <th class="w-28">{{ t('globals.terms.status') }}</th>
            <th>{{ t('globals.terms.subject') }}</th>
            <th class="w-44">{{ t('zendesk.requester') }}</th>
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
import { MessageCircleQuestion, ChevronDown, Loader2 } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@shared-ui/components/ui/dropdown-menu'
import { useConversationStore } from '@/stores/conversation'
import { useBulkActionPermissions } from '@/composables/useBulkActionPermissions'
import EmptyList from '@/features/conversation/list/ConversationEmptyList.vue'
import ConversationBulkActionToolbar from '@/features/conversation/list/ConversationBulkActionToolbar.vue'
import ConversationListItemSkeleton from '@/features/conversation/list/ConversationListItemSkeleton.vue'
import ZendeskConversationTableRow from './ZendeskConversationTableRow.vue'

const conversationStore = useConversationStore()
const { canBulkAct } = useBulkActionPermissions()
const route = useRoute()
const { t } = useI18n()

const hasSelection = computed(() => conversationStore.selectedCount > 0)
const hasConversations = computed(() => conversationStore.conversationsList.length > 0)
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
