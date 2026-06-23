<template>
  <div class="flex flex-col h-full w-full min-h-0 min-w-0 bg-background">
    <ZendeskComposeChrome :compose-id="composeId" />

    <div class="zendesk-ticket-workspace flex-1 min-h-0 min-w-0 context-collapsed">
      <aside class="zendesk-ticket-props p-4">
        <CreateConversationFields
            v-model:email-query="emailQuery"
            layout="sidebar"
            :inbox-store="inboxStore"
            :u-store="uStore"
            :team-store="teamStore"
            :search-results="searchResults"
            :highlighted-index="highlightedIndex"
            :selected-contact="selectedContact"
            :handle-search-contacts="handleSearchContacts"
            :handle-search-keydown="handleSearchKeydown"
            :select-contact="selectContact"
          />
      </aside>

      <div class="zendesk-ticket-center">
        <div class="px-4 py-2.5 border-b shrink-0 bg-background">
          <h2 class="zendesk-title truncate leading-snug">
            {{ form.values.subject?.trim() || t('conversation.newConversation') }}
          </h2>
          <p class="zendesk-meta mt-0.5">
            {{ t('zendesk.newTicketHint') }}
          </p>
        </div>

        <form @submit.prevent class="flex flex-col flex-1 min-h-0 overflow-hidden">
          <div class="flex-1 flex flex-col min-h-0 px-4 py-3 overflow-hidden">
            <FormField v-slot="{ componentField }" name="content">
              <FormItem class="flex flex-col h-full min-h-0">
                <FormControl class="flex-1 flex flex-col min-h-0">
                  <Editor
                    ref="editorRef"
                    v-model:htmlContent="componentField.modelValue"
                    @update:htmlContent="(value) => componentField.onChange(value)"
                    :placeholder="t('editor.hint.newLineCtrlK')"
                    :insertContent="insertContent"
                    :autoFocus="false"
                    :enableInlineImages="true"
                    class="w-full flex-1 overflow-y-auto p-2 box min-h-0 zendesk-new-ticket-editor"
                    @send="submitPrimaryStatus"
                    @filesDropped="uploadFiles"
                  />

                  <MacroActionsPreview
                    v-if="conversationStore.getMacro(MACRO_CONTEXT.NEW_CONVERSATION).actions?.length > 0"
                    :actions="conversationStore.getMacro(MACRO_CONTEXT.NEW_CONVERSATION)?.actions || []"
                    :onRemove="(action) => conversationStore.removeMacroAction(action, MACRO_CONTEXT.NEW_CONVERSATION)"
                    class="mt-2 shrink-0"
                  />

                  <ReplyBoxAttachmentPreview
                    v-if="mediaFiles.length > 0 || uploadingFiles.length > 0"
                    :attachments="mediaFiles"
                    :uploadingFiles="uploadingFiles"
                    :onDelete="handleFileDelete"
                    class="mt-2 shrink-0"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <div class="zendesk-reply-stack shrink-0">
            <div class="px-3 pb-1">
              <ReplyBoxMenuBar
                :handleFileUpload="handleFileUpload"
                @emojiSelect="handleEmojiSelect"
                :showSendButton="false"
                :showFormatting="true"
                :editorApi="editorRef"
              />
            </div>
            <ZendeskSubmitBar
              compose
              :disabled="isDisabled"
              :loading="loading"
              @submit="createConversationWithStatus"
            />
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { FormControl, FormField, FormItem, FormMessage } from '@shared-ui/components/ui/form'
import Editor from '@/components/editor/TextEditor.vue'
import CreateConversationFields from '@/features/conversation/CreateConversationFields.vue'
import ReplyBoxAttachmentPreview from '@/features/conversation/message/attachment/ReplyBoxAttachmentPreview.vue'
import MacroActionsPreview from '@/features/conversation/MacroActionsPreview.vue'
import ReplyBoxMenuBar from '@/features/conversation/ReplyBoxMenuBar.vue'
import ZendeskComposeChrome from '@main/layouts/zendesk/ZendeskComposeChrome.vue'
import ZendeskSubmitBar from '@main/layouts/zendesk/ZendeskSubmitBar.vue'
import { useCreateConversationForm } from '@main/composables/useCreateConversationForm'
import { useZendeskPlayMode } from '@main/composables/useZendeskPlayMode'
import { useZendeskTicketNav } from '@main/composables/useZendeskTicketNav'
import { useConversationStore } from '@main/stores/conversation'
import {
  useZendeskTabs,
  removeComposeTab,
  conversationRouteForContext
} from '@main/composables/useZendeskTabs'

const props = defineProps({
  composeId: { type: String, required: true }
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { updateComposeTab } = useZendeskTabs()
const conversationStore = useConversationStore()
const { openNext, playActive, exitToList } = useZendeskPlayMode()
const { hasNext, goToNext } = useZendeskTicketNav()
const editorRef = ref(null)

const primaryStatus = computed(
  () => conversationStore.statusOptionsNoSnooze[0]?.label || t('zendesk.submit')
)

const {
  inboxStore,
  uStore,
  teamStore,
  form,
  loading,
  searchResults,
  emailQuery,
  insertContent,
  selectedContact,
  highlightedIndex,
  uploadingFiles,
  mediaFiles,
  isDisabled,
  tabPreview,
  tabTitle,
  handleEmojiSelect,
  handleSearchContacts,
  handleSearchKeydown,
  selectContact,
  handleFileUpload,
  handleFileDelete,
  uploadFiles,
  createConversationWithStatus,
  MACRO_CONTEXT
} = useCreateConversationForm({
  onSuccess: async (conversationUUID) => {
    removeComposeTab(props.composeId)

    if (!openNext.value) {
      await router.push(conversationRouteForContext(route, conversationUUID))
      return
    }
    if (hasNext.value) {
      goToNext()
      return
    }
    if (playActive.value) {
      exitToList()
      return
    }
    await router.push(conversationRouteForContext(route, conversationUUID))
  }
})

const submitPrimaryStatus = () => createConversationWithStatus(primaryStatus.value)

watch([tabTitle, tabPreview], () => {
  updateComposeTab(props.composeId, {
    contact_name: tabTitle.value,
    message_preview: tabPreview.value || t('zendesk.newTicketDraft')
  })
}, { immediate: true })
</script>
