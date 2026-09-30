<template>
  <div class="space-y-5 max-w-xl">
    <div v-if="teams.length" class="flex flex-wrap gap-2">
      <span v-for="team in teams" :key="team.id" class="rounded-full border px-3 py-1 text-sm">
        {{ team.emoji }} {{ team.name }}
      </span>
    </div>
    <form class="space-y-4" @submit.prevent="create">
      <div class="grid gap-3 sm:grid-cols-[1fr_11rem]">
        <div class="space-y-2">
          <Label for="ob-team-name">{{ $t('onboarding.teams.name') }}</Label>
          <Input id="ob-team-name" v-model.trim="name" placeholder="Support" required />
        </div>
        <div class="space-y-2">
          <Label>{{ $t('onboarding.teams.assignment') }}</Label>
          <Select v-model="assignment">
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="Round robin">{{ $t('onboarding.teams.roundRobin') }}</SelectItem>
              <SelectItem value="Manual">{{ $t('onboarding.teams.manual') }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      <p class="text-xs text-muted-foreground">{{ $t('onboarding.teams.hint') }}</p>
      <Button type="submit" :isLoading="isSaving">{{ $t('onboarding.teams.add') }}</Button>
    </form>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Button } from '@shared-ui/components/ui/button'
import { Input } from '@shared-ui/components/ui/input'
import { Label } from '@shared-ui/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared-ui/components/ui/select'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '@main/api'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'

const emit = defineEmits(['saved'])
const emitter = useEmitter()
const isSaving = ref(false)
const teams = ref([])
const name = ref('')
const assignment = ref('Round robin')

const load = async () => {
  try {
    const resp = await api.getTeams()
    teams.value = resp.data.data || []
  } catch {
    teams.value = []
  }
}
onMounted(load)

// Stays on this step so admins can add several teams.
const create = async () => {
  isSaving.value = true
  try {
    await api.createTeam({
      name: name.value,
      emoji: '',
      timezone: '',
      conversation_assignment_type: assignment.value,
      max_auto_assigned_conversations: 0
    })
    name.value = ''
    await load()
    emit('saved', { advance: false })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  } finally {
    isSaving.value = false
  }
}
</script>
