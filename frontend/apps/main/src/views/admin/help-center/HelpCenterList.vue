<template>
  <AdminSplitLayout>
    <template #content>
      <AdminEmptyState
        v-if="!loading && !helpCenters.length"
        :icon="BookOpen"
        :title="t('admin.helpCenter.emptyTitle')"
        :description="t('admin.helpCenter.help')"
      >
        <template #action>
          <Button @click="openCreateModal">
            <Plus class="size-4" />
            {{ t('helpCenter.new') }}
          </Button>
        </template>
      </AdminEmptyState>

      <DataTable v-else :columns="columns" :data="helpCenters" :loading="loading">
        <template #actions>
          <Button @click="openCreateModal">
            <Plus class="size-4" />
            {{ t('helpCenter.new') }}
          </Button>
        </template>
      </DataTable>
    </template>

    <template #help>
      <AdminHelpCard>
        <p class="text-sm leading-relaxed text-muted-foreground">
          {{ t('admin.helpCenter.help') }}
        </p>
      </AdminHelpCard>
    </template>
  </AdminSplitLayout>

  <Sheet :open="showCreateModal" @update:open="closeCreateModal">
    <SheetContent class="sm:max-w-lg overflow-y-auto">
      <SheetHeader>
        <SheetTitle>{{ t('helpCenter.new') }}</SheetTitle>
      </SheetHeader>

      <HelpCenterBasicsForm
        :submit-form="handleSave"
        :is-loading="isSubmitting"
        @cancel="closeCreateModal"
      />
    </SheetContent>
  </Sheet>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useEmitter } from '@/composables/useEmitter.js'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { BookOpen, Plus } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@shared-ui/components/ui/sheet'
import DataTable from '@main/components/datatable/DataTable.vue'
import AdminEmptyState from '@main/components/layout/AdminEmptyState.vue'
import AdminHelpCard from '@main/components/layout/AdminHelpCard.vue'
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import { createHelpCenterColumns } from '@/features/admin/help-center/helpCenterColumns.js'
import HelpCenterBasicsForm from '@/features/admin/help-center/HelpCenterBasicsForm.vue'
import api from '@/api'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const emitter = useEmitter()
const { t } = useI18n()
const loading = ref(true)
const isSubmitting = ref(false)
const helpCenters = ref([])
const showCreateModal = ref(false)

onMounted(() => {
  fetchHelpCenters()
})

const fetchHelpCenters = async () => {
  try {
    loading.value = true
    const { data } = await api.getHelpCenters()
    helpCenters.value = data.data || []
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    loading.value = false
  }
}

const goToTree = (helpCenterId) => {
  router.push({ name: 'help-center-tree', params: { id: helpCenterId } })
}

const openCreateModal = () => {
  showCreateModal.value = true
}

const openEditModal = (helpCenter) => {
  router.push({ name: 'help-center-customize', params: { id: helpCenter.id } })
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const handleSave = async (formData) => {
  try {
    isSubmitting.value = true
    const { data } = await api.createHelpCenter(formData)
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.savedSuccessfully')
    })
    closeCreateModal()
    router.push({ name: 'help-center-customize', params: { id: data.data.id } })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isSubmitting.value = false
  }
}

const handleToggle = async (helpCenter) => {
  try {
    await api.toggleHelpCenter(helpCenter.id)
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.savedSuccessfully')
    })
    fetchHelpCenters()
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  }
}

const handleDelete = async (helpCenter) => {
  try {
    await api.deleteHelpCenter(helpCenter.id)
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.deletedSuccessfully')
    })
    fetchHelpCenters()
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  }
}

const columns = createHelpCenterColumns(t, {
  onOpen: (helpCenter) => goToTree(helpCenter.id),
  onEdit: openEditModal,
  onDelete: handleDelete,
  onToggle: handleToggle
})
</script>
