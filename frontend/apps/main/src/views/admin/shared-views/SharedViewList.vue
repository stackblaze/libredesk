<template>
  <DataTable :columns="createColumns(t)" :data="sharedViews" :loading="formLoading">
    <template #actions>
      <router-link :to="{ name: 'new-shared-view' }">
        <Button>
          <Plus class="size-4" />
          {{ $t('sharedView.new') }}
        </Button>
      </router-link>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import DataTable from '@/components/datatable/DataTable.vue'
import { createColumns } from '@/features/admin/shared-views/dataTableColumns.js'
import { useEmitter } from '@/composables/useEmitter'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { Button } from '@shared-ui/components/ui/button'
import { useI18n } from 'vue-i18n'
import api from '@/api'

const { t } = useI18n()
const formLoading = ref(true)
const sharedViews = ref([])
const emit = useEmitter()

onMounted(() => {
  getSharedViews()
  emit.on(EMITTER_EVENTS.REFRESH_LIST, refreshList)
})

onUnmounted(() => {
  emit.off(EMITTER_EVENTS.REFRESH_LIST, refreshList)
})

const refreshList = (data) => {
  if (data?.model === 'shared-views') getSharedViews()
}

const getSharedViews = async () => {
  try {
    formLoading.value = true
    const resp = await api.getAllSharedViews()
    sharedViews.value = resp.data.data
  } catch (error) {
    emit.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    formLoading.value = false
  }
}
</script>
