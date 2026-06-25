import { computed, unref } from 'vue'
import { useUsersStore } from '@main/stores/users'

/**
 * Agent options for assignee pickers. When a team is assigned, only agents
 * on that team are shown. The current assignee is kept visible if they are
 * not on the team (e.g. legacy assignment before team was set).
 */
export function useTeamFilteredAgentOptions(assignedTeamId, assignedUserId) {
  const usersStore = useUsersStore()

  return computed(() => {
    const teamId = unref(assignedTeamId)
    const userId = unref(assignedUserId)
    const agents = usersStore.options

    if (!teamId || teamId === 'none' || Number(teamId) === 0) {
      return agents
    }

    const teamIdStr = String(teamId)
    const filtered = agents.filter((agent) => agent.team_ids?.includes(teamIdStr))

    if (userId && userId !== 'none' && Number(userId) !== 0) {
      const current = agents.find((agent) => agent.value === String(userId))
      if (current && !filtered.some((agent) => agent.value === current.value)) {
        return [current, ...filtered]
      }
    }

    return filtered
  })
}
