<template>
  <header class="zendesk-top-bar flex items-center gap-3 px-3 shrink-0">
    <button
      type="button"
      class="zendesk-top-search flex items-center gap-2 flex-1 max-w-md h-8 px-3 text-sm text-muted-foreground"
      @click="openSearch"
    >
      <Search class="size-4 shrink-0" />
      <span class="truncate">{{ t('zendesk.searchPlaceholder') }}</span>
      <kbd class="ml-auto hidden sm:inline text-[10px] font-medium px-1.5 py-0.5 rounded border bg-background">
        {{ shortcutHint }}
      </kbd>
    </button>

    <div class="ml-auto flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon"
        class="size-8"
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
import { Search, LayoutGrid } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'

const emitter = useEmitter()

const shortcutHint = computed(() =>
  navigator.platform?.toLowerCase().includes('mac') ? '\u2318K' : 'Ctrl K'
)

const openSearch = () => {
  emitter.emit(EMITTER_EVENTS.SET_NESTED_COMMAND, { command: null, open: true })
}
</script>
