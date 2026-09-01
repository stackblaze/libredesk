<template>
  <ContextMenuContent class="w-56">
    <ContextMenuItem v-if="canBulkAct" @click="actions.toggleSelect(conversation.uuid)">
      <SquareCheck class="w-4 h-4 mr-2" />
      {{
        isSelected
          ? t('conversation.bulkActions.clearSelection')
          : t('conversation.bulkActions.selectConversation')
      }}
      <ContextMenuShortcut>X</ContextMenuShortcut>
    </ContextMenuItem>
    <ContextMenuItem @click="actions.markUnread(conversation.uuid)">
      <MailOpen class="w-4 h-4 mr-2" />
      {{ t('globals.messages.markAsUnread') }}
    </ContextMenuItem>

    <ContextMenuSeparator v-if="canAssignAgent || canUpdateStatus" />

    <ContextMenuItem v-if="canAssignAgent" @click="actions.assignToMe(conversation.uuid)">
      <UserPlus class="w-4 h-4 mr-2" />
      {{ t('actions.assignToMe') }}
      <ContextMenuShortcut>=</ContextMenuShortcut>
    </ContextMenuItem>

    <ContextMenuSub v-if="canUpdateStatus">
      <ContextMenuSubTrigger>
        <CircleDot class="w-4 h-4 mr-2" />
        {{ t('actions.setStatus') }}
      </ContextMenuSubTrigger>
      <ContextMenuSubContent>
        <ContextMenuItem
          v-for="status in conversationStore.statusOptionsNoSnooze"
          :key="status.value"
          @click="actions.setStatus(conversation.uuid, status.label)"
        >
          {{ status.label }}
        </ContextMenuItem>
      </ContextMenuSubContent>
    </ContextMenuSub>

    <ContextMenuSeparator v-if="canWrite" />

    <ContextMenuItem v-if="canWrite" @click="actions.createLinked(conversation.uuid, 'child')">
      <GitBranch class="w-4 h-4 mr-2" />
      {{ t('conversation.related.createChild') }}
    </ContextMenuItem>
    <ContextMenuItem v-if="canWrite" @click="actions.createLinked(conversation.uuid, 'follow_up')">
      <CornerDownRight class="w-4 h-4 mr-2" />
      {{ t('conversation.related.createFollowUp') }}
    </ContextMenuItem>
    <ContextMenuItem v-if="canWrite" @click="actions.openMerge(conversation.uuid)">
      <GitMerge class="w-4 h-4 mr-2" />
      {{ t('conversation.merge.action') }}
    </ContextMenuItem>
    <ContextMenuItem v-if="canWrite" @click="actions.startSplit(conversation.uuid)">
      <Scissors class="w-4 h-4 mr-2" />
      {{ t('conversation.split.start') }}
    </ContextMenuItem>

    <ContextMenuSeparator v-if="canReply || canNote" />

    <ContextMenuItem v-if="canReply" @click="actions.focusComposer(conversation.uuid, 'reply')">
      <Reply class="w-4 h-4 mr-2" />
      {{ t('shortcuts.reply') }}
      <ContextMenuShortcut>R</ContextMenuShortcut>
    </ContextMenuItem>
    <ContextMenuItem v-if="canNote" @click="actions.focusComposer(conversation.uuid, 'private_note')">
      <StickyNote class="w-4 h-4 mr-2" />
      {{ t('actions.addPrivateNote') }}
      <ContextMenuShortcut>N</ContextMenuShortcut>
    </ContextMenuItem>

    <ContextMenuSeparator v-if="canWrite" />
    <ContextMenuItem v-if="canWrite" class="text-destructive focus:text-destructive" @click="askDelete">
      <Trash2 class="w-4 h-4 mr-2" />
      {{ deleteLabel }}
    </ContextMenuItem>
  </ContextMenuContent>

  <AlertDialog :open="deleteOpen" @update:open="deleteOpen = $event">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ deleteLabel }}</AlertDialogTitle>
        <AlertDialogDescription>{{ deleteDescription }}</AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t('globals.messages.cancel') }}</AlertDialogCancel>
        <AlertDialogAction variant="destructive" @click="actions.deleteTickets(conversation.uuid)">
          {{ deleteLabel }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup>
import { computed, nextTick, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  CircleDot,
  CornerDownRight,
  GitBranch,
  GitMerge,
  MailOpen,
  Scissors,
  Reply,
  SquareCheck,
  StickyNote,
  Trash2,
  UserPlus
} from 'lucide-vue-next'
import {
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuShortcut,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger
} from '@shared-ui/components/ui/context-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle
} from '@shared-ui/components/ui/alert-dialog'
import { useConversationStore } from '@main/stores/conversation'
import { useUserStore } from '@main/stores/user'
import { useBulkActionPermissions } from '@/composables/useBulkActionPermissions'
import { useTicketActions } from '@main/composables/useTicketActions'
import { permissions as perms } from '@main/constants/permissions'

const props = defineProps({
  conversation: { type: Object, required: true }
})

const { t } = useI18n()
const conversationStore = useConversationStore()
const userStore = useUserStore()
const actions = useTicketActions()
const { canBulkAct, canAssignAgent, canUpdateStatus } = useBulkActionPermissions()

const canWrite = computed(() => userStore.can(perms.CONVERSATIONS_WRITE))
const canReply = computed(() => userStore.can(perms.MESSAGES_WRITE))
const canNote = computed(() => userStore.can(perms.MESSAGES_WRITE_PRIVATE))
const isSelected = computed(() => conversationStore.isSelected(props.conversation.uuid))
const deleteOpen = ref(false)
const deleteCount = computed(() => actions.statusTargets(props.conversation.uuid).length)
const deleteLabel = computed(() =>
  deleteCount.value > 1
    ? t('conversation.spam.deleteTickets', deleteCount.value, { count: deleteCount.value })
    : t('conversation.spam.deleteTicket')
)
const deleteDescription = computed(() =>
  deleteCount.value > 1
    ? t('conversation.spam.deleteTicketsConfirm', deleteCount.value, { count: deleteCount.value })
    : t('conversation.spam.deleteTicketConfirm')
)

const askDelete = () => {
  nextTick(() => {
    deleteOpen.value = true
  })
}
</script>
