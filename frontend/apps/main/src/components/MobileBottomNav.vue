<template>
  <nav
    class="md:hidden fixed inset-x-0 bottom-0 z-50 flex items-center justify-around border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80 pb-[max(0.5rem,env(safe-area-inset-bottom))] pt-2"
    aria-label="Main navigation"
  >
    <router-link
      :to="lastInboxPath || { name: 'inboxes' }"
      class="mobile-nav-item"
      :class="{ active: route.path.startsWith('/inboxes') }"
      :aria-label="t('globals.terms.inbox', 2)"
    >
      <Inbox class="size-5" aria-hidden="true" />
    </router-link>

    <router-link
      v-if="userStore.can('contacts:read_all')"
      :to="{ name: 'contacts' }"
      class="mobile-nav-item"
      :class="{ active: route.path.startsWith('/contacts') }"
      :aria-label="t('globals.terms.contact', 2)"
    >
      <BookUser class="size-5" aria-hidden="true" />
    </router-link>

    <router-link
      v-if="userStore.hasReportTabPermissions"
      :to="{ name: 'reports' }"
      class="mobile-nav-item"
      :class="{ active: route.path.startsWith('/reports') }"
      :aria-label="t('globals.terms.report', 2)"
    >
      <FileLineChart class="size-5" aria-hidden="true" />
    </router-link>

    <router-link
      v-if="userStore.hasAdminTabPermissions"
      :to="{ name: userStore.can('general_settings:manage') ? 'general' : 'admin' }"
      class="mobile-nav-item"
      :class="{ active: route.path.startsWith('/admin') }"
      :aria-label="t('globals.terms.admin')"
    >
      <Shield class="size-5" aria-hidden="true" />
    </router-link>

    <SidebarNavUser zendesk menu-side="top" />
  </nav>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { useStorage } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { Inbox, Shield, FileLineChart, BookUser } from 'lucide-vue-next'
import { useUserStore } from '@main/stores/user'
import SidebarNavUser from '@main/components/sidebar/SidebarNavUser.vue'

const route = useRoute()
const userStore = useUserStore()
const { t } = useI18n()
const lastInboxPath = useStorage('lastInboxPath', '')
</script>

<style scoped>
.mobile-nav-item {
  @apply flex size-10 items-center justify-center rounded-md text-muted-foreground transition-colors;
}

.mobile-nav-item.active {
  @apply bg-accent text-accent-foreground;
}
</style>
