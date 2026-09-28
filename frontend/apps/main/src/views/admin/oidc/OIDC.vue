<template>
  <AdminSplitLayout>
    <template #content>
      <router-view />
    </template>

    <template #help>
      <AdminHelpCard>
        <p class="text-sm leading-relaxed text-muted-foreground">{{ $t('admin.oidc.help') }}</p>
        <a
          href="https://docs.libredesk.io/configuration/sso"
          target="_blank"
          rel="noopener noreferrer"
          class="link-style text-sm"
        >
          {{ $t('globals.terms.learnMore') }}
        </a>
      </AdminHelpCard>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import AdminHelpCard from '@main/components/layout/AdminHelpCard.vue'
</script>
