<template>
  <div class="flex flex-wrap gap-2">
    <Select :model-value="teamValue" @update:model-value="onTeam">
      <SelectTrigger class="h-8 w-[160px]">
        <SelectValue :placeholder="t('report.allTeams')" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="0">{{ t('report.allTeams') }}</SelectItem>
        <SelectItem v-for="team in teams" :key="team.id" :value="String(team.id)">
          {{ team.name }}
        </SelectItem>
      </SelectContent>
    </Select>
    <Select :model-value="priorityValue" @update:model-value="onPriority">
      <SelectTrigger class="h-8 w-[160px]">
        <SelectValue :placeholder="t('report.allPriorities')" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="0">{{ t('report.allPriorities') }}</SelectItem>
        <SelectItem v-for="priority in priorities" :key="priority.id" :value="String(priority.id)">
          {{ priority.name }}
        </SelectItem>
      </SelectContent>
    </Select>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@shared-ui/components/ui/select'

const teamId = defineModel('teamId', { type: Number, default: 0 })
const priorityId = defineModel('priorityId', { type: Number, default: 0 })

const { t } = useI18n()
const teams = ref([])
const priorities = ref([])

const teamValue = computed(() => String(teamId.value || 0))
const priorityValue = computed(() => String(priorityId.value || 0))

function onTeam(value) {
  teamId.value = Number(value) || 0
}
function onPriority(value) {
  priorityId.value = Number(value) || 0
}

onMounted(async () => {
  try {
    const [teamResp, priorityResp] = await Promise.all([api.getTeamsCompact(), api.getPriorities()])
    teams.value = teamResp.data.data || []
    priorities.value = priorityResp.data.data || []
  } catch {
    teams.value = []
    priorities.value = []
  }
})
</script>
