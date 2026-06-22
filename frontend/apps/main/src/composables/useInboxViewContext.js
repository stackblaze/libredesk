import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@main/stores/user'
import { useSharedViewStore } from '@main/stores/sharedView'
import api from '@main/api'

/** Resolves the current inbox/view label and list route for Zendesk-style breadcrumbs. */
export function useInboxViewContext () {
  const route = useRoute()
  const { t } = useI18n()
  const userStore = useUserStore()
  const sharedViewStore = useSharedViewStore()
  const userViews = ref([])

  const listRoute = computed(() => {
    if (route.params.teamID) {
      return { name: 'team-inbox', params: { teamID: route.params.teamID } }
    }
    if (route.params.viewID) {
      return { name: 'view-inbox', params: { viewID: route.params.viewID } }
    }
    return { name: 'inbox', params: { type: route.params.type || 'assigned' } }
  })

  const viewLabel = computed(() => {
    if (route.params.viewID) {
      const id = String(route.params.viewID)
      const view =
        userViews.value.find((v) => String(v.id) === id) ||
        sharedViewStore.sharedViewList.find((v) => String(v.id) === id)
      return view?.name || t('globals.terms.view')
    }
    if (route.params.teamID) {
      const team = userStore.teams.find((tm) => String(tm.id) === String(route.params.teamID))
      if (team) {
        return team.emoji ? `${team.emoji} ${team.name}` : team.name
      }
      return t('globals.terms.teamInbox')
    }
    const typeLabels = {
      assigned: t('globals.terms.myInbox'),
      mentioned: t('globals.terms.mention', 2),
      unassigned: t('globals.terms.unassigned'),
      all: t('globals.messages.all')
    }
    return typeLabels[route.params.type] || typeLabels.assigned
  })

  onMounted(async () => {
    if (!route.params.viewID) return
    sharedViewStore.loadSharedViews()
    try {
      const response = await api.getCurrentUserViews()
      userViews.value = response.data.data || []
    } catch {
      // Breadcrumb falls back to generic "View" label
    }
  })

  return { viewLabel, listRoute }
}
