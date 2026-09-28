<template>
  <AdminSplitLayout>
    <template #content>
      <DataTable :columns="columns" :data="orgs" :loading="isLoading">
        <template #actions>
          <Dialog v-model:open="dialogOpen">
            <DialogTrigger as-child @click="newOrg">
              <Button>
                <Plus class="size-4" />
                {{ t('organization.new') }}
              </Button>
            </DialogTrigger>
            <DialogContent class="sm:max-w-[480px]">
              <DialogHeader>
                <DialogTitle class="mb-1">
                  {{ isEditing ? t('organization.edit') : t('organization.new') }}
                </DialogTitle>
                <DialogDescription>
                  {{ t('organization.formDescription') }}
                </DialogDescription>
              </DialogHeader>
              <form class="space-y-4" @submit.prevent="onSubmit">
                <div class="space-y-1">
                  <label class="text-sm font-medium">{{ t('globals.terms.name') }}</label>
                  <Input v-model="form.name" type="text" required />
                </div>
                <div class="space-y-1">
                  <label class="text-sm font-medium">{{ t('organization.domains') }}</label>
                  <Input
                    v-model="form.domainsText"
                    type="text"
                    :placeholder="t('organization.domainsPlaceholder')"
                  />
                  <p class="text-xs text-muted-foreground">{{ t('organization.domainsHint') }}</p>
                </div>
                <div class="space-y-1">
                  <label class="text-sm font-medium">{{ t('globals.terms.note', 2) }}</label>
                  <Input v-model="form.notes" type="text" />
                </div>
                <DialogFooter class="mt-6">
                  <Button type="submit">
                    {{ isEditing ? t('globals.messages.save') : t('globals.messages.create') }}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </template>
      </DataTable>
    </template>
    <template #help>
      <div class="rounded-lg border bg-card p-4 shadow-xs space-y-2">
        <p class="text-sm leading-relaxed text-muted-foreground">
          {{ $t('admin.organization.help') }}
        </p>
      </div>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { Plus } from 'lucide-vue-next'
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import DataTable from '@main/components/datatable/DataTable.vue'
import { Button } from '@shared-ui/components/ui/button/index.js'
import { Input } from '@shared-ui/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@shared-ui/components/ui/dialog/index.js'
import { useEmitter } from '../../../composables/useEmitter.js'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useI18n } from 'vue-i18n'
import api from '../../../api/index.js'

const { t } = useI18n()
const isLoading = ref(true)
const orgs = ref([])
const dialogOpen = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const emitter = useEmitter()
const form = ref({ name: '', domainsText: '', notes: '' })

onMounted(getOrgs)

const columns = [
  {
    accessorKey: 'name',
    header: () => t('globals.terms.name'),
    cell: ({ row }) =>
      h(
        'button',
        {
          class: 'font-medium hover:underline text-left',
          onClick: () => editOrg(row.original)
        },
        row.getValue('name')
      )
  },
  {
    accessorKey: 'domains',
    header: () => t('organization.domains'),
    cell: ({ row }) => {
      const domains = row.original.domains || []
      return domains.length ? domains.join(', ') : t('organization.noDomains')
    }
  },
  {
    accessorKey: 'notes',
    header: () => t('globals.terms.note', 2),
    cell: ({ row }) => row.original.notes || '—'
  },
  {
    id: 'actions',
    enableHiding: false,
    enableSorting: false,
    cell: ({ row }) =>
      h(
        Button,
        {
          variant: 'ghost',
          size: 'sm',
          class: 'text-destructive',
          onClick: () => confirmDelete(row.original)
        },
        () => t('globals.messages.delete')
      )
  }
]

async function getOrgs() {
  isLoading.value = true
  try {
    const resp = await api.getOrganizations()
    orgs.value = resp.data.data || []
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

function newOrg() {
  editingId.value = null
  isEditing.value = false
  form.value = { name: '', domainsText: '', notes: '' }
}

function editOrg(org) {
  editingId.value = org.id
  isEditing.value = true
  form.value = {
    name: org.name,
    domainsText: (org.domains || []).join(', '),
    notes: org.notes || ''
  }
  dialogOpen.value = true
}

function parseDomains(text) {
  return text
    .split(/[,\s]+/)
    .map((d) => d.trim().replace(/^@/, ''))
    .filter(Boolean)
}

async function onSubmit() {
  isLoading.value = true
  const payload = {
    name: form.value.name.trim(),
    domains: parseDomains(form.value.domainsText),
    notes: form.value.notes
  }
  try {
    if (isEditing.value) {
      await api.updateOrganization(editingId.value, payload)
    } else {
      await api.createOrganization(payload)
    }
    dialogOpen.value = false
    await getOrgs()
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.savedSuccessfully')
    })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

async function confirmDelete(org) {
  if (!window.confirm(t('organization.deleteConfirmation'))) return
  try {
    await api.deleteOrganization(org.id)
    await getOrgs()
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  }
}
</script>
