<template>
  <AdminSplitLayout>
    <template #content>
      <DataTable :columns="columns" :data="forms" :loading="loading">
        <template #actions>
          <Button @click="openNew">
            <Plus class="size-4" />
            {{ t('ticketForm.new') }}
          </Button>
        </template>
      </DataTable>
    </template>
    <template #help>
      <div class="rounded-lg border bg-card p-4 shadow-xs">
        <p class="text-sm leading-relaxed text-muted-foreground">{{ t('ticketForm.help') }}</p>
      </div>
    </template>
  </AdminSplitLayout>

  <Dialog :open="open" @update:open="open = $event">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ editingId ? t('globals.messages.edit') : t('ticketForm.new') }}</DialogTitle>
        <DialogDescription>{{ t('ticketForm.description') }}</DialogDescription>
      </DialogHeader>
      <form class="space-y-3" @submit.prevent="save">
        <div class="space-y-1">
          <Label>{{ t('globals.terms.name') }}</Label>
          <Input v-model="draft.name" required maxlength="120" />
        </div>
        <div class="space-y-1">
          <Label>{{ t('globals.terms.description') }}</Label>
          <Textarea v-model="draft.description" rows="2" maxlength="500" />
        </div>
        <div class="space-y-1">
          <Label>{{ t('globals.terms.inbox') }}</Label>
          <Select v-model="draft.inbox_id">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="inbox in inboxes" :key="inbox.id" :value="String(inbox.id)">
                {{ inbox.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <label class="flex items-center gap-2 text-sm">
          <input v-model="draft.enabled" type="checkbox" class="size-4" />
          {{ t('globals.terms.enabled') }}
        </label>
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <Label>{{ t('ticketForm.fields') }}</Label>
            <Button type="button" variant="outline" size="sm" @click="addField">
              {{ t('ticketForm.addField') }}
            </Button>
          </div>
          <p class="text-xs text-muted-foreground">{{ t('ticketForm.keyHint') }}</p>
          <div v-for="(field, index) in draft.fields" :key="index" class="grid grid-cols-[1fr_1fr_auto_auto] gap-2 items-center">
            <Input v-model="field.label" :placeholder="t('ticketForm.fieldLabel')" />
            <Input v-model="field.key" :placeholder="t('ticketForm.fieldKey')" />
            <label class="flex items-center gap-1 text-xs whitespace-nowrap">
              <input v-model="field.required" type="checkbox" />
              {{ t('ticketForm.required') }}
            </label>
            <Button type="button" variant="ghost" size="sm" @click="draft.fields.splice(index, 1)">
              {{ t('globals.messages.delete') }}
            </Button>
          </div>
        </div>
        <p v-if="editingId" class="text-xs text-muted-foreground break-all">
          {{ t('ticketForm.publicLink') }}: {{ publicLink }}
        </p>
        <DialogFooter class="gap-2">
          <Button v-if="editingId" type="button" variant="outline" @click="remove">
            {{ t('globals.messages.delete') }}
          </Button>
          <Button type="submit" :disabled="saving">{{ t('globals.messages.save') }}</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { computed, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus } from 'lucide-vue-next'
import api from '@/api'
import DataTable from '@main/components/datatable/DataTable.vue'
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import { Button } from '@shared-ui/components/ui/button'
import { Input } from '@shared-ui/components/ui/input'
import { Textarea } from '@shared-ui/components/ui/textarea'
import { Label } from '@shared-ui/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@shared-ui/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@shared-ui/components/ui/select'
import { useEmitter } from '@/composables/useEmitter'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'

const { t } = useI18n()
const emitter = useEmitter()
const forms = ref([])
const inboxes = ref([])
const loading = ref(true)
const saving = ref(false)
const open = ref(false)
const editingId = ref(null)
const draft = ref(blank())

const publicLink = computed(() =>
  editingId.value ? `${window.location.origin}/forms/${editingId.value}` : ''
)

const columns = computed(() => [
  {
    accessorKey: 'name',
    header: () => t('globals.terms.name'),
    cell: ({ row }) =>
      h(
        'button',
        {
          class: 'font-medium hover:underline',
          type: 'button',
          onClick: () => edit(row.original)
        },
        row.original.name
      )
  },
  {
    accessorKey: 'inbox_name',
    header: () => t('globals.terms.inbox'),
    cell: ({ row }) => row.original.inbox_name || '—'
  },
  {
    accessorKey: 'enabled',
    header: () => t('globals.terms.enabled'),
    cell: ({ row }) => (row.original.enabled ? t('globals.messages.yes') : t('globals.messages.no'))
  }
])

function blank() {
  return {
    name: '',
    description: '',
    inbox_id: inboxes.value[0] ? String(inboxes.value[0].id) : '',
    enabled: true,
    fields: []
  }
}

function addField() {
  draft.value.fields.push({ key: '', label: '', type: 'text', required: false })
}

function openNew() {
  editingId.value = null
  draft.value = blank()
  open.value = true
}

function edit(form) {
  editingId.value = form.id
  draft.value = {
    name: form.name,
    description: form.description || '',
    inbox_id: String(form.inbox_id),
    enabled: form.enabled,
    fields: Array.isArray(form.fields) ? form.fields.map((field) => ({ ...field })) : []
  }
  open.value = true
}

function payload() {
  return {
    name: draft.value.name.trim(),
    description: draft.value.description.trim(),
    inbox_id: Number(draft.value.inbox_id),
    enabled: draft.value.enabled,
    fields: draft.value.fields
      .map((field) => ({
        key: field.key.trim().toLowerCase().replace(/[^a-z0-9_]/g, ''),
        label: field.label.trim(),
        type: field.type === 'textarea' ? 'textarea' : 'text',
        required: !!field.required
      }))
      .filter((field) => field.key && field.label)
  }
}

async function load() {
  loading.value = true
  try {
    const [formResp, inboxResp] = await Promise.all([api.getTicketForms(), api.getInboxes()])
    forms.value = formResp.data.data || []
    inboxes.value = (inboxResp.data.data || []).filter((inbox) => inbox.enabled)
  } catch (error) {
    toastError(error)
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    if (editingId.value) await api.updateTicketForm(editingId.value, payload())
    else await api.createTicketForm(payload())
    open.value = false
    await load()
  } catch (error) {
    toastError(error)
  } finally {
    saving.value = false
  }
}

async function remove() {
  if (!editingId.value) return
  saving.value = true
  try {
    await api.deleteTicketForm(editingId.value)
    open.value = false
    await load()
  } catch (error) {
    toastError(error)
  } finally {
    saving.value = false
  }
}

function toastError(error) {
  emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
    variant: 'destructive',
    description: handleHTTPError(error).message
  })
}

onMounted(load)
</script>
