<template>
  <div class="flex flex-col h-full">
    <!-- Chat header -->
    <ChatHeader @goBack="goBack" />

    <!-- Pre-chat form -->
    <PreChatForm
      v-if="showPreChatForm"
      @submit="handlePreChatFormSubmit"
      :exclude-default-fields="!!userStore.userSessionToken"
      class="flex-1 min-h-0"
    />

    <!-- Messages container (when no pre-chat form) -->
    <ChatMessages v-else ref="chatMessages" :showPreChatForm="showPreChatForm" />

    <!-- Error display -->
    <WidgetError :errorMessage="errorMessage" />

    <!-- Message input (only when pre-chat form is not shown) -->
    <MessageInput v-if="!showPreChatForm && !isConversationClosed" @error="handleError" />

    <!-- Closed conversation notice -->
    <div v-if="isConversationClosed" class="border-t p-4 text-center text-sm text-muted-foreground">
      {{ $t('widget.conversationClosed') }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useWidgetStore } from '../store/widget.js'
import { useUserStore } from '../store/user.js'
import { useChatStore } from '../store/chat.js'
import WidgetError from '@widget/components/WidgetError.vue'
import ChatHeader from '@widget/components/ChatHeader.vue'
import ChatMessages from '@widget/components/ChatMessages.vue'
import MessageInput from '@widget/components/MessageInput.vue'
import PreChatForm from '@widget/components/PreChatForm.vue'

const widgetStore = useWidgetStore()
const userStore = useUserStore()
const chatStore = useChatStore()
const errorMessage = ref('')
const preChatFormSubmitted = ref(false)
const config = computed(() => widgetStore.config)

// Determine if pre-chat form should be shown
const showPreChatForm = computed(() => {
  const preChatForm = config.value?.prechat_form

  // Must be enabled and not submitted
  if (!preChatForm?.enabled || preChatFormSubmitted.value) {
    return false
  }

  // Atleast one field must be enabled
  const hasEnabledFields = preChatForm.fields?.some((field) => field.enabled)
  if (!hasEnabledFields) {
    return false
  }

  const isAnonymous = !userStore.userSessionToken
  const isNewConversation = !!userStore.userSessionToken && !chatStore.currentConversation?.uuid
  return isAnonymous || isNewConversation
})

// Check if conversation is closed and replies are not allowed
const isConversationClosed = computed(() => {
  const status = chatStore.currentConversation?.status
  if (status !== 'Closed') return false

  const settingsKey = userStore.isVisitor ? 'visitors' : 'users'
  return config.value?.[settingsKey]?.prevent_reply_to_closed_conversation ?? false
})

const goBack = () => {
  widgetStore.navigateToMessages()
}

const handleError = (message) => {
  errorMessage.value = message
  if (message) {
    setTimeout(() => {
      errorMessage.value = ''
    }, 5000)
  }
}

// Handle pre-chat form submission. The conversation is created lazily once the
// visitor sends their first message, so here we only stash the collected form
// data and reveal the message composer.
const handlePreChatFormSubmit = ({ formData }) => {
  const hasFormData = Object.keys(formData).length > 0
  chatStore.pendingFormData = hasFormData ? formData : null
  errorMessage.value = ''
  preChatFormSubmitted.value = true
}
</script>
