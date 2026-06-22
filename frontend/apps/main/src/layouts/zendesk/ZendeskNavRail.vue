<template>
  <nav class="zendesk-nav-rail w-14 shrink-0 flex flex-col h-full py-2">
    <router-link
      :to="lastInboxPath || { name: 'inboxes' }"
      class="zendesk-nav-item"
      :class="{ active: route.path.startsWith('/inboxes') }"
    >
      <Inbox class="size-5" />
      <span>{{ t('globals.terms.inbox', 2) }}</span>
    </router-link>

    <router-link
      v-if="userStore.can('contacts:read_all')"
      :to="{ name: 'contacts' }"
      class="zendesk-nav-item"
      :class="{ active: route.path.startsWith('/contacts') }"
    >
      <BookUser class="size-5" />
      <span>{{ t('globals.terms.contact', 2) }}</span>
    </router-link>

    <router-link
      v-if="userStore.hasReportTabPermissions"
      :to="{ name: 'reports' }"
      class="zendesk-nav-item"
      :class="{ active: route.path.startsWith('/reports') }"
    >
      <FileLineChart class="size-5" />
      <span>{{ t('globals.terms.report', 2) }}</span>
    </router-link>

    <router-link
      v-if="userStore.hasAdminTabPermissions"
      :to="{ name: userStore.can('general_settings:manage') ? 'general' : 'admin' }"
      class="zendesk-nav-item"
      :class="{ active: route.path.startsWith('/admin') }"
    >
      <Shield class="size-5" />
      <span>{{ t('globals.terms.admin') }}</span>
    </router-link>

    <div class="mt-auto flex flex-col gap-1 pb-2">
      <ZendeskNotificationBell />
      <SidebarNavUser zendesk />
    </div>
  </nav>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { useStorage } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { Inbox, Shield, FileLineChart, BookUser } from 'lucide-vue-next'
import { useUserStore } from '@main/stores/user'
import ZendeskNotificationBell from './ZendeskNotificationBell.vue'
import SidebarNavUser from '@main/components/sidebar/SidebarNavUser.vue'

const route = useRoute()
const userStore = useUserStore()
const { t } = useI18n()
const lastInboxPath = useStorage('lastInboxPath', '')
</script>
