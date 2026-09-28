<template>
  <DataTable :columns="createColumns(t)" :data="oidc" :loading="isLoading">
    <template #actions>
      <RouterLink :to="{ name: 'new-sso' }">
        <Button>
          <Plus class="size-4" />
          {{ $t('oidc.new') }}
        </Button>
      </RouterLink>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import DataTable from '@main/components/datatable/DataTable.vue'
import { createColumns } from '../../../features/admin/oidc/dataTableColumns.js'
import { Button } from '@shared-ui/components/ui/button'
import { useEmitter } from '../../../composables/useEmitter'
import { useI18n } from 'vue-i18n'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import api from '../../../api'

const oidc = ref([])
const { t } = useI18n()
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
  if (data?.model === 'oidc') fetchAll()
}

const fetchAll = async () => {
  try {
    isLoading.value = true
    const resp = await api.getAllOIDC()
    oidc.value = resp.data.data
  } finally {
    isLoading.value = false
  }
}
</script>
