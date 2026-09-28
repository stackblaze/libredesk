<template>
  <DataTable :columns="createColumns(t)" :data="data" :loading="isLoading">
    <template #actions>
      <Importer
        entity-key="globals.terms.agent"
        :upload-fn="api.importAgents"
        :get-status-fn="api.getAgentImportStatus"
        @import-complete="getData"
      >
        <template #csv-example>
          <div class="bg-muted p-3 rounded-md text-xs font-mono overflow-x-auto leading-relaxed">
            <div>first_name,last_name,email,roles,teams</div>
            <div>John,Doe,john@example.com,Agent,Sales</div>
            <div>Jane,Smith,jane@example.com,Admin,Support</div>
            <div>Bob,Test,bob@example.com,"Agent,Admin",Support</div>
          </div>
          <p class="text-xs mt-2 text-muted-foreground">
            {{ $t('importer.agentCaseSensitiveNote') }}
          </p>
        </template>
      </Importer>
      <router-link :to="{ name: 'new-agent' }">
        <Button>
          <Plus class="size-4" />
          {{ $t('agent.new') }}
        </Button>
      </router-link>
    </template>
  </DataTable>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { Plus } from 'lucide-vue-next'
import { createColumns } from '@/features/admin/agents/dataTableColumns.js'
import { Button } from '@shared-ui/components/ui/button'
import DataTable from '@/components/datatable/DataTable.vue'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useEmitter } from '@/composables/useEmitter'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { useI18n } from 'vue-i18n'
import Importer from '@/components/importer/Importer.vue'
import api from '@/api'

const isLoading = ref(true)
const { t } = useI18n()
const data = ref([])
const emitter = useEmitter()

const refreshHandler = (payload) => {
  if (payload?.model === 'agent') getData()
}

onMounted(async () => {
  getData()
  emitter.on(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
})

onUnmounted(() => {
  emitter.off(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
})

const getData = async () => {
  try {
    isLoading.value = true
    const response = await api.getUsers()
    data.value = response?.data?.data || []
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}
</script>
