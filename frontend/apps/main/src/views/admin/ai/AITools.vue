<template>
  <AdminSplitLayout>
    <template #content>
      <DataTable
        :columns="createToolColumns(t, { onEdit: editTool })"
        :data="tools"
        :loading="isLoading"
      >
        <template #actions>
          <Button @click="router.push({ name: 'new-ai-tool' })">
            <Plus class="size-4" />
            {{ t('admin.ai.tool.new') }}
          </Button>
        </template>
      </DataTable>
    </template>

    <template #help>
      <AdminHelpCard>
        <p class="text-sm leading-relaxed text-muted-foreground">{{ t('admin.ai.toolsHelp') }}</p>
        <a
          href="https://docs.libredesk.io/configuration/ai#tools"
          target="_blank"
          rel="noopener noreferrer"
          class="link-style text-sm"
        >
          {{ t('globals.terms.learnMore') }}
        </a>
      </AdminHelpCard>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import AdminHelpCard from '@main/components/layout/AdminHelpCard.vue'
import DataTable from '@main/components/datatable/DataTable.vue'
import { Button } from '@shared-ui/components/ui/button/index.js'
import { createToolColumns } from '@/features/admin/ai/toolColumns.js'
import { useEmitter } from '@/composables/useEmitter.js'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import api from '@/api'

const { t } = useI18n()
const emitter = useEmitter()
const router = useRouter()
const isLoading = ref(true)
const tools = ref([])

const refreshHandler = (data) => {
  if (data?.model === 'ai_tools') getTools()
}

const editHandler = (data) => {
  if (data?.model === 'ai_tools') editTool(data.data)
}

const editTool = (item) => {
  router.push({ name: 'edit-ai-tool', params: { id: item.id } })
}

onMounted(() => {
  getTools()
  emitter.on(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
  emitter.on(EMITTER_EVENTS.EDIT_MODEL, editHandler)
})

onUnmounted(() => {
  emitter.off(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
  emitter.off(EMITTER_EVENTS.EDIT_MODEL, editHandler)
})

const getTools = async () => {
  try {
    isLoading.value = true
    const resp = await api.getAITools()
    tools.value = resp.data.data || []
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
