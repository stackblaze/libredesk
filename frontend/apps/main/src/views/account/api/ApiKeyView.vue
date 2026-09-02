<template>
  <div class="h-full">
    <div class="space-y-1 mb-6">
      <span class="sub-title">{{ t('account.apiKey.title') }}</span>
      <p class="text-muted-foreground text-xs">{{ t('account.apiKey.pageDescription') }}</p>
    </div>
    <ApiKeyManager
      class="max-w-xl"
      :api-key="userStore.user.api_key"
      :last-used-at="userStore.user.api_key_last_used_at"
      :description="t('account.apiKey.description')"
      :empty-label="t('account.apiKey.noKey')"
      :generate-fn="api.generateMyAPIKey"
      :revoke-fn="api.revokeMyAPIKey"
      :fetch-fn="api.getMyAPIKey"
      @updated="onApiKeyUpdated"
    />
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { useUserStore } from '../../../stores/user'
import api from '../../../api'
import ApiKeyManager from '@/features/account/ApiKeyManager.vue'

const { t } = useI18n()
const userStore = useUserStore()

const onApiKeyUpdated = ({ api_key, api_key_last_used_at }) => {
  userStore.user.api_key = api_key
  userStore.user.api_key_last_used_at = api_key_last_used_at
}
</script>
