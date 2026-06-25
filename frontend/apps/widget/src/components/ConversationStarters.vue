<template>
  <div v-if="starters.length" class="flex flex-col gap-2 px-4 pt-3 pb-4 bg-background">
    <div v-if="heading" class="text-sm font-medium text-foreground">{{ heading }}</div>
    <div class="space-y-1.5">
      <Card
        v-for="(starter, index) in starters"
        :key="index"
        class="hover:bg-accent transition-colors cursor-pointer rounded-md"
        role="button"
        @click="start(starter)"
      >
        <CardContent class="px-4 py-2.5">
          <div class="flex justify-between items-center gap-3">
            <span class="text-sm text-foreground font-medium">{{ starter.text }}</span>
            <ArrowRight size="16" class="text-muted-foreground shrink-0" />
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ArrowRight } from 'lucide-vue-next'
import { Card, CardContent } from '@shared-ui/components/ui/card'
import { useWidgetStore } from '@widget/store/widget.js'
import { useChatStore } from '@widget/store/chat.js'

const widgetStore = useWidgetStore()
const chatStore = useChatStore()
const config = computed(() => widgetStore.config)

// Sensible defaults; all route to the live-chat agent. Override per-deploy by
// adding home_apps items of type 'conversation_starter' ({ text, message }).
const DEFAULT_STARTERS = [
  { text: 'Talk to sales', message: "I'd like to talk to sales." },
  { text: 'Get help with my account', message: 'I need help with my account.' },
  { text: 'Pricing & plans', message: 'Can you tell me about pricing and plans?' },
  { text: 'Book a demo', message: "I'd like to book a demo." }
]

const heading = computed(() => config.value?.conversation_starters_heading || 'How can we help you today?')

const starters = computed(() => {
  const configured = (config.value?.home_apps || [])
    .filter((item) => item.type === 'conversation_starter' && item.text)
    .map((item) => ({ text: item.text, message: item.message || item.text }))
  return configured.length ? configured : DEFAULT_STARTERS
})

const start = (starter) => {
  chatStore.pendingStarterMessage = starter.message || starter.text
  chatStore.setCurrentConversation(null)
  widgetStore.navigateToChat()
}
</script>
