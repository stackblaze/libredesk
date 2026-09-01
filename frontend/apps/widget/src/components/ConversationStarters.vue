<template>
  <div v-if="starters.length" class="widget-starters" :class="{ 'pt-2': heading }">
    <div v-if="heading" class="widget-starters__heading">{{ heading }}</div>
    <div class="widget-starters__list">
      <button
        v-for="(starter, index) in starters"
        :key="index"
        type="button"
        class="widget-starter"
        @click="start(starter)"
      >
        <span>{{ starter.text }}</span>
        <ArrowRight class="widget-starter__icon" />
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ArrowRight } from 'lucide-vue-next'
import { useWidgetStore } from '@widget/store/widget.js'
import { useChatStore } from '@widget/store/chat.js'

const widgetStore = useWidgetStore()
const chatStore = useChatStore()
const config = computed(() => widgetStore.config)

// Sales vs support: the black Start conversation button and sales starters
// land on the Sales team; account help is Support. Override per-deploy by
// adding home_apps items of type 'conversation_starter' ({ text, message, intent }).
const DEFAULT_STARTERS = [
  { text: 'Talk to sales', message: "I'd like to talk to sales.", intent: 'sales' },
  { text: 'Get help with my account', message: 'I need help with my account.', intent: 'support' },
  { text: 'Pricing & plans', message: 'Can you tell me about pricing and plans?', intent: 'sales' },
  { text: 'Book a demo', message: "I'd like to book a demo.", intent: 'sales' }
]

// Optional section title; leave empty by default — the home header greeting
// already welcomes the visitor, so a second heading feels redundant.
const heading = computed(() => config.value?.conversation_starters_heading || '')

const starters = computed(() => {
  const configured = (config.value?.home_apps || [])
    .filter((item) => item.type === 'conversation_starter' && item.text)
    .map((item) => ({
      text: item.text,
      message: item.message || item.text,
      intent: item.intent === 'support' ? 'support' : 'sales'
    }))
  return configured.length ? configured : DEFAULT_STARTERS
})

const start = (starter) => {
  chatStore.pendingStarterMessage = starter.message || starter.text
  chatStore.pendingIntent = starter.intent === 'support' ? 'support' : 'sales'
  chatStore.setCurrentConversation(null)
  widgetStore.navigateToChat()
}
</script>
