<template>
  <div class="zendesk-submit-bar flex items-center justify-end gap-3 px-4 shrink-0">
    <label class="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer select-none">
      <input type="checkbox" v-model="openNext" class="size-3.5 accent-primary" />
      {{ t('zendesk.openNextTicket') }}
    </label>

    <div class="flex items-stretch">
      <Button
        class="rounded-r-none"
        :disabled="!conversationStore.current"
        :is-loading="submitting"
        @click="submit(primaryStatus)"
      >
        {{ t('zendesk.submitAs', { status: primaryStatus }) }}
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button
            class="rounded-l-none border-l border-primary-foreground/30 px-2 [&[data-state=open]>svg]:rotate-180"
            :disabled="!conversationStore.current"
          >
            <ChevronUp class="size-4 text-primary-foreground transition-transform" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" side="top">
          <DropdownMenuLabel>{{ t('zendesk.submitAsLabel') }}</DropdownMenuLabel>
          <DropdownMenuItem
            v-for="status in statusOptions"
            :key="status.value"
            @click="submit(status.label)"
          >
            {{ status.label }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ChevronUp } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { Button } from '@shared-ui/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel
} from '@shared-ui/components/ui/dropdown-menu'
import { useConversationStore } from '@main/stores/conversation'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'
import { useZendeskTicketNav } from '@main/composables/useZendeskTicketNav'
import { useZendeskPlayMode } from '@main/composables/useZendeskPlayMode'

const { t } = useI18n()
const conversationStore = useConversationStore()
const emitter = useEmitter()
const { hasNext, goToNext } = useZendeskTicketNav()
const { openNext, playActive, exitToList } = useZendeskPlayMode()
const submitting = ref(false)

const statusOptions = computed(() => conversationStore.statusOptionsNoSnooze)

const primaryStatus = computed(
  () => conversationStore.current?.status || statusOptions.value[0]?.label || t('zendesk.submit')
)

const submit = (status) => {
  if (!conversationStore.current) return
  submitting.value = true
  emitter.emit(EMITTER_EVENTS.CONVERSATION_SUBMIT_AS, status)
}

const handleSubmitted = () => {
  submitting.value = false
  if (!openNext.value) return
  if (hasNext.value) {
    goToNext()
  } else if (playActive.value) {
    exitToList()
  }
}

onMounted(() => emitter.on(EMITTER_EVENTS.CONVERSATION_SUBMITTED, handleSubmitted))
onUnmounted(() => emitter.off(EMITTER_EVENTS.CONVERSATION_SUBMITTED, handleSubmitted))
</script>
