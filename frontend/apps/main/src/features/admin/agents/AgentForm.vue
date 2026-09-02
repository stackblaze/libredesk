<template>
  <form @submit.prevent="onSubmit" novalidate class="space-y-8">
    <!-- Summary Section -->
    <div class="bg-muted/30 box py-6 px-3" v-if="!isNewForm">
      <div class="flex items-start gap-6">
        <Avatar class="w-20 h-20">
          <AvatarImage :src="props.initialValues.avatar_url || ''" :alt="Avatar" />
          <AvatarFallback>
            {{ getInitials(props.initialValues.first_name, props.initialValues.last_name) }}
          </AvatarFallback>
        </Avatar>

        <div class="space-y-4 flex-2">
          <div class="flex items-center gap-3">
            <h3 class="text-lg font-semibold text-foreground">
              {{ props.initialValues.first_name }} {{ props.initialValues.last_name }}
            </h3>
            <Badge :class="['px-2 rounded-full text-xs font-medium', availabilityStatus.color]">
              {{ availabilityStatus.text }}
            </Badge>
          </div>

          <div class="flex flex-wrap items-center gap-6">
            <div class="flex items-center gap-2">
              <Clock class="w-5 h-5 text-muted-foreground" />
              <div>
                <p class="text-sm text-muted-foreground">{{ $t('globals.terms.lastActive') }}</p>
                <p class="text-sm font-medium text-foreground">
                  {{
                    props.initialValues.last_active_at
                      ? format(new Date(props.initialValues.last_active_at), 'PPpp')
                      : 'N/A'
                  }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <LogIn class="w-5 h-5 text-muted-foreground" />
              <div>
                <p class="text-sm text-muted-foreground">{{ $t('globals.terms.lastLogin') }}</p>
                <p class="text-sm font-medium text-foreground">
                  {{
                    props.initialValues.last_login_at
                      ? format(new Date(props.initialValues.last_login_at), 'PPpp')
                      : 'N/A'
                  }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <FormField v-slot="{ value, handleChange }" type="checkbox" name="enabled" v-if="!isNewForm">
      <FormItem class="flex flex-row items-start gap-x-3 space-y-0">
        <FormControl>
          <Checkbox :checked="value" @update:checked="handleChange" />
        </FormControl>
        <div class="space-y-1 leading-none">
          <FormLabel> {{ $t('globals.terms.enabled') }} </FormLabel>
          <FormMessage />
        </div>
      </FormItem>
    </FormField>

    <!-- Form Fields -->
    <div class="grid gap-6 md:grid-cols-2">
      <FormField v-slot="{ field }" name="first_name">
        <FormItem v-auto-animate>
          <FormLabel>{{ $t('globals.terms.firstName') }}</FormLabel>
          <FormControl>
            <Input type="text" placeholder="" v-bind="field" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ field }" name="last_name">
        <FormItem>
          <FormLabel>{{ $t('globals.terms.lastName') }}</FormLabel>
          <FormControl>
            <Input type="text" placeholder="" v-bind="field" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ field }" name="email">
        <FormItem v-auto-animate>
          <FormLabel>{{ $t('globals.terms.email') }}</FormLabel>
          <FormControl>
            <Input type="email" placeholder="" v-bind="field" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField, handleChange }" name="teams">
        <FormItem v-auto-animate>
          <FormLabel>{{ $t('globals.terms.team', 2) }}</FormLabel>
          <FormControl>
            <SelectTag
              :items="teamOptions"
              :placeholder="t('placeholders.selectTeams')"
              v-model="componentField.modelValue"
              @update:modelValue="handleChange"
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField, handleChange }" name="roles">
        <FormItem v-auto-animate>
          <FormLabel>{{ $t('globals.terms.role', 2) }}</FormLabel>
          <FormControl>
            <SelectTag
              :items="roleOptions"
              :placeholder="t('role.select')"
              v-model="componentField.modelValue"
              @update:modelValue="handleChange"
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="availability_status" v-if="!isNewForm">
        <FormItem>
          <FormLabel>{{ t('globals.terms.availabilityStatus') }}</FormLabel>
          <FormControl>
            <Select v-bind="componentField" v-model="componentField.modelValue">
              <SelectTrigger>
                <SelectValue :placeholder="t('agent.selectAvailabilityStatus')" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="active_group">{{ t('globals.terms.active') }}</SelectItem>
                  <SelectItem value="away_manual">{{ t('globals.terms.away') }}</SelectItem>
                  <SelectItem value="away_and_reassigning">
                    {{ t('globals.terms.awayReassigning') }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ field }" name="new_password" v-if="!isNewForm">
        <FormItem v-auto-animate>
          <FormLabel>{{ t('globals.terms.setPassword') }}</FormLabel>
          <FormControl>
            <Input type="password" placeholder="" v-bind="field" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>
    </div>

    <ApiKeyManager
      v-if="!isNewForm && initialValues?.id"
      :api-key="initialValues.api_key"
      :last-used-at="initialValues.api_key_last_used_at"
      :description="t('admin.agent.apiKey.description')"
      :empty-label="t('admin.agent.apiKey.noKey')"
      :generate-fn="generateAgentAPIKey"
      :revoke-fn="revokeAgentAPIKey"
      :fetch-fn="fetchAgentAPIKey"
      :fetch-key="initialValues.id"
    />

    <FormField name="send_welcome_email" v-slot="{ value, handleChange }" v-if="isNewForm">
      <FormItem>
        <FormControl>
          <div class="flex items-center space-x-2">
            <Checkbox :checked="value" @update:checked="handleChange" />
            <Label>{{ $t('globals.terms.sendWelcomeEmail') }}</Label>
          </div>
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>

    <div v-if="!isNewForm && initialValues?.id" class="space-y-3">
      <div>
        <p class="text-base font-semibold text-foreground">{{ t('admin.skill.assign') }}</p>
        <p class="text-sm text-muted-foreground">{{ t('admin.skill.help') }}</p>
      </div>
      <div class="flex flex-wrap gap-3">
        <label v-for="skill in skills" :key="skill.id" class="flex items-center gap-2 text-sm">
          <Checkbox :checked="selectedSkillIds.includes(skill.id)" @update:checked="() => toggleSkill(skill.id)" />
          {{ skill.name }}
        </label>
        <p v-if="!skills.length" class="text-sm text-muted-foreground">{{ t('admin.skill.empty') }}</p>
      </div>
      <div class="flex gap-2 max-w-md">
        <Input v-model.trim="newSkillName" :placeholder="t('admin.skill.name')" />
        <Button type="button" variant="outline" :disabled="!newSkillName || skillBusy" @click="createSkill">
          {{ t('admin.skill.create') }}
        </Button>
      </div>
      <Button type="button" variant="outline" :disabled="skillBusy" @click="saveSkills">
        {{ t('admin.skill.save') }}
      </Button>
    </div>

    <Button type="submit" :isLoading="isLoading"> {{ submitLabel }} </Button>
  </form>
</template>

<script setup>
import { watch, onMounted, ref, computed } from 'vue'
import { Button } from '@shared-ui/components/ui/button/index.js'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { createFormSchema } from './formSchema.js'
import { Checkbox } from '@shared-ui/components/ui/checkbox/index.js'
import { Label } from '@shared-ui/components/ui/label/index.js'
import { vAutoAnimate } from '@formkit/auto-animate/vue'
import { Badge } from '@shared-ui/components/ui/badge/index.js'
import { Clock, LogIn } from 'lucide-vue-next'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@shared-ui/components/ui/form/index.js'
import { Avatar, AvatarFallback, AvatarImage } from '@shared-ui/components/ui/avatar/index.js'
import ApiKeyManager from '@/features/account/ApiKeyManager.vue'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@shared-ui/components/ui/select/index.js'
import { SelectTag } from '@shared-ui/components/ui/select/index.js'
import { Input } from '@shared-ui/components/ui/input/index.js'
import { useI18n } from 'vue-i18n'
import { useEmitter } from '../../../composables/useEmitter.js'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import { format } from 'date-fns'
import api from '../../../api/index.js'

const props = defineProps({
  initialValues: {
    type: Object,
    required: false
  },
  submitForm: {
    type: Function,
    required: true
  },
  submitLabel: {
    type: String,
    required: false,
    default: ''
  },
  isNewForm: {
    type: Boolean,
    required: false,
    default: false
  },
  isLoading: {
    Type: Boolean,
    required: false
  }
})
const { t } = useI18n()
const submitLabel = computed(() => {
  return (
    props.submitLabel ||
    (props.isNewForm ? t('globals.messages.create') : t('globals.messages.save'))
  )
})
const teams = ref([])
const roles = ref([])
const emitter = useEmitter()

const skills = ref([])
const selectedSkillIds = ref([])
const newSkillName = ref('')
const skillBusy = ref(false)

const loadSkills = async () => {
  const [allResp, agentResp] = await Promise.allSettled([
    api.getSkills(),
    props.initialValues?.id ? api.getAgentSkills(props.initialValues.id) : Promise.resolve(null)
  ])
  if (allResp.status === 'fulfilled') skills.value = allResp.value.data.data || []
  if (agentResp.status === 'fulfilled' && agentResp.value) {
    selectedSkillIds.value = (agentResp.value.data.data || []).map((s) => s.id)
  }
}

onMounted(async () => {
  try {
    const [teamsResp, rolesResp] = await Promise.allSettled([api.getTeamsCompact(), api.getRoles()])
    if (teamsResp.status === 'fulfilled') teams.value = teamsResp.value.data.data
    if (rolesResp.status === 'fulfilled') roles.value = rolesResp.value.data.data
    await loadSkills()
  } catch (err) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: t('globals.messages.somethingWentWrong')
    })
  }
})

const availabilityStatus = computed(() => {
  const status = form.values.availability_status
  if (status === 'active_group')
    return { text: t('globals.terms.active'), color: 'bg-success text-success-foreground' }
  if (status === 'away_manual')
    return { text: t('globals.terms.away'), color: 'bg-warning text-warning-foreground' }
  if (status === 'away_and_reassigning')
    return { text: t('globals.terms.awayReassigning'), color: 'bg-warning text-warning-foreground' }
  return { text: t('globals.terms.offline'), color: 'bg-muted text-muted-foreground' }
})

const teamOptions = computed(() =>
  teams.value.map((team) => ({ label: team.name, value: team.name }))
)
const roleOptions = computed(() =>
  roles.value.map((role) => ({ label: role.name, value: role.name }))
)

const form = useForm({
  validationSchema: toTypedSchema(createFormSchema(t))
})

const onSubmit = form.handleSubmit((values) => {
  if (values.availability_status === 'active_group') {
    values.availability_status = 'online'
  }
  props.submitForm(values)
})

const getInitials = (firstName, lastName) => {
  if (!firstName && !lastName) return ''
  if (!firstName) return lastName.charAt(0).toUpperCase()
  if (!lastName) return firstName.charAt(0).toUpperCase()
  return `${firstName.charAt(0).toUpperCase()}${lastName.charAt(0).toUpperCase()}`
}

const generateAgentAPIKey = () => api.generateAPIKey(props.initialValues.id)
const revokeAgentAPIKey = () => api.revokeAPIKey(props.initialValues.id)
const fetchAgentAPIKey = () => api.getAPIKey(props.initialValues.id)

const toggleSkill = (id) => {
  if (selectedSkillIds.value.includes(id)) {
    selectedSkillIds.value = selectedSkillIds.value.filter((x) => x !== id)
  } else {
    selectedSkillIds.value = [...selectedSkillIds.value, id]
  }
}

const createSkill = async () => {
  if (!newSkillName.value) return
  skillBusy.value = true
  try {
    const { data } = await api.createSkill({ name: newSkillName.value })
    if (data.data) {
      skills.value = [...skills.value, data.data].sort((a, b) => a.name.localeCompare(b.name))
      selectedSkillIds.value = [...selectedSkillIds.value, data.data.id]
      newSkillName.value = ''
    }
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: t('globals.messages.somethingWentWrong')
    })
  } finally {
    skillBusy.value = false
  }
}

const saveSkills = async () => {
  if (!props.initialValues?.id) return
  skillBusy.value = true
  try {
    await api.setAgentSkills(props.initialValues.id, { skill_ids: selectedSkillIds.value })
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.savedSuccessfully')
    })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: t('globals.messages.somethingWentWrong')
    })
  } finally {
    skillBusy.value = false
  }
}

watch(
  () => props.initialValues,
  (newValues) => {
    // Hack.
    if (Object.keys(newValues).length > 0) {
      setTimeout(() => {
        if (
          newValues.availability_status === 'away' ||
          newValues.availability_status === 'offline' ||
          newValues.availability_status === 'online'
        ) {
          newValues.availability_status = 'active_group'
        }
        form.setValues(newValues, false)
        form.setFieldValue(
          'teams',
          newValues.teams.map((team) => team.name)
        )
      }, 0)
    }
  },
  { deep: true, immediate: true }
)
</script>
