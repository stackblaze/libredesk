<template>
  <AdminSplitLayout>
    <template #content>
      <router-view />
    </template>

    <template #help>
      <AdminHelpCard>
        <p class="text-sm leading-relaxed text-muted-foreground">
          {{ $t('admin.automation.help') }}
        </p>
      </AdminHelpCard>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import AdminHelpCard from '@main/components/layout/AdminHelpCard.vue'
</script>
