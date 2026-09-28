<template>
  <div class="space-y-4">
    <AdminEmptyState
      v-if="!isLoading && !data.length"
      :icon="Inbox"
      :title="t('admin.inbox.emptyTitle')"
      :description="t('admin.inbox.emptyDescription')"
    >
      <template #action>
        <router-link :to="{ name: 'new-inbox' }">
          <Button>
            <Plus class="size-4" />
            {{ t('inbox.newInbox') }}
          </Button>
        </router-link>
      </template>
    </AdminEmptyState>

    <DataTable
      v-else
      :columns="columns"
      :data="data"
      :loading="isLoading"
      :empty-text="t('admin.inbox.emptyTitle')"
    >
      <template #actions>
        <router-link :to="{ name: 'new-inbox' }">
          <Button>
            <Plus class="size-4" />
            {{ t('inbox.newInbox') }}
          </Button>
        </router-link>
      </template>
    </DataTable>
  </div>
</template>

<script setup>
import { onMounted, ref, h } from 'vue'
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { format } from 'date-fns'
import { Inbox, Plus } from 'lucide-vue-next'
import InboxDataTableDropDown from '@main/features/admin/inbox/InboxDataTableDropDown.vue'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { Button } from '@shared-ui/components/ui/button'
import { Badge } from '@shared-ui/components/ui/badge/index.js'
import DataTable from '@main/components/datatable/DataTable.vue'
import AdminEmptyState from '@main/components/layout/AdminEmptyState.vue'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'
import { useEmitter } from '@main/composables/useEmitter'
import { useInboxStore } from '@main/stores/inbox'
import api from '@main/api'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const emitter = useEmitter()
const inboxStore = useInboxStore()
const isLoading = ref(true)
const data = ref([])

onMounted(async () => {
  const errorCode = route.query.error
  const successCode = route.query.success

  if (errorCode) {
    let msg
    if (errorCode === 'oauth_denied') {
      msg = t('toast.authorizationDenied')
    } else if (errorCode === 'inbox_already_exists') {
      msg = t('inbox.oauthAlreadyExists')
    } else if (errorCode === 'inbox_not_found') {
      msg = t('inbox.oauthNotFound')
    } else if (errorCode === 'email_mismatch') {
      msg = t('inbox.oauthEmailMismatch')
    } else {
      msg = t('toast.errorConnectingInbox')
    }
    setTimeout(() => {
      emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
        variant: 'destructive',
        description: msg
      })
    }, 500)
  } else if (successCode) {
    const msg =
      successCode === 'oauth_reconnected'
        ? t('toast.inboxReconnected')
        : t('toast.inboxConnected')
    setTimeout(() => {
      emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { description: msg })
    }, 500)
  }

  await getInboxes()
})

const getInboxes = async () => {
  try {
    isLoading.value = true
    await inboxStore.fetchInboxes(true)
    data.value = inboxStore.inboxes
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

const columns = [
  {
    accessorKey: 'name',
    header: () => t('globals.terms.name'),
    cell: ({ row }) =>
      h(
        RouterLink,
        {
          to: { name: 'edit-inbox', params: { id: row.original.id } },
          class: 'text-foreground font-medium hover:underline'
        },
        () => row.getValue('name')
      )
  },
  {
    accessorKey: 'channel',
    header: () => t('globals.terms.channel'),
    cell: ({ row }) => row.getValue('channel')
  },
  {
    accessorKey: 'enabled',
    header: () => t('globals.terms.status'),
    cell: ({ row }) =>
      h(Badge, { variant: row.getValue('enabled') ? 'success' : 'secondary' }, () =>
        row.getValue('enabled') ? t('globals.terms.enabled') : t('globals.terms.disabled')
      )
  },
  {
    accessorKey: 'created_at',
    header: () => t('globals.terms.createdAt'),
    cell: ({ row }) => format(row.getValue('created_at'), 'PPpp')
  },
  {
    accessorKey: 'updated_at',
    header: () => t('globals.terms.updatedAt'),
    cell: ({ row }) => format(row.getValue('updated_at'), 'PPpp')
  },
  {
    id: 'actions',
    enableHiding: false,
    enableSorting: false,
    cell: ({ row }) =>
      h(InboxDataTableDropDown, {
        inbox: row.original,
        onEditInbox: (id) => handleEditInbox(id),
        onDeleteInbox: (id) => handleDeleteInbox(id),
        onToggleInbox: (id) => handleToggleInbox(id)
      })
  }
]

const handleEditInbox = (id) => {
  router.push({ path: `/admin/inboxes/${id}/edit` })
}

const handleDeleteInbox = async (id) => {
  await api.deleteInbox(id)
  getInboxes()
}

const handleToggleInbox = async (id) => {
  await api.toggleInbox(id)
  getInboxes()
}
</script>
