<template>
  <DataTable :columns="createColumns(t)" :data="webhooks" :loading="isLoading">
    <template #actions>
      <RouterLink :to="{ name: 'new-webhook' }">
        <Button>
          <Plus class="size-4" />
          {{ $t('webhook.new') }}
        </Button>
      </RouterLink>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import DataTable from '@main/components/datatable/DataTable.vue'
import { createColumns } from '../../../features/admin/webhooks/dataTableColumns.js'
import { Button } from '@shared-ui/components/ui/button'
import { useEmitter } from '../../../composables/useEmitter'
import { useI18n } from 'vue-i18n'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import api from '../../../api'

const webhooks = ref([])
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
  if (data?.model === 'webhook') fetchAll()
}

const fetchAll = async () => {
  try {
    isLoading.value = true
    const resp = await api.getWebhooks()
    webhooks.value = resp.data.data
  } finally {
    isLoading.value = false
  }
}
</script>
