<template>
  <AdminSplitLayout>
    <template #content>
      <div :class="{ 'opacity-50 transition-opacity duration-300': isLoading }">
        <Spinner v-if="isLoading" />
        <AISettingsForm v-else :initial-values="initialValues" :submit-form="submitForm" />
      </div>
    </template>

    <template #help>
      <p>{{ $t('admin.ai.help.description') }}</p>
      <p>{{ $t('admin.ai.help.detail') }}</p>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '@main/api'
import AdminSplitLayout from '@main/layouts/admin/AdminSplitLayout.vue'
import { useI18n } from 'vue-i18n'
import AISettingsForm from '@/features/admin/ai/AISettingsForm.vue'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'
import { useEmitter } from '@main/composables/useEmitter'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { Spinner } from '@shared-ui/components/ui/spinner'

const initialValues = ref({})
const { t } = useI18n()
const isLoading = ref(false)
const emitter = useEmitter()

onMounted(() => {
  loadSettings()
})

const loadSettings = async () => {
  try {
    isLoading.value = true
    const resp = await api.getAIProvider()
    const data = resp.data.data
    initialValues.value = {
      provider: data.provider || 'openai',
      model: data.model || 'gpt-4o-mini',
      api_key: data.api_key_hint || ''
    }
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

const submitForm = async (values) => {
  const payload = {
    provider: values.provider,
    model: values.model,
    api_key: values.api_key?.includes('•') ? '' : values.api_key
  }
  await api.updateAIProvider(payload)
  emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
    description: t('globals.messages.savedSuccessfully')
  })
  await loadSettings()
}
</script>
