<template>
  <header class="zendesk-top-bar flex items-stretch shrink-0 min-w-0">
    <ZendeskTicketTabs v-if="isConversationRoute" class="flex-1 min-w-0" />

    <button
      v-else
      type="button"
      class="zendesk-top-search zendesk-top-search-wide flex items-center gap-2 flex-1 max-w-md h-8 px-3 mx-3 my-auto text-sm text-muted-foreground"
      @click="openSearch"
    >
      <Search class="size-4 shrink-0" />
      <span class="truncate">{{ t('zendesk.searchPlaceholder') }}</span>
      <kbd class="ml-auto hidden sm:inline text-[10px] font-medium px-1.5 py-0.5 rounded border bg-background">
        {{ shortcutHint }}
      </kbd>
    </button>

    <div
      class="zendesk-top-bar-actions flex items-center gap-1 px-2 shrink-0"
      :class="isConversationRoute ? 'border-l' : 'ml-auto'"
    >
      <button
        v-if="isConversationRoute"
        type="button"
        class="zendesk-top-search flex items-center justify-center size-8 text-muted-foreground"
        @click="openSearch"
      >
        <Search class="size-4" />
      </button>

      <Button
        variant="ghost"
        size="icon"
        class="size-8 shrink-0"
        :title="t('command.typeCmdOrSearch')"
        @click="openSearch"
      >
        <LayoutGrid class="size-4" />
      </Button>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Search, LayoutGrid } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { Button } from '@shared-ui/components/ui/button'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'
import ZendeskTicketTabs from './ZendeskTicketTabs.vue'
import { TICKET_ROUTE_NAMES } from '@main/composables/useZendeskTabs'

const route = useRoute()
const { t } = useI18n()
const emitter = useEmitter()

const isConversationRoute = computed(() => TICKET_ROUTE_NAMES.has(route.name))

const shortcutHint = computed(() =>
  navigator.platform?.toLowerCase().includes('mac') ? '\u2318K' : 'Ctrl K'
)

const openSearch = () => {
  emitter.emit(EMITTER_EVENTS.SET_NESTED_COMMAND, { command: null, open: true })
}
</script>
