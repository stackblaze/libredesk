import { watch, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useNotificationStore } from '@main/stores/notification'

/** Sync unread notifications to the tab title and installed PWA icon badge. */
export function useNotificationBadge () {
  const route = useRoute()
  const notificationStore = useNotificationStore()

  const sync = (count) => {
    const unread = Math.max(0, Number(count) || 0)

    const base = document.title.replace(/^\(\d+\)\s*/, '')
    document.title = unread > 0 ? `(${unread}) ${base}` : base

    if (!('setAppBadge' in navigator)) return
    const badge = unread > 99 ? 99 : unread
    if (badge > 0) {
      navigator.setAppBadge(badge).catch(() => {})
    } else if ('clearAppBadge' in navigator) {
      navigator.clearAppBadge().catch(() => {})
    }
  }

  watch(
    [() => notificationStore.unreadCount, () => route.fullPath],
    ([count]) => sync(count),
    { immediate: true }
  )

  const onVisibility = () => {
    if (document.visibilityState === 'visible') {
      sync(notificationStore.unreadCount)
    }
  }

  onMounted(() => {
    document.addEventListener('visibilitychange', onVisibility)
  })

  onUnmounted(() => {
    document.removeEventListener('visibilitychange', onVisibility)
    if ('clearAppBadge' in navigator) {
      navigator.clearAppBadge().catch(() => {})
    }
  })
}
