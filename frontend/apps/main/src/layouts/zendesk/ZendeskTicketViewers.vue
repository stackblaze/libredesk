<template>
  <div
    v-if="viewers.length"
    class="zendesk-ticket-viewers flex items-center -space-x-2"
    :aria-label="t('zendesk.viewersAria', viewers.length, { count: viewers.length })"
  >
    <Tooltip v-for="viewer in viewers" :key="viewer.id">
      <TooltipTrigger as-child>
        <Avatar class="size-7 ring-2 ring-background">
          <AvatarImage :src="viewer.avatar_url" :alt="viewer.name" />
          <AvatarFallback class="text-[10px] font-medium">
            {{ viewer.initials }}
          </AvatarFallback>
        </Avatar>
      </TooltipTrigger>
      <TooltipContent>
        {{ t('zendesk.viewingTicket', { name: viewer.name }) }}
      </TooltipContent>
    </Tooltip>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Avatar, AvatarFallback, AvatarImage } from '@shared-ui/components/ui/avatar'
import { Tooltip, TooltipContent, TooltipTrigger } from '@shared-ui/components/ui/tooltip'
import { useConversationStore } from '@main/stores/conversation'
import { useUsersStore } from '@main/stores/users'

const { t } = useI18n()
const conversationStore = useConversationStore()
const usersStore = useUsersStore()

const viewers = computed(() =>
  conversationStore.otherViewersOnCurrent.map((id) => {
    const user = usersStore.users.find((u) => Number(u.id) === Number(id))
    const first = user?.first_name || ''
    const last = user?.last_name || ''
    return {
      id,
      name: `${first} ${last}`.trim() || t('zendesk.viewerFallback', { id }),
      avatar_url: user?.avatar_url || '',
      initials: `${first.charAt(0) || 'A'}${last.charAt(0) || ''}`.toUpperCase()
    }
  })
)
</script>
