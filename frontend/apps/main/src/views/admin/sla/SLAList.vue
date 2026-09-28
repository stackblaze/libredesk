<template>
  <DataTable :columns="createColumns(t)" :data="slas" :loading="isLoading">
    <template #actions>
      <router-link :to="{ name: 'new-sla' }">
        <Button>
          <Plus class="size-4" />
          {{ t('sla.new') }}
        </Button>
      </router-link>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import DataTable from '@main/components/datatable/DataTable.vue'
import { createColumns } from '../../../features/admin/sla/dataTableColumns.js'
import { Button } from '@shared-ui/components/ui/button'
import { useEmitter } from '../../../composables/useEmitter'
import { useI18n } from 'vue-i18n'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import api from '../../../api'

const { t } = useI18n()
const slas = ref([])
const isLoading = ref(true)
const emit = useEmitter()

onMounted(() => {
  fetchAll()
  emit.on(EMITTER_EVENTS.REFRESH_LIST, refreshList)
})

onUnmounted(() => {
  emit.off(EMITTER_EVENTS.REFRESH_LIST, refreshList)
})

const refreshList = (data) => {
  if (data?.model === 'sla') fetchAll()
}

const fetchAll = async () => {
  try {
    isLoading.value = true
    const resp = await api.getAllSLAs()
    slas.value = resp.data.data
  } finally {
    isLoading.value = false
  }
}
</script>
