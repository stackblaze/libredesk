import { computed } from 'vue'
import { useConversationStore } from '@main/stores/conversation'

const CATEGORY_CLASS = {
  open: 'category-open',
  waiting: 'category-waiting',
  resolved: 'category-resolved'
}

/**
 * Maps a conversation status name to its category-based badge class.
 * Categories are fixed (open | waiting | resolved); custom statuses inherit
 * the color of whichever category an admin assigned them to.
 */
export function useStatusCategory () {
  const conversationStore = useConversationStore()

  const categoryByName = computed(() => {
    const map = {}
    for (const s of conversationStore.statuses) {
      if (s?.name) map[s.name.toLowerCase()] = (s.category || '').toLowerCase()
    }
    return map
  })

  const categoryClass = (statusName) => {
    const category = categoryByName.value[(statusName || '').toLowerCase()]
    return CATEGORY_CLASS[category] || 'category-default'
  }

  return { categoryClass }
}
