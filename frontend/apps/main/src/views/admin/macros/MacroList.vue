<template>
  <DataTable :columns="createColumns(t)" :data="macros" :loading="formLoading">
    <template #actions>
      <router-link :to="{ name: 'new-macro' }">
        <Button>
          <Plus class="size-4" />
          {{ $t('macro.new') }}
        </Button>
      </router-link>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import DataTable from '@main/components/datatable/DataTable.vue'
import { createColumns } from '../../../features/admin/macros/dataTableColumns.js'
import { useEmitter } from '../../../composables/useEmitter'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { Button } from '@shared-ui/components/ui/button'
import { useI18n } from 'vue-i18n'
import api from '../../../api'

const { t } = useI18n()
const formLoading = ref(true)
const macros = ref([])
const emit = useEmitter()

onMounted(() => {
  getMacros()
  emit.on(EMITTER_EVENTS.REFRESH_LIST, refreshList)
})

onUnmounted(() => {
  emit.off(EMITTER_EVENTS.REFRESH_LIST, refreshList)
})

const refreshList = (data) => {
  if (data?.model === 'macros') getMacros()
}

const getMacros = async () => {
  try {
    formLoading.value = true
    const resp = await api.getAllMacros()
    macros.value = resp.data.data
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
