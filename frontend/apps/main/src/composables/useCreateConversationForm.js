import { ref, watch, onUnmounted, nextTick, onMounted, computed } from 'vue'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import { useI18n } from 'vue-i18n'
import { useConversationStore } from '@main/stores/conversation'
import { useInboxStore } from '@main/stores/inbox'
import { useUsersStore } from '@main/stores/users'
import { useTeamStore } from '@main/stores/team'
import { useMacroStore } from '@main/stores/macro'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'
import { MACRO_CONTEXT } from '@main/constants/conversation'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useFileUpload } from '@main/composables/useFileUpload'
import { hasPendingInlineUpload } from '@main/composables/useInlineImageUpload'
import { UserTypeAgent } from '@/constants/user'
import api from '@/api'

export function useCreateConversationForm (options = {}) {
  const { onSuccess, autoFocusEmail = true } = options

  const inboxStore = useInboxStore()
  const { t } = useI18n()
  const uStore = useUsersStore()
  const teamStore = useTeamStore()
  const emitter = useEmitter()
  const conversationStore = useConversationStore()
  const macroStore = useMacroStore()

  const loading = ref(false)
  const searchResults = ref([])
  const emailQuery = ref('')
  const insertContent = ref('')
  const selectedContact = ref(null)
  const emailInputRef = ref(null)
  const highlightedIndex = ref(-1)
  let timeoutId = null
  let previousMacroView = ''

  const {
    uploadingFiles,
    handleFileUpload,
    handleFileDelete,
    uploadFiles,
    mediaFiles,
    clearMediaFiles
  } = useFileUpload({
    linkedModel: 'messages'
  })

  const formSchema = z.object({
    subject: z.string().min(1, t('validation.subjectCannotBeEmpty')),
    content: z.string().min(1, t('validation.messageCannotBeEmpty')),
    inbox_id: z
      .any()
      .refine((val) => inboxStore.emailOptions.some((option) => option.value === val), {
        message: t('globals.messages.required')
      }),
    team_id: z.any().optional(),
    agent_id: z.any().optional(),
    contact_email: z.string().email(t('validation.invalidEmail')),
    first_name: z.string().min(1, t('globals.messages.required')),
    last_name: z.string().optional()
  })

  const form = useForm({
    validationSchema: toTypedSchema(formSchema),
    initialValues: {
      inbox_id: null,
      team_id: null,
      agent_id: null,
      subject: '',
      content: '',
      contact_email: '',
      first_name: '',
      last_name: ''
    }
  })

  const isDisabled = computed(() => {
    if (loading.value || uploadingFiles.value.length > 0) return true
    if (hasPendingInlineUpload(form?.values?.content)) return true
    return false
  })

  const tabPreview = computed(() => {
    const subject = form.values.subject?.trim()
    if (subject) return subject
    const name = `${form.values.first_name || ''} ${form.values.last_name || ''}`.trim()
    if (name) return name
    return ''
  })

  const tabTitle = computed(() => {
    const name = `${form.values.first_name || ''} ${form.values.last_name || ''}`.trim()
    if (name) return name
    if (emailQuery.value.trim()) return emailQuery.value.trim()
    return t('conversation.newConversation')
  })

  const resetForm = () => {
    form.resetForm()
    emailQuery.value = ''
    selectedContact.value = null
    searchResults.value = []
    clearMediaFiles()
  }

  const initMacroContext = () => {
    previousMacroView = macroStore.currentView
    macroStore.setCurrentView('starting_conversation')
    emitter.emit(EMITTER_EVENTS.SET_NESTED_COMMAND, {
      command: 'apply-macro-to-new-conversation',
      open: false
    })
  }

  const teardownMacroContext = () => {
    clearTimeout(timeoutId)
    clearMediaFiles()
    conversationStore.resetMacro(MACRO_CONTEXT.NEW_CONVERSATION)
    macroStore.setCurrentView(previousMacroView)
    emitter.emit(EMITTER_EVENTS.SET_NESTED_COMMAND, {
      command: null,
      open: false
    })
  }

  onMounted(() => {
    initMacroContext()
    if (autoFocusEmail) {
      nextTick(() => {
        document.querySelector('[data-new-ticket-email]')?.focus()
      })
    }
  })

  onUnmounted(() => {
    teardownMacroContext()
  })

  watch(emailQuery, (newVal) => {
    form.setFieldValue('contact_email', newVal)
    if (selectedContact.value && newVal !== selectedContact.value.email) {
      selectedContact.value = null
      form.setFieldValue('first_name', '')
      form.setFieldValue('last_name', '')
    }
  })

  watch(
    () => conversationStore.getMacro(MACRO_CONTEXT.NEW_CONVERSATION).id,
    () => {
      form.setFieldValue(
        'content',
        conversationStore.getMacro(MACRO_CONTEXT.NEW_CONVERSATION).message_content
      )
    },
    { deep: true }
  )

  const handleEmojiSelect = (emoji) => {
    insertContent.value = undefined
    nextTick(() => (insertContent.value = emoji))
  }

  const handleSearchContacts = async () => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(async () => {
      const query = emailQuery.value.trim()

      if (query.length < 3) {
        searchResults.value.splice(0)
        return
      }

      try {
        const resp = await api.searchContacts({ query })
        searchResults.value = [...resp.data.data]
        highlightedIndex.value = -1
      } catch (error) {
        emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
          variant: 'destructive',
          description: handleHTTPError(error).message
        })
        searchResults.value.splice(0)
      }
    }, 300)
  }

  const handleSearchKeydown = (e) => {
    if (!searchResults.value.length) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      highlightedIndex.value = Math.min(highlightedIndex.value + 1, searchResults.value.length - 1)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      highlightedIndex.value = Math.max(highlightedIndex.value - 1, 0)
    } else if (e.key === 'Enter' && highlightedIndex.value >= 0) {
      e.preventDefault()
      selectContact(searchResults.value[highlightedIndex.value])
    } else if (e.key === 'Escape') {
      searchResults.value.splice(0)
      highlightedIndex.value = -1
    }
  }

  const selectContact = (contact) => {
    selectedContact.value = contact
    emailQuery.value = contact.email
    form.setFieldValue('first_name', contact.first_name)
    form.setFieldValue('last_name', contact.last_name || '')
    searchResults.value.splice(0)
    highlightedIndex.value = -1
  }

  const createConversation = form.handleSubmit(async (values) => {
    loading.value = true
    try {
      values.inbox_id = Number(values.inbox_id)
      values.team_id = values.team_id ? Number(values.team_id) : null
      values.agent_id = values.agent_id ? Number(values.agent_id) : null
      values.attachments = mediaFiles.value.map((file) => file.id)
      values.initiator = UserTypeAgent

      const conversation = await api.createConversation(values)
      const conversationUUID = conversation.data.data.uuid

      const macro = conversationStore.getMacro(MACRO_CONTEXT.NEW_CONVERSATION)
      if (conversationUUID !== '' && macro?.id && macro?.actions?.length > 0) {
        try {
          await api.applyMacro(conversationUUID, macro.id, macro.actions)
        } catch (error) {
          emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
            variant: 'destructive',
            description: handleHTTPError(error).message
          })
        }
      }

      if (onSuccess) {
        await onSuccess(conversationUUID)
      }

      resetForm()
      emitter.emit(EMITTER_EVENTS.REFRESH_LIST, { model: 'conversation' })
    } catch (error) {
      emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
        variant: 'destructive',
        description: handleHTTPError(error).message
      })
    } finally {
      loading.value = false
    }
  })

  return {
    inboxStore,
    uStore,
    teamStore,
    conversationStore,
    form,
    loading,
    searchResults,
    emailQuery,
    insertContent,
    selectedContact,
    emailInputRef,
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
    createConversation,
    resetForm,
    MACRO_CONTEXT
  }
}
