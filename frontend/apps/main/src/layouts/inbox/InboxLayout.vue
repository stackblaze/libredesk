<template>
  <template v-if="!isSearchRoute">
    <!-- Mobile: one panel at a time (list OR conversation detail) -->
    <div v-if="isMobile" class="h-full w-full min-h-0">
      <ConversationList v-if="!hasOpenConversation" class="h-full" />
      <router-view v-else v-slot="{ Component }">
        <keep-alive>
          <component :is="Component" />
        </keep-alive>
      </router-view>
    </div>

    <!-- Desktop: resizable split view -->
    <ResizablePanelGroup
      v-else
      direction="horizontal"
      class="h-full w-full min-h-0"
      @layout="onLayoutChange"
    >
      <ResizablePanel :default-size="panelSizes[0]" :min-size="20" :max-size="45">
        <ConversationList />
      </ResizablePanel>

      <ResizableHandle />

      <ResizablePanel :default-size="panelSizes[1]" :min-size="30">
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </ResizablePanel>
    </ResizablePanelGroup>
  </template>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useStorage } from '@vueuse/core'
import { useIsMobileLayout } from '@main/composables/useIsMobileLayout'
import ConversationList from '@/features/conversation/list/ConversationList.vue'
import {
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle
} from '@shared-ui/components/ui/resizable'

defineOptions({ name: 'InboxLayout' })

const route = useRoute()
const isMobile = useIsMobileLayout()
const isSearchRoute = computed(() => route.name === 'search')
const hasOpenConversation = computed(() => Boolean(route.params.uuid))

const panelSizes = useStorage('inboxPanelSizes', [25, 75])

const onLayoutChange = (sizes) => {
  panelSizes.value = sizes
}
</script>
