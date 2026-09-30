<template>
  <div class="space-y-5 max-w-2xl">
    <p v-if="existingCount > 1" class="text-sm text-muted-foreground">
      {{ $t('onboarding.teammates.existing', { count: existingCount }) }}
    </p>
    <form class="space-y-3" @submit.prevent="invite">
      <div v-for="(row, index) in rows" :key="index" class="grid gap-2 sm:grid-cols-[1fr_1.4fr_8rem_auto] sm:items-start">
        <Input v-model.trim="row.first_name" :placeholder="$t('onboarding.teammates.name')" :aria-label="$t('onboarding.teammates.name')" />
        <Input v-model.trim="row.email" type="email" :placeholder="$t('globals.terms.email')" :aria-label="$t('globals.terms.email')" />
        <Select v-model="row.role">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="Agent">Agent</SelectItem>
            <SelectItem value="Admin">Admin</SelectItem>
          </SelectContent>
        </Select>
        <div class="flex h-9 items-center gap-2 text-sm">
          <Check v-if="row.status === 'sent'" class="size-4 text-green-600" />
          <span v-else-if="row.error" class="text-destructive text-xs">{{ row.error }}</span>
          <Button v-else-if="rows.length > 1" type="button" variant="ghost" size="icon" :aria-label="$t('onboarding.teammates.remove')" @click="rows.splice(index, 1)">
            <X class="size-4" />
          </Button>
        </div>
      </div>
      <Button type="button" variant="ghost" size="sm" @click="addRow"><Plus class="size-4" />{{ $t('onboarding.teammates.addAnother') }}</Button>
      <p class="text-xs text-muted-foreground">{{ $t('onboarding.teammates.welcomeNote') }}</p>
      <Button type="submit" :isLoading="isSaving" :disabled="!pendingRows.length">{{ $t('onboarding.teammates.sendInvites') }}</Button>
    </form>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Check, Plus, X } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { Input } from '@shared-ui/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared-ui/components/ui/select'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '@main/api'

const emit = defineEmits(['saved'])
const isSaving = ref(false)
const existingCount = ref(0)
const newRow = () => ({ first_name: '', email: '', role: 'Agent', status: '', error: '' })
const rows = ref([newRow(), newRow()])

const pendingRows = computed(() => rows.value.filter((r) => r.email && r.status !== 'sent'))

onMounted(async () => {
  try {
    const resp = await api.getUsers()
    existingCount.value = (resp.data.data || []).filter((u) => u.type === 'agent').length
  } catch {
    existingCount.value = 0
  }
})

const addRow = () => rows.value.push(newRow())

// Each invite is independent, so one bad address does not block the rest.
const invite = async () => {
  isSaving.value = true
  let sent = 0
  for (const row of pendingRows.value) {
    row.error = ''
    try {
      await api.createUser({
        first_name: row.first_name || row.email.split('@')[0],
        last_name: '',
        email: row.email,
        roles: [row.role],
        teams: [],
        send_welcome_email: true
      })
      row.status = 'sent'
      sent++
    } catch (error) {
      row.error = handleHTTPError(error).message
    }
  }
  isSaving.value = false
  if (sent) emit('saved', { advance: rows.value.every((r) => !r.email || r.status === 'sent') })
}
</script>
