<template>
  <Form v-slot="{ handleSubmit }" as="" keep-values :validation-schema="formSchema">
    <form @submit="handleSubmit($event, onSubmit)" class="space-y-6 w-full">
      <FormField v-slot="{ componentField }" name="provider">
        <FormItem>
          <FormLabel>{{ $t('admin.ai.provider') }}</FormLabel>
          <FormControl>
            <Input type="text" v-bind="componentField" disabled />
          </FormControl>
          <FormDescription>{{ $t('admin.ai.providerDescription') }}</FormDescription>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="model">
        <FormItem>
          <FormLabel>{{ $t('admin.ai.model') }}</FormLabel>
          <FormControl>
            <Input type="text" placeholder="gpt-4o-mini" v-bind="componentField" />
          </FormControl>
          <FormDescription>{{ $t('admin.ai.modelDescription') }}</FormDescription>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="api_key">
        <FormItem>
          <FormLabel>{{ $t('globals.terms.apiKey') }}</FormLabel>
          <FormControl>
            <Input type="password" placeholder="sk-..." v-bind="componentField" autocomplete="off" />
          </FormControl>
          <FormDescription>{{ $t('admin.ai.apiKeyDescription') }}</FormDescription>
          <FormMessage />
        </FormItem>
      </FormField>

      <Button type="submit" :is-loading="formLoading" :disabled="formLoading">
        {{ $t('globals.messages.save') }}
      </Button>
    </form>
  </Form>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { useI18n } from 'vue-i18n'
import { Button } from '@shared-ui/components/ui/button'
import { Input } from '@shared-ui/components/ui/input'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@shared-ui/components/ui/form'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'
import { useEmitter } from '@main/composables/useEmitter'
import { handleHTTPError } from '@shared-ui/utils/http.js'

const props = defineProps({
  initialValues: {
    type: Object,
    required: true
  },
  submitForm: {
    type: Function,
    required: true
  }
})

const { t } = useI18n()
const emitter = useEmitter()
const formLoading = ref(false)

const formSchema = toTypedSchema(
  z.object({
    provider: z.string().min(1),
    model: z.string().min(1, t('admin.ai.modelRequired')),
    api_key: z.string().optional()
  })
)

const form = useForm({
  validationSchema: formSchema
})

const onSubmit = async (values) => {
  try {
    formLoading.value = true
    await props.submitForm(values)
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
      variant: 'destructive',
      description: handleHTTPError(error).message
    })
  } finally {
    formLoading.value = false
  }
}

watch(
  () => props.initialValues,
  (newValues) => {
    if (!newValues || Object.keys(newValues).length === 0) return
    form.setValues(newValues)
  },
  { immediate: true, deep: true }
)
</script>
