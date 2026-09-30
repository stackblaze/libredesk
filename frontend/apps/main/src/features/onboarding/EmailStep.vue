<template>
  <form class="space-y-5 max-w-xl" @submit.prevent="save">
    <div class="grid gap-4 sm:grid-cols-[1fr_8rem]">
      <div class="space-y-2">
        <Label for="ob-smtp-host">{{ $t('onboarding.email.host') }}</Label>
        <Input id="ob-smtp-host" v-model.trim="form.host" placeholder="smtp.example.com" required />
      </div>
      <div class="space-y-2">
        <Label for="ob-smtp-port">{{ $t('onboarding.email.port') }}</Label>
        <Input id="ob-smtp-port" v-model.number="form.port" type="number" min="1" required />
      </div>
    </div>
    <div class="grid gap-4 sm:grid-cols-2">
      <div class="space-y-2">
        <Label for="ob-smtp-user">{{ $t('onboarding.email.username') }}</Label>
        <Input id="ob-smtp-user" v-model.trim="form.username" autocomplete="off" />
      </div>
      <div class="space-y-2">
        <Label for="ob-smtp-pass">{{ $t('onboarding.email.password') }}</Label>
        <Input
          id="ob-smtp-pass"
          v-model="form.password"
          type="password"
          autocomplete="new-password"
          :placeholder="hasSavedPassword ? $t('onboarding.email.passwordSaved') : ''"
        />
      </div>
    </div>
    <div class="grid gap-4 sm:grid-cols-2">
      <div class="space-y-2">
        <Label>{{ $t('onboarding.email.security') }}</Label>
        <Select v-model="form.tls_type">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="starttls">STARTTLS</SelectItem>
            <SelectItem value="tls">SSL/TLS</SelectItem>
            <SelectItem value="none">{{ $t('onboarding.email.none') }}</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="space-y-2">
        <Label>{{ $t('onboarding.email.auth') }}</Label>
        <Select v-model="form.auth_protocol">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="plain">PLAIN</SelectItem>
            <SelectItem value="login">LOGIN</SelectItem>
            <SelectItem value="cram">CRAM-MD5</SelectItem>
            <SelectItem value="none">{{ $t('onboarding.email.none') }}</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
    <div class="space-y-2">
      <Label for="ob-smtp-from">{{ $t('onboarding.email.fromAddress') }}</Label>
      <Input id="ob-smtp-from" v-model.trim="form.email_address" type="email" placeholder="support@example.com" required />
    </div>
    <p class="text-xs text-muted-foreground">{{ $t('onboarding.email.restartNote') }}</p>
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
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'

const emit = defineEmits(['saved'])
const emitter = useEmitter()
const prefix = 'notification.email.'
const isSaving = ref(false)
const hasSavedPassword = ref(false)
const current = ref({})
const form = ref({
  host: '',
  port: 587,
  username: '',
  password: '',
  tls_type: 'starttls',
  auth_protocol: 'plain',
  email_address: ''
})

onMounted(async () => {
  try {
    const resp = await api.getEmailNotificationSettings()
    current.value = resp.data.data
    const get = (k) => current.value[prefix + k]
    hasSavedPassword.value = !!get('password')
    form.value = {
      host: get('host') || '',
      port: get('port') || 587,
      username: get('username') === 'admin@yourcompany.com' ? '' : get('username') || '',
      password: '',
      tls_type: get('tls_type') || 'starttls',
      auth_protocol: get('auth_protocol') || 'plain',
      email_address: get('email_address') === 'admin@yourcompany.com' ? '' : get('email_address') || ''
    }
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  }
})

// The API replaces every key and keeps the stored password only when it receives an empty one.
const save = async () => {
  isSaving.value = true
  try {
    const merged = { ...current.value }
    for (const [k, v] of Object.entries(form.value)) merged[prefix + k] = v
    merged[prefix + 'password'] = form.value.password
    merged[prefix + 'enabled'] = true
    await api.updateEmailNotificationSettings(merged)
    emit('saved')
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  } finally {
    isSaving.value = false
  }
}
</script>
