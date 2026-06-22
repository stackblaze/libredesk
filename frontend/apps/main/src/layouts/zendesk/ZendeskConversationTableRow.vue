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
      <span class="text-[#1f73b7] hover:underline">
        {{ conversation.subject || t('zendesk.noSubject') }}
      </span>
    </td>
    <td class="text-muted-foreground">
      {{ contactFullName }}
    </td>
  </tr>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Checkbox } from '@shared-ui/components/ui/checkbox'
import { useConversationStore } from '@main/stores/conversation'
import { useBulkActionPermissions } from '@/composables/useBulkActionPermissions'

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

const statusClass = computed(() => {
  const s = (props.conversation.status || '').toLowerCase()
  if (s === 'new') return 'status-new'
  if (s === 'open') return 'status-open'
  return 'status-default'
})

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
