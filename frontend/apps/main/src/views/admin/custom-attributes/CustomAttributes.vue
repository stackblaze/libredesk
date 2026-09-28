<template>
  <AdminSplitLayout>
    <template #content>
      <Dialog v-model:open="dialogOpen">
        <DialogContent class="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>
              {{ isEditing ? $t('customAttribute.edit') : $t('customAttribute.new') }}
            </DialogTitle>
            <DialogDescription />
          </DialogHeader>
          <CustomAttributesForm @submit.prevent="onSubmit" :form="form">
            <template #footer>
              <DialogFooter class="mt-10">
                <Button type="submit" :isLoading="isLoading">
                  {{ isEditing ? $t('globals.messages.save') : $t('globals.messages.create') }}
                </Button>
              </DialogFooter>
            </template>
          </CustomAttributesForm>
        </DialogContent>
      </Dialog>

      <Tabs default-value="contact" v-model="appliesTo">
        <TabsList class="grid w-full grid-cols-2 mb-4">
          <TabsTrigger value="contact">
            {{ $t('globals.terms.contact') }}
          </TabsTrigger>
          <TabsTrigger value="conversation">
            {{ $t('globals.terms.conversation') }}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="contact">
          <DataTable
            :columns="createColumns(t, { onEdit: editCustomAttribute })"
            :data="customAttributes"
            :loading="isLoading"
          >
            <template #actions>
              <Button @click="newCustomAttribute">
                <Plus class="size-4" />
                {{ $t('customAttribute.new') }}
              </Button>
            </template>
          </DataTable>
        </TabsContent>
        <TabsContent value="conversation">
          <DataTable
            :columns="createColumns(t, { onEdit: editCustomAttribute })"
            :data="customAttributes"
            :loading="isLoading"
          >
            <template #actions>
              <Button @click="newCustomAttribute">
                <Plus class="size-4" />
                {{ $t('customAttribute.new') }}
              </Button>
            </template>
          </DataTable>
        </TabsContent>
      </Tabs>
    </template>

    <template #help>
      <div class="rounded-lg border bg-card p-4 shadow-xs space-y-2">
        <p class="text-sm leading-relaxed text-muted-foreground">
          {{ $t('admin.customAttribute.help') }}
        </p>
      </div>
    </template>
  </AdminSplitLayout>
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { Plus } from 'lucide-vue-next'
import DataTable from '@main/components/datatable/DataTable.vue'
import { createColumns } from '../../../features/admin/custom-attributes/dataTableColumns.js'
import CustomAttributesForm from '@/features/admin/custom-attributes/CustomAttributesForm.vue'
import { Button } from '@shared-ui/components/ui/button'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { createFormSchema } from '../../../features/admin/custom-attributes/formSchema.js'
import { useEmitter } from '../../../composables/useEmitter'
import { EMITTER_EVENTS } from '../../../constants/emitterEvents.js'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@shared-ui/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@shared-ui/components/ui/tabs'
import { useStorage } from '@vueuse/core'
import AdminSplitLayout from '@/layouts/admin/AdminSplitLayout.vue'
import { useI18n } from 'vue-i18n'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '../../../api'

const appliesTo = useStorage('appliesTo', 'contact')
const { t } = useI18n()
const customAttributes = ref([])
const isLoading = ref(true)
const emitter = useEmitter()
const dialogOpen = ref(false)
const isEditing = ref(false)

const refreshHandler = (data) => {
  if (data?.model === 'custom-attributes') fetchAll()
}
const editHandler = (data) => {
  if (data?.model === 'custom-attributes') {
    form.setValues(data.data, false)
    form.setErrors({})
    isEditing.value = true
    dialogOpen.value = true
  }
}

onMounted(async () => {
  fetchAll()
  emitter.on(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
  emitter.on(EMITTER_EVENTS.EDIT_MODEL, editHandler)
})

onUnmounted(() => {
  emitter.off(EMITTER_EVENTS.REFRESH_LIST, refreshHandler)
  emitter.off(EMITTER_EVENTS.EDIT_MODEL, editHandler)
})

const editCustomAttribute = (item) => {
  form.setValues(item, false)
  form.setErrors({})
  isEditing.value = true
  dialogOpen.value = true
}

const newCustomAttribute = () => {
  form.resetForm()
  form.setErrors({})
  isEditing.value = false
  dialogOpen.value = true
}

const form = useForm({
  validationSchema: toTypedSchema(createFormSchema(t)),
  initialValues: {
    id: 0,
    name: '',
    data_type: 'text',
    applies_to: appliesTo.value,
    values: []
  }
})

const fetchAll = async () => {
  if (!appliesTo.value) return
  try {
    isLoading.value = true
    const resp = await api.getCustomAttributes(appliesTo.value)
    customAttributes.value = resp.data.data
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
}

const onSubmit = form.handleSubmit(async (values) => {
  try {
    isLoading.value = true
    if (values.id) {
      await api.updateCustomAttribute(values.id, values)
      emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
        description: t('globals.messages.savedSuccessfully')
      })
    } else {
      await api.createCustomAttribute(values)
      emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
        description: t('globals.messages.savedSuccessfully')
      })
    }

    dialogOpen.value = false
    fetchAll()
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    isLoading.value = false
  }
})

watch(
  appliesTo,
  (newVal) => {
    form.resetForm({
      values: {
        ...form.values,
        applies_to: newVal
      }
    })
    fetchAll()
  },
  { immediate: true }
)
</script>
