<template>
  <div class="space-y-6">
    <section class="rounded-lg border p-4 space-y-3">
      <div class="flex items-start justify-between gap-4">
        <div class="space-y-1">
          <h3 class="font-medium">{{ $t('admin.sso.magicLink') }}</h3>
          <p class="text-sm text-muted-foreground">{{ $t('admin.sso.magicLinkDescription') }}</p>
        </div>
        <Switch :checked="form.magic_link_enabled" @update:checked="(v) => (form.magic_link_enabled = v)" />
      </div>
    </section>

    <section class="rounded-lg border p-4 space-y-4">
      <div class="flex items-start justify-between gap-4">
        <div class="space-y-1">
          <h3 class="font-medium">{{ $t('admin.sso.saml') }}</h3>
          <p class="text-sm text-muted-foreground">{{ $t('admin.sso.samlDescription') }}</p>
        </div>
        <Switch :checked="form.saml.enabled" @update:checked="(v) => (form.saml.enabled = v)" />
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <Label>{{ $t('admin.sso.samlEntityID') }}</Label>
          <Input :model-value="form.saml.entity_id" readonly />
        </div>
        <div class="space-y-2">
          <Label>{{ $t('admin.sso.samlACSURL') }}</Label>
          <Input :model-value="form.saml.acs_url" readonly />
        </div>
      </div>

      <template v-if="form.saml.enabled">
        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-2">
            <Label for="saml-name">{{ $t('admin.sso.samlButtonName') }}</Label>
            <Input id="saml-name" v-model.trim="form.saml.name" placeholder="Okta" />
          </div>
          <div class="space-y-2">
            <Label for="saml-email-attr">{{ $t('admin.sso.samlEmailAttribute') }}</Label>
            <Input id="saml-email-attr" v-model.trim="form.saml.email_attribute" placeholder="email" />
            <p class="text-xs text-muted-foreground">{{ $t('admin.sso.samlEmailAttributeHint') }}</p>
          </div>
        </div>
        <div class="space-y-2">
          <Label for="saml-meta-url">{{ $t('admin.sso.samlMetadataURL') }}</Label>
          <Input
            id="saml-meta-url"
            v-model.trim="form.saml.idp_metadata_url"
            placeholder="https://your-org.okta.com/app/abc123/sso/saml/metadata"
          />
        </div>
        <div class="space-y-2">
          <Label for="saml-meta-xml">{{ $t('admin.sso.samlMetadataXML') }}</Label>
          <Textarea id="saml-meta-xml" v-model="form.saml.idp_metadata_xml" rows="5" class="font-mono text-xs" />
          <p class="text-xs text-muted-foreground">{{ $t('admin.sso.samlMetadataHint') }}</p>
        </div>
        <div class="flex items-center gap-2">
          <Checkbox
            id="saml-idp-init"
            :checked="form.saml.allow_idp_initiated"
            @update:checked="(v) => (form.saml.allow_idp_initiated = v)"
          />
          <Label for="saml-idp-init">{{ $t('admin.sso.samlAllowIdpInitiated') }}</Label>
        </div>
        <div v-if="form.saml.sp_certificate" class="space-y-2">
          <Label>{{ $t('admin.sso.samlSPMetadataURL') }}</Label>
          <Input :model-value="form.saml.metadata_url" readonly />
        </div>
      </template>
    </section>

    <Button :isLoading="isSaving" @click="save">{{ $t('globals.messages.save') }}</Button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@shared-ui/components/ui/button'
import { Input } from '@shared-ui/components/ui/input'
import { Label } from '@shared-ui/components/ui/label'
import { Switch } from '@shared-ui/components/ui/switch'
import { Checkbox } from '@shared-ui/components/ui/checkbox'
import { Textarea } from '@shared-ui/components/ui/textarea'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useEmitter } from '../../../composables/useEmitter'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import api from '../../../api'

const { t } = useI18n()
const emitter = useEmitter()
const isSaving = ref(false)
const form = ref({
  magic_link_enabled: false,
  saml: {
    enabled: false,
    name: '',
    idp_metadata_url: '',
    idp_metadata_xml: '',
    allow_idp_initiated: false,
    email_attribute: '',
    entity_id: '',
    acs_url: '',
    metadata_url: '',
    sp_certificate: ''
  }
})

onMounted(async () => {
  try {
    const resp = await api.getSettings('sso')
    form.value = resp.data.data
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  }
})

const save = async () => {
  isSaving.value = true
  try {
    const resp = await api.updateSettings('sso', form.value)
    form.value = resp.data.data
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { description: t('admin.sso.saved') })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  } finally {
    isSaving.value = false
  }
}
</script>
