<template>
  <Popover v-model:open="isOpen">
    <PopoverTrigger as-child>
      <button type="button" class="zendesk-nav-item w-full relative" :title="t('globals.terms.notification', 2)">
        <Bell class="size-5" aria-hidden="true" />
        <span
          v-if="notificationStore.unreadCount > 0"
          class="absolute top-2 right-0.5 inline-flex size-3.5 items-center justify-center rounded-full bg-destructive text-[9px] font-medium text-destructive-foreground"
        >
          {{ notificationStore.unreadCount > 99 ? '99' : notificationStore.unreadCount }}
        </span>
        <span class="sr-only">{{ t('globals.terms.notification', 2) }}</span>
      </button>
    </PopoverTrigger>
    <PopoverContent side="right" :side-offset="8" align="end" class="w-96 p-0">
      <NotificationPanel @close="isOpen = false" />
    </PopoverContent>
  </Popover>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { Bell } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { Popover, PopoverContent, PopoverTrigger } from '@shared-ui/components/ui/popover'
import { useNotificationStore } from '@main/stores/notification'
import NotificationPanel from '@main/components/sidebar/NotificationPanel.vue'

const { t } = useI18n()
const notificationStore = useNotificationStore()
const isOpen = ref(false)

onMounted(() => {
  notificationStore.fetchStats()
})

watch(isOpen, (open) => {
  if (open) notificationStore.fetchNotifications()
})
</script>
