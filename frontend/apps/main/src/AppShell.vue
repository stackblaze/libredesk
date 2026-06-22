<template>
  <App v-if="layout === UI_LAYOUT_DEFAULT" />
  <ZendeskApp v-else />
</template>

<script setup>
import { useUiLayout, UI_LAYOUT_DEFAULT } from '@main/composables/useUiLayout'
import { useZendeskViewport } from '@main/composables/useZendeskViewport'
import App from '@main/App.vue'
import ZendeskApp from '@main/layouts/zendesk/ZendeskApp.vue'

const { layout } = useUiLayout()
useZendeskViewport()
</script>
