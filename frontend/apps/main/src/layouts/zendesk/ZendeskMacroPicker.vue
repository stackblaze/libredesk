<template>
  <Select v-model="selectedMacro" @update:model-value="applySelected">
    <SelectTrigger class="h-9">
      <SelectValue :placeholder="t('actions.applyMacro')" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem v-for="macro in macroOptions" :key="macro.id" :value="String(macro.id)">
        {{ macro.name }}
      </SelectItem>
    </SelectContent>
  </Select>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@shared-ui/components/ui/select'
import { useMacroStore } from '@main/stores/macro'
import { useConversationStore } from '@main/stores/conversation'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents'
import { handleHTTPError } from '@shared-ui/utils/http'
import api from '@main/api'

const { t } = useI18n()
const macroStore = useMacroStore()
const conversationStore = useConversationStore()
const emitter = useEmitter()
const selectedMacro = ref('')

const macroOptions = computed(() => macroStore.macroOptions || [])

const applySelected = async (macroId) => {
  if (!macroId || !conversationStore.current?.uuid) return
  const macro = macroOptions.value.find((m) => String(m.id) === macroId)
  if (!macro) return
  try {
    await api.applyMacro(conversationStore.current.uuid, macro.id, macro.actions)
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.savedSuccessfully')
    })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    selectedMacro.value = ''
  }
}
</script>
