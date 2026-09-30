<template>
  <div class="space-y-5 max-w-xl">
    <div class="flex items-start justify-between gap-4 rounded-lg border p-4">
      <div class="space-y-1">
        <p class="font-medium">{{ $t('admin.sso.magicLink') }}</p>
        <p class="text-sm text-muted-foreground">{{ $t('admin.sso.magicLinkDescription') }}</p>
      </div>
      <Switch :checked="magicLink" :disabled="!loaded || isSaving" @update:checked="toggleMagicLink" />
    </div>
    <div class="rounded-lg border p-4 space-y-3">
      <p class="font-medium">{{ $t('onboarding.signIn.sso') }}</p>
      <p class="text-sm text-muted-foreground">{{ $t('onboarding.signIn.ssoDescription') }}</p>
      <Button variant="outline" size="sm" @click="router.push({ name: 'sso-list' })">
        {{ $t('onboarding.signIn.configureSSO') }}
      </Button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@shared-ui/components/ui/button'
import { Switch } from '@shared-ui/components/ui/switch'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '@main/api'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'

const emit = defineEmits(['saved'])
const router = useRouter()
const emitter = useEmitter()
const loaded = ref(false)
const isSaving = ref(false)
const magicLink = ref(false)
let settings = null

onMounted(async () => {
  try {
    const resp = await api.getSettings('sso')
    settings = resp.data.data
    magicLink.value = settings.magic_link_enabled
    loaded.value = true
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  }
})

// The SSO settings endpoint saves SAML too, so resend the stored SAML config unchanged.
const toggleMagicLink = async (value) => {
  isSaving.value = true
  try {
    const resp = await api.updateSettings('sso', { ...settings, magic_link_enabled: value })
    settings = resp.data.data
    magicLink.value = settings.magic_link_enabled
    emit('saved', { advance: false })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  } finally {
    isSaving.value = false
  }
}
</script>
