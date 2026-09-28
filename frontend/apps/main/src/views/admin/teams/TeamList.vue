<template>
  <DataTable :columns="columns" :data="data" :loading="isLoading">
    <template #actions>
      <router-link :to="{ name: 'new-team' }">
        <Button>
          <Plus class="size-4" />
          {{ $t('globals.messages.new') }}
        </Button>
      </router-link>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { columns } from '../../../features/admin/teams/TeamsDataTableColumns.js'
import { Button } from '@shared-ui/components/ui/button'
import { useEmitter } from '../../../composables/useEmitter'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import DataTable from '@main/components/datatable/DataTable.vue'
import api from '../../../api'

const emitter = useEmitter()
const data = ref([])
const isLoading = ref(true)

const getData = async () => {
  try {
    isLoading.value = true
    const response = await api.getTeams()
    data.value = response.data.data
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      title: 'Error',
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

const refreshHandler = (event) => {
  if (event.model === 'team') {
    getData()
  }
}

onMounted(async () => {
  getData()
  emitter.on(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
})

onUnmounted(() => {
  emitter.off(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
})
</script>
