<template>
  <AdminSplitLayout>
    <template #content>
      <div class="space-y-6">
        <ProviderForm
          type="completion"
          :title="t('admin.ai.completion')"
          :description="t('admin.ai.completionDescription')"
          :show-completion-fields="true"
        />
        <ProviderForm
          type="embedding"
          :title="t('admin.ai.embedding')"
          :description="t('admin.ai.embeddingDescription')"
          :show-embedding-fields="true"
        />
      </div>
    </template>

    <template #help>
      <AdminHelpCard>
        <p class="text-sm leading-relaxed text-muted-foreground">{{ t('admin.ai.providersHelp') }}</p>
        <a
          href="https://docs.libredesk.io/configuration/ai#providers"
          target="_blank"
          rel="noopener noreferrer"
          class="link-style text-sm"
        >
          {{ t('globals.terms.learnMore') }}
        </a>
      </AdminHelpCard>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import AdminHelpCard from '@main/components/layout/AdminHelpCard.vue'
import ProviderForm from '@/features/admin/ai/ProviderForm.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
