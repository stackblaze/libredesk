<template>
  <tr
    class="cursor-pointer"
    :class="{ selected: isCurrent }"
    @click="openConversation"
  >
    <td @click.stop>
      <Checkbox
        v-if="canBulkAct"
        :checked="isSelected"
        class="ml-1"
        @update:checked="onCheckboxChange"
      />
    </td>
    <td>
      <span class="zendesk-status-badge" :class="statusClass">
        {{ conversation.status }}
      </span>
    </td>
    <td>
      <div class="flex items-center gap-2 min-w-0">
        <span
          v-if="conversation.priority"
          class="zendesk-priority-dot shrink-0"
          :class="priorityClass"
          :title="conversation.priority"
        />
        <component
          :is="channelIcon"
          class="size-3.5 shrink-0 text-muted-foreground"
          aria-hidden="true"
        />
        <span class="text-[#1f73b7] hover:underline truncate">
          {{ conversation.subject || t('zendesk.noSubject') }}
        </span>
      </div>
    </td>
    <td class="text-muted-foreground">
      {{ contactFullName }}
    </td>
    <td>
      <ZendeskSlaIndicator :conversation="conversation" />
    </td>
  </tr>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Mail, MessageSquare } from 'lucide-vue-next'
import { Checkbox } from '@shared-ui/components/ui/checkbox'
import { useConversationStore } from '@main/stores/conversation'
import { useBulkActionPermissions } from '@/composables/useBulkActionPermissions'
import { useStatusCategory } from '@main/composables/useStatusCategory'
import { priorityDotClass } from '@main/composables/useConversationPriority'
import ZendeskSlaIndicator from './ZendeskSlaIndicator.vue'

const props = defineProps({
  conversation: { type: Object, required: true },
  contactFullName: { type: String, default: '' },
  isCurrent: { type: Boolean, default: false }
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const conversationStore = useConversationStore()
const { canBulkAct } = useBulkActionPermissions()
const { categoryClass } = useStatusCategory()

const statusClass = computed(() => categoryClass(props.conversation.status))
const priorityClass = computed(() => priorityDotClass(props.conversation.priority))
const channelIcon = computed(() =>
  props.conversation.inbox_channel === 'livechat' ? MessageSquare : Mail
)

const conversationRoute = computed(() => {
  const baseRoute = route.params.teamID
    ? 'team-inbox-conversation'
    : route.params.viewID
      ? 'view-inbox-conversation'
      : 'inbox-conversation'
  return {
    name: baseRoute,
    params: {
      uuid: props.conversation.uuid,
      ...(baseRoute === 'team-inbox-conversation' && { teamID: route.params.teamID }),
      ...(baseRoute === 'view-inbox-conversation' && { viewID: route.params.viewID }),
      ...(baseRoute === 'inbox-conversation' && { type: route.params.type || 'assigned' })
    },
    query: props.conversation.mentioned_message_uuid
      ? { scrollTo: props.conversation.mentioned_message_uuid }
      : {}
  }
})

const isSelected = computed(() => conversationStore.isSelected(props.conversation.uuid))

const onCheckboxChange = () => {
  conversationStore.toggleSelect(props.conversation.uuid, false)
}

const openConversation = () => {
  router.push(conversationRoute.value)
}
</script>
