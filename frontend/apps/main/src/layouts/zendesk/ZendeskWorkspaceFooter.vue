<template>
  <div class="zendesk-workspace-footer">
    <Button variant="outline" size="sm" @click="closeTab">
      {{ t('zendesk.closeTab') }}
    </Button>
    <div class="flex items-center">
      <Button
        size="sm"
        class="rounded-r-none"
        :disabled="!canSubmit"
        @click="submitWithStatus(currentStatus)"
      >
        {{ t('zendesk.submitAs', { status: pendingStatus }) }}
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button size="sm" class="rounded-l-none px-2 border-l border-primary-foreground/30">
            <ChevronDown class="size-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            v-for="status in conversationStore.statusOptionsNoSnooze"
            :key="status.value"
            @click="pendingStatus = status.label"
          >
            {{ status.label }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ChevronDown } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@shared-ui/components/ui/dropdown-menu'
import { useI18n } from 'vue-i18n'
import { useConversationStore } from '@main/stores/conversation'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const conversationStore = useConversationStore()
const emitter = useEmitter()

const pendingStatus = ref(conversationStore.current?.status || 'Open')
const canSubmit = computed(() => !!conversationStore.current)
const currentStatus = computed(() => pendingStatus.value)

const listRoute = computed(() => {
  if (route.params.teamID) {
    return { name: 'team-inbox', params: { teamID: route.params.teamID } }
  }
  if (route.params.viewID) {
    return { name: 'view-inbox', params: { viewID: route.params.viewID } }
  }
  return { name: 'inbox', params: { type: route.params.type || 'assigned' } }
})

const closeTab = () => {
  router.push(listRoute.value)
}

const submitWithStatus = async (status) => {
  if (!conversationStore.current) return
  await conversationStore.updateStatus(status)
  emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
    description: t('globals.messages.savedSuccessfully')
  })
}
</script>
