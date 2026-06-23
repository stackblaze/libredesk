import { computed } from 'vue'
import { useUsersStore } from '@main/stores/users'

/** Resolve an agent (assignee) id to a display name / avatar from the users store. */
export function useAgentLookup () {
  const usersStore = useUsersStore()

  const byId = computed(() => {
    const map = {}
    for (const u of usersStore.users) map[String(u.id)] = u
    return map
  })

  const agentName = (id) => {
    if (!id) return ''
    const u = byId.value[String(id)]
    if (!u) return ''
    return `${u.first_name || ''} ${u.last_name || ''}`.trim()
  }

  const agentAvatar = (id) => byId.value[String(id)]?.avatar_url || ''

  return { agentName, agentAvatar, byId }
}
