<template>
  <InboxLayout v-if="layout === UI_LAYOUT_DEFAULT" />
  <ZendeskInboxLayout v-else />
</template>

<script setup>
import { useUiLayout, UI_LAYOUT_DEFAULT } from '@main/composables/useUiLayout'
import InboxLayout from '@main/layouts/inbox/InboxLayout.vue'
import ZendeskInboxLayout from '@main/layouts/zendesk/ZendeskInboxLayout.vue'

const { layout } = useUiLayout()
</script>
