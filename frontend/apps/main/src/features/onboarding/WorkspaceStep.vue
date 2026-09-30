<template>
  <form class="space-y-5 max-w-xl" @submit.prevent="save">
    <div class="space-y-2">
      <Label for="ob-site-name">{{ $t('onboarding.workspace.siteName') }}</Label>
      <Input id="ob-site-name" v-model.trim="form.site_name" placeholder="Acme Support" required />
      <p class="text-xs text-muted-foreground">{{ $t('onboarding.workspace.siteNameHint') }}</p>
    </div>
    <div class="space-y-2">
      <Label for="ob-root-url">{{ $t('onboarding.workspace.rootURL') }}</Label>
      <Input id="ob-root-url" v-model.trim="form.root_url" type="url" required />
      <p class="text-xs text-muted-foreground">{{ $t('onboarding.workspace.rootURLHint') }}</p>
    </div>
    <div class="space-y-2">
      <Label for="ob-logo">{{ $t('onboarding.workspace.logoURL') }}</Label>
      <Input id="ob-logo" v-model.trim="form.logo_url" type="url" placeholder="https://example.com/logo.png" />
    </div>
    <div class="space-y-2">
      <Label>{{ $t('globals.terms.timezone') }}</Label>
      <Select v-model="form.timezone">
        <SelectTrigger><SelectValue :placeholder="$t('admin.general.timezone.placeholder')" /></SelectTrigger>
        <SelectContent>
          <SelectItem v-for="(value, label) in timeZones" :key="value" :value="value">{{ label }}</SelectItem>
        </SelectContent>
      </Select>
    </div>
    <Button type="submit" :isLoading="isSaving">{{ $t('onboarding.saveAndContinue') }}</Button>
  </form>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Button } from '@shared-ui/components/ui/button'
import { Input } from '@shared-ui/components/ui/input'
import { Label } from '@shared-ui/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared-ui/components/ui/select'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '@main/api'
import { timeZones } from '@main/constants/timezones.js'
import { useAppSettingsStore } from '@main/stores/appSettings'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'

const emit = defineEmits(['saved'])
const emitter = useEmitter()
const appSettingsStore = useAppSettingsStore()
const isSaving = ref(false)
const current = ref({})
const form = ref({ site_name: '', root_url: '', logo_url: '', timezone: '' })

// Keys the server adds to GET responses that are not settings.
const readOnlyKeys = ['app.update', 'app.version', 'app.restart_required']

const isPlaceholderURL = (url) => !url || /localhost|127\.0\.0\.1/.test(url)

onMounted(async () => {
  try {
    const resp = await api.getSettings('general')
    current.value = resp.data.data
    const s = current.value
    form.value = {
      site_name: s['app.site_name'] === 'libredesk' ? '' : s['app.site_name'] || '',
      root_url: isPlaceholderURL(s['app.root_url']) ? window.location.origin : s['app.root_url'],
      logo_url: s['app.logo_url'] || '',
      timezone: s['app.timezone'] || Intl.DateTimeFormat().resolvedOptions().timeZone || ''
    }
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  }
})

// The general settings API replaces every key, so send the full current object with the edits applied.
const save = async () => {
  isSaving.value = true
  try {
    const merged = Object.fromEntries(Object.entries(current.value).filter(([k]) => !readOnlyKeys.includes(k)))
    merged['app.site_name'] = form.value.site_name
    merged['app.root_url'] = form.value.root_url.replace(/\/+$/, '')
    merged['app.logo_url'] = form.value.logo_url
    merged['app.timezone'] = form.value.timezone
    await api.updateSettings('general', merged)
    appSettingsStore.fetchSettings('general')
    emit('saved')
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  } finally {
    isSaving.value = false
  }
}
</script>
