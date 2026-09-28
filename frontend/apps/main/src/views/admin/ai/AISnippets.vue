<template>
  <AdminSplitLayout>
    <template #content>
      <Dialog v-model:open="importDialogOpen">
        <DialogContent class="sm:max-w-[560px]">
          <DialogHeader>
            <DialogTitle>{{ t('admin.ai.snippet.importUrl') }}</DialogTitle>
          </DialogHeader>
          <form class="space-y-4" @submit.prevent="importFromUrl">
            <Input
              v-model="importUrl"
              type="url"
              :placeholder="t('admin.ai.snippet.importUrlPlaceholder')"
            />
            <div class="flex justify-end">
              <Button type="submit" :isLoading="isImporting" :disabled="!importUrl.trim()">
                {{ t('globals.messages.import') }}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog v-model:open="snippetDialogOpen">
        <DialogContent class="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>
              {{ snippetEditing ? t('admin.ai.snippet.edit') : t('admin.ai.snippet.new') }}
            </DialogTitle>
          </DialogHeader>
          <SnippetForm
            :initial-values="snippetInitial"
            :is-editing="snippetEditing"
            :submit-form="submitSnippet"
          />
        </DialogContent>
      </Dialog>

      <DataTable
        :columns="createSnippetColumns(t, { onEdit: editSnippet })"
        :data="snippets"
        :loading="isLoading"
      >
        <template #actions>
          <Button variant="outline" @click="importDialogOpen = true">
            {{ t('admin.ai.snippet.importUrl') }}
          </Button>
          <Button @click="newSnippet">
            <Plus class="size-4" />
            {{ t('admin.ai.snippet.new') }}
          </Button>
        </template>
      </DataTable>
    </template>

    <template #help>
      <div class="rounded-lg border bg-card p-4 shadow-xs space-y-2">
        <p class="text-sm leading-relaxed text-muted-foreground">{{ t('admin.ai.snippetsHelp') }}</p>
        <a
          href="https://docs.libredesk.io/configuration/ai#snippets"
          target="_blank"
          rel="noopener noreferrer"
          class="link-style text-sm"
        >
          {{ t('globals.terms.learnMore') }}
        </a>
      </div>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import DataTable from '@main/components/datatable/DataTable.vue'
import { Button } from '@shared-ui/components/ui/button/index.js'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle
} from '@shared-ui/components/ui/dialog/index.js'
import { Input } from '@shared-ui/components/ui/input/index.js'
import SnippetForm from '@/features/admin/ai/SnippetForm.vue'
import { createSnippetColumns } from '@/features/admin/ai/snippetColumns.js'
import { useEmitter } from '@/composables/useEmitter.js'
import { EMITTER_EVENTS } from '@/constants/emitterEvents.js'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useI18n } from 'vue-i18n'
import api from '@/api'

const { t } = useI18n()
const emitter = useEmitter()
const isLoading = ref(true)

const snippets = ref([])

const snippetDialogOpen = ref(false)
const importDialogOpen = ref(false)
const importUrl = ref('')
const isImporting = ref(false)
const snippetEditing = ref(false)
const snippetInitial = ref({})
const editingSnippetId = ref(null)

const refreshHandler = (data) => {
  if (data?.model === 'ai_snippets') getSnippets()
}

const editHandler = (data) => {
  if (data?.model === 'ai_snippets') editSnippet(data.data)
}

onMounted(() => {
  getSnippets()
  emitter.on(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
  emitter.on(EMITTER_EVENTS.EDIT_MODEL, editHandler)
})

onUnmounted(() => {
  emitter.off(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
  emitter.off(EMITTER_EVENTS.EDIT_MODEL, editHandler)
})

const getSnippets = async () => {
  try {
    isLoading.value = true
    const resp = await api.getAISnippets()
    snippets.value = resp.data.data || []
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

const importFromUrl = async () => {
  if (!importUrl.value.trim() || isImporting.value) return
  try {
    isImporting.value = true
    await api.importAISnippetFromURL({ url: importUrl.value.trim() })
    importDialogOpen.value = false
    importUrl.value = ''
    getSnippets()
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.savedSuccessfully')
    })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isImporting.value = false
  }
}

const newSnippet = () => {
  snippetEditing.value = false
  editingSnippetId.value = null
  snippetInitial.value = {}
  snippetDialogOpen.value = true
}

const editSnippet = (item) => {
  snippetEditing.value = true
  editingSnippetId.value = item.id
  snippetInitial.value = { ...item }
  snippetDialogOpen.value = true
}

const submitSnippet = async (values) => {
  try {
    if (snippetEditing.value) {
      await api.updateAISnippet(editingSnippetId.value, values)
    } else {
      await api.createAISnippet(values)
    }
    snippetDialogOpen.value = false
    getSnippets()
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      description: t('globals.messages.savedSuccessfully')
    })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  }
}
</script>
