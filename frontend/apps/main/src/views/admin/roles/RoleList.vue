<template>
  <DataTable :columns="createColumns(t)" :data="roles" :loading="isLoading">
    <template #actions>
      <router-link :to="{ name: 'new-role' }">
        <Button>
          <Plus class="size-4" />
          {{ $t('role.new') }}
        </Button>
      </router-link>
    </template>
  </DataTable>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { Plus } from 'lucide-vue-next'
import { createColumns } from '../../../features/admin/roles/dataTableColumns.js'
import { Button } from '@shared-ui/components/ui/button'
import DataTable from '@main/components/datatable/DataTable.vue'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useEmitter } from '../../../composables/useEmitter'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import { useI18n } from 'vue-i18n'
import api from '../../../api'

const emitter = useEmitter()
const { t } = useI18n()
const roles = ref([])
const isLoading = ref(true)

const getRoles = async () => {
  try {
    isLoading.value = true
    const resp = await api.getRoles()
    roles.value = resp.data.data
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

const refreshHandler = (data) => {
  if (data?.model === 'team') getRoles()
}

onMounted(async () => {
  getRoles()
  emitter.on(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
})

onUnmounted(() => {
  emitter.off(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
})
</script>
