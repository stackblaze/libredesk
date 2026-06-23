<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="h-12 flex-shrink-0 px-2 border-b flex items-center justify-between gap-2">
      <div class="flex items-center gap-1 min-w-0">
        <Button
          v-if="isMobile"
          variant="ghost"
          size="icon"
          class="shrink-0"
          :aria-label="t('globals.terms.back')"
          @click="goBackToList"
        >
          <ArrowLeft class="size-4" />
        </Button>
        <span class="truncate">{{ conversationStore.currentContactName }}</span>
      </div>
      <div class="flex items-center gap-1 shrink-0">
        <Button
          v-if="isMobile"
          variant="ghost"
          size="icon"
          :aria-label="t('globals.terms.details')"
          @click="openConversationSidebar"
        >
          <PanelRight class="size-4" />
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger>
            <div
              v-if="conversationStore.current?.status"
              class="flex items-center space-x-1 cursor-pointer bg-primary px-2 py-1 rounded text-sm"
            >
              <span class="text-secondary font-medium inline-block">
                {{ conversationStore.current?.status }}
              </span>
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem
              v-for="status in conversationStore.statusOptions"
              :key="status.value"
              @click="handleUpdateStatus(status.label)"
            >
              {{ status.label }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" class="w-8 h-8 p-0">
              <MoreHorizontal class="w-4 h-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem @click="downloadTranscript">
              {{ t('conversation.downloadTranscript') }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>

    <!-- Messages & reply box -->
    <div class="flex flex-col flex-grow overflow-hidden">
      <MessageList class="flex-1 overflow-y-auto" />
      <ReplyBox />
    </div>
  </div>
</template>

<script setup>
import { useConversationStore } from '../../stores/conversation'
import { MoreHorizontal, ArrowLeft, PanelRight } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'
import { useIsMobileLayout } from '@main/composables/useIsMobileLayout'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@shared-ui/components/ui/dropdown-menu'
import { Button } from '@shared-ui/components/ui/button'
import MessageList from '@/features/conversation/message/MessageList.vue'
import ReplyBox from './ReplyBox.vue'
import { EMITTER_EVENTS } from '../../constants/emitterEvents.js'
import { CONVERSATION_DEFAULT_STATUSES } from '../../constants/conversation'
import { useEmitter } from '../../composables/useEmitter'
import { useI18n } from 'vue-i18n'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '@main/api'
const conversationStore = useConversationStore()
const emitter = useEmitter()
const route = useRoute()
const router = useRouter()
const isMobile = useIsMobileLayout()
const { t } = useI18n()

const goBackToList = () => {
  if (route.name === 'inbox-conversation') {
    router.push({ name: 'inbox', params: { type: route.params.type || 'assigned' } })
    return
  }
  if (route.name === 'team-inbox-conversation') {
    router.push({ name: 'team-inbox', params: { teamID: route.params.teamID } })
    return
  }
  if (route.name === 'view-inbox-conversation') {
    router.push({ name: 'view-inbox', params: { viewID: route.params.viewID } })
    return
  }
  router.back()
}

const openConversationSidebar = () => {
  emitter.emit(EMITTER_EVENTS.CONVERSATION_SIDEBAR_TOGGLE)
}

const downloadTranscript = async () => {
  const conversation = conversationStore.current
  if (!conversation) return
  try {
    const response = await api.getConversationTranscript(conversation.uuid)
    const url = URL.createObjectURL(response.data)
    const link = document.createElement('a')
    link.href = url
    link.download = `transcript-${conversation.reference_number}.txt`
    document.body.appendChild(link)
    link.click()
    link.remove()
    setTimeout(() => URL.revokeObjectURL(url), 0)
  } catch (error) {
    if (error.response?.data instanceof Blob) {
      try {
        error.response.data = JSON.parse(await error.response.data.text())
      } catch {
        // keep the original blob, handleHTTPError falls back to a generic message
      }
    }
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  }
}

const handleUpdateStatus = (status) => {
  if (status === CONVERSATION_DEFAULT_STATUSES.SNOOZED) {
    emitter.emit(EMITTER_EVENTS.SET_NESTED_COMMAND, {
      command: 'snooze',
      open: true
    })
    return
  }
  conversationStore.updateStatus(status)
}
</script>
