<template>
  <AuthLayout>
    <Card class="auth-card w-full rounded-xl border bg-card shadow-lg sm:shadow-sm" id="magic-login-container">
      <CardContent class="p-5 sm:p-6 space-y-5">
        <CardTitle class="text-xl font-bold text-center text-foreground">{{ t('auth.signInButton') }}</CardTitle>

        <p v-if="isLoading && !pendingToken" class="text-sm text-center text-muted-foreground">
          {{ t('auth.signingInWithLink') }}
        </p>

        <form v-if="pendingToken" @submit.prevent="verifyCode" class="space-y-4">
          <div class="space-y-2">
            <Label for="totp">{{ t('auth.totp.code') }}</Label>
            <Input id="totp" v-model.trim="totpCode" inputmode="numeric" autocomplete="one-time-code" autofocus class="h-11" />
          </div>
          <Button class="w-full h-11 text-base" :disabled="isLoading" type="submit">
            {{ t('auth.totp.verify') }}
          </Button>
        </form>

        <Error
          v-if="errorMessage"
          :errorMessage="errorMessage"
          :border="true"
          class="w-full bg-destructive/10 text-destructive border-destructive/20 p-3 rounded-md text-sm"
        />

        <div v-if="errorMessage && !pendingToken" class="text-center">
          <router-link :to="{ name: 'login' }" class="text-sm text-muted-foreground hover:text-foreground">
            {{ t('auth.backToLogin') }}
          </router-link>
        </div>
      </CardContent>
    </Card>
  </AuthLayout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { Button } from '@shared-ui/components/ui/button'
import { Error } from '@shared-ui/components/ui/error'
import { Card, CardContent, CardTitle } from '@shared-ui/components/ui/card'
import { Input } from '@shared-ui/components/ui/input'
import { Label } from '@shared-ui/components/ui/label'
import AuthLayout from '@/layouts/auth/AuthLayout.vue'
import api from '../../api'
import { useUserStore } from '../../stores/user'
import { useAppSettingsStore } from '../../stores/appSettings'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const appSettingsStore = useAppSettingsStore()
const isLoading = ref(true)
const errorMessage = ref('')
const pendingToken = ref('')
const totpCode = ref('')

const finishLogin = (user) => {
  if (user) userStore.setCurrentUser(user)
  appSettingsStore.fetchSettings('general')
  router.replace({ name: 'inboxes' })
}

const handleResponse = (resp) => {
  const data = resp?.data?.data
  if (data?.requires_totp) {
    pendingToken.value = data.pending_token
    return
  }
  finishLogin(data)
}

onMounted(async () => {
  const token = router.currentRoute.value.query.token
  // Drop the token from the address bar and history before using it.
  router.replace({ query: {} })
  if (!token) {
    errorMessage.value = t('auth.magicLinkInvalid')
    isLoading.value = false
    return
  }
  try {
    handleResponse(await api.verifyMagicLink({ token }))
  } catch (error) {
    errorMessage.value = handleHTTPError(error).message
  } finally {
    isLoading.value = false
  }
})

const verifyCode = async () => {
  if (!totpCode.value) {
    errorMessage.value = t('auth.totp.codeRequired')
    return
  }
  errorMessage.value = ''
  isLoading.value = true
  try {
    handleResponse(await api.verifyTOTP({ pending_token: pendingToken.value, code: totpCode.value }))
  } catch (error) {
    errorMessage.value = handleHTTPError(error).message
  } finally {
    isLoading.value = false
  }
}
</script>
