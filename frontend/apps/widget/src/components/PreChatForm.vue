<template>
  <div class="prechat-form bg-background flex-1 flex flex-col">
    <div v-if="showForm" class="flex-1 flex flex-col max-h-full">
      <div
        class="prechat-form__body flex-1 overflow-y-auto scrollbar-thin scrollbar-track-transparent scrollbar-thumb-muted-foreground/30 hover:scrollbar-thumb-muted-foreground/50 px-4 py-4"
      >
        <!-- Form title -->
        <div v-if="formTitle" class="prechat-form__title text-foreground mb-4 text-center">
          {{ formTitle }}
        </div>

        <form ref="formRef" @submit.prevent="submitForm" class="prechat-form__fields space-y-3.5">
          <!-- Dynamic fields -->
          <div v-for="field in sortedFields" :key="field.key" class="prechat-form__field">
            <!-- Text input -->
            <FormField v-if="field.type === 'text'" v-slot="{ componentField }" :name="field.key">
              <FormItem>
                <FormLabel :class="fieldLabelClass">
                  {{ field.label }}
                  <span v-if="field.required" class="text-destructive/80">*</span>
                </FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="text"
                    :class="fieldInputClass"
                    :placeholder="field.placeholder || ''"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <!-- Email input -->
            <FormField
              v-else-if="field.type === 'email'"
              v-slot="{ componentField }"
              :name="field.key"
            >
              <FormItem>
                <FormLabel :class="fieldLabelClass">
                  {{ field.label }}
                  <span v-if="field.required" class="text-destructive/80">*</span>
                </FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="email"
                    :class="fieldInputClass"
                    :placeholder="field.placeholder || ''"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <!-- Number input -->
            <FormField
              v-else-if="field.type === 'number'"
              v-slot="{ componentField }"
              :name="field.key"
            >
              <FormItem>
                <FormLabel :class="fieldLabelClass">
                  {{ field.label }}
                  <span v-if="field.required" class="text-destructive/80">*</span>
                </FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="number"
                    :class="fieldInputClass"
                    :placeholder="field.placeholder || ''"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <!-- Date input -->
            <FormField
              v-else-if="field.type === 'date'"
              v-slot="{ componentField }"
              :name="field.key"
            >
              <FormItem>
                <FormLabel :class="fieldLabelClass">
                  {{ field.label }}
                  <span v-if="field.required" class="text-destructive/80">*</span>
                </FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="date"
                    :class="fieldInputClass"
                    :placeholder="field.placeholder || ''"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <!-- Link/URL input -->
            <FormField
              v-else-if="field.type === 'link'"
              v-slot="{ componentField }"
              :name="field.key"
            >
              <FormItem>
                <FormLabel :class="fieldLabelClass">
                  {{ field.label }}
                  <span v-if="field.required" class="text-destructive/80">*</span>
                </FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="url"
                    :class="fieldInputClass"
                    :placeholder="field.placeholder || 'https://'"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <!-- Checkbox input -->
            <FormField
              v-else-if="field.type === 'checkbox'"
              v-slot="{ componentField, handleChange }"
              :name="field.key"
            >
              <FormItem class="flex flex-row items-start space-x-3 space-y-0">
                <FormControl>
                  <Checkbox :checked="componentField.modelValue" @update:checked="handleChange" />
                </FormControl>
                <div class="space-y-1 leading-none">
                  <FormLabel class="prechat-form__checkbox-label">
                    {{ field.label }}
                    <span v-if="field.required" class="text-destructive/80">*</span>
                  </FormLabel>
                  <FormMessage />
                </div>
              </FormItem>
            </FormField>

            <!-- List/Select input -->
            <FormField
              v-else-if="field.type === 'list'"
              v-slot="{ componentField }"
              :name="field.key"
            >
              <FormItem>
                <FormLabel :class="fieldLabelClass">
                  {{ field.label }}
                  <span v-if="field.required" class="text-destructive/80">*</span>
                </FormLabel>
                <FormControl>
                  <Select v-bind="componentField">
                    <SelectTrigger :class="fieldInputClass">
                      <SelectValue :placeholder="field.placeholder || $t('globals.terms.select')" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem
                        v-for="option in getFieldOptions(field)"
                        :key="option.value"
                        :value="option.value"
                      >
                        {{ option.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>
        </form>
      </div>

      <!-- Submit button - fixed at bottom -->
      <div class="prechat-form__footer px-4 py-3 border-t border-border/40">
        <Button
          variant="outline"
          type="submit"
          @click="submitForm"
          class="prechat-form__submit w-full shadow-none cursor-pointer disabled:cursor-not-allowed disabled:pointer-events-auto disabled:opacity-100"
          :disabled="!requiredFieldsFilled || !meta.valid || props.isSubmitting"
        >
          <div
            v-if="props.isSubmitting"
            class="w-4 h-4 border-2 border-background border-t-current rounded-full animate-spin mr-2"
          ></div>
          {{ $t('widget.prechatForm.startChat') }}
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { Button } from '@shared-ui/components/ui/button'
import { Input } from '@shared-ui/components/ui/input'
import { Checkbox } from '@shared-ui/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@shared-ui/components/ui/select'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@shared-ui/components/ui/form'
import { useWidgetStore } from '../store/widget.js'
import { useI18n } from 'vue-i18n'
import { createPreChatFormSchema } from './preChatFormSchema.js'

const props = defineProps({
  excludeDefaultFields: {
    type: Boolean,
    default: false
  },
  isSubmitting: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['submit'])
const { t } = useI18n()
const widgetStore = useWidgetStore()
const formRef = ref(null)

const fieldLabelClass = 'prechat-form__label'
const fieldInputClass = 'prechat-form__input'

const config = computed(() => widgetStore.config?.prechat_form || {})
const preChatFormEnabled = computed(() => config.value.enabled || false)
const formTitle = computed(() => config.value.title || '')
const formFields = computed(() => config.value.fields || [])

// Sort and filter enabled fields, excluding default name/email for an already
// identified visitor (verified user or a returning visitor we recognize).
const sortedFields = computed(() => {
  let fields = formFields.value.filter((field) => field.enabled)

  if (props.excludeDefaultFields) {
    fields = fields.filter((field) => !['name', 'email'].includes(field.key))
  }

  // When we do ask for name/email (an unidentified visitor), they are always
  // required, regardless of the admin's per-field setting.
  fields = fields.map((field) =>
    ['name', 'email'].includes(field.key) ? { ...field, required: true } : field
  )

  return fields.sort((a, b) => (a.order || 0) - (b.order || 0))
})

const showForm = computed(() => preChatFormEnabled.value && sortedFields.value.length > 0)

// Create form with dynamic schema based on fields
const formSchema = computed(() => toTypedSchema(createPreChatFormSchema(t, sortedFields.value)))

// Generate initial values dynamically
const initialValues = computed(() => {
  const values = {}
  sortedFields.value.forEach((field) => {
    if (field.type === 'checkbox') {
      values[field.key] = false
    } else {
      values[field.key] = ''
    }
  })
  return values
})

const { handleSubmit, meta, values } = useForm({
  validationSchema: formSchema,
  initialValues
})

const requiredFieldsFilled = computed(() => {
  return sortedFields.value
    .filter((field) => field.required)
    .every((field) => {
      const value = values[field.key]
      if (field.type === 'checkbox') return true
      return value && String(value).trim() !== ''
    })
})

const submitForm = handleSubmit((values) => {
  // Filter out empty values (except for checkboxes)
  const filteredValues = {}
  Object.keys(values).forEach((key) => {
    const field = sortedFields.value.find((f) => f.key === key)
    if (field?.type === 'checkbox' || (values[key] && String(values[key]).trim())) {
      filteredValues[key] = values[key]
    }
  })

  emit('submit', { formData: filteredValues })
})

// Get options for list fields
const getFieldOptions = (field) => {
  if (field.type === 'list' && field.custom_attribute_id) {
    const customAttr = widgetStore.config?.custom_attributes?.[field.custom_attribute_id]
    if (customAttr?.values) {
      return customAttr.values.map((value) => ({
        value: value,
        label: value
      }))
    }
  }
  return []
}

const focusFirstField = () => {
  nextTick(() => {
    const firstInput = formRef.value?.querySelector('input, textarea, select')
    firstInput?.focus()
  })
}

onMounted(focusFirstField)
watch(() => widgetStore.isOpen, (open) => {
  if (open) focusFirstField()
})

// Auto-submit when no fields to show (e.g., all fields excluded)
watch(
  showForm,
  (newValue) => {
    if (!newValue) {
      emit('submit', { formData: {} })
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.prechat-form__title {
  font-family: 'Space Grotesk', system-ui, sans-serif;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: -0.03em;
  line-height: 1.3;
}

.prechat-form__field :deep(.prechat-form__label) {
  margin-bottom: 0.375rem;
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.12em;
  line-height: 1.4;
  text-transform: uppercase;
  color: hsl(var(--muted-foreground) / 0.72);
}

.prechat-form__field :deep(.prechat-form__label.text-destructive) {
  color: hsl(var(--destructive));
  text-transform: uppercase;
}

.prechat-form__field :deep(.prechat-form__checkbox-label) {
  font-size: 0.875rem;
  font-weight: 400;
  line-height: 1.45;
  letter-spacing: 0;
  text-transform: none;
  color: hsl(var(--foreground) / 0.82);
}

.prechat-form__field :deep(.prechat-form__input) {
  height: 2.25rem;
  min-height: 2.25rem;
  border-radius: 0.5rem;
  border: 1px solid hsl(var(--border) / 0.9);
  background-color: hsl(var(--muted) / 0.22);
  padding-inline: 0.75rem;
  font-size: 0.875rem;
  font-weight: 400;
  line-height: 1.35;
  color: hsl(var(--foreground));
  box-shadow: none;
  cursor: text;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease,
    box-shadow 0.15s ease;
}

@media (max-width: 639px) {
  .prechat-form__field :deep(.prechat-form__input) {
    font-size: 16px;
  }
}

.prechat-form__field :deep(.prechat-form__input:hover:not(:disabled):not(:focus-visible)) {
  border-color: hsl(var(--primary) / 0.28);
  background-color: hsl(var(--muted) / 0.32);
}

.prechat-form__field :deep(.prechat-form__input::placeholder) {
  color: hsl(var(--muted-foreground) / 0.45);
  font-weight: 400;
}

.prechat-form__field :deep(.prechat-form__input:focus-visible) {
  border-color: hsl(var(--primary) / 0.5);
  background-color: hsl(var(--muted) / 0.28);
  outline: none;
  box-shadow: 0 0 0 2px hsl(var(--primary) / 0.18);
}

.prechat-form__field :deep(p.text-destructive) {
  margin-top: 0.125rem;
  font-size: 0.75rem;
  font-weight: 400;
  letter-spacing: 0;
  text-transform: none;
}

.prechat-form__footer :deep(.prechat-form__submit) {
  height: 2.375rem;
  min-height: 2.375rem;
  border-radius: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background-color: rgba(255, 255, 255, 0.06);
  font-family: 'Inter', system-ui, sans-serif;
  font-size: 0.875rem;
  font-weight: 600;
  letter-spacing: 0;
  text-transform: none;
  color: hsl(var(--foreground));
  box-shadow: none;
  text-shadow: none;
  backdrop-filter: none;
  -webkit-font-smoothing: antialiased;
  transform: none;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease;
}

.prechat-form__footer :deep(.prechat-form__submit:not(:disabled):hover) {
  border-color: rgba(255, 255, 255, 0.2);
  background-color: rgba(255, 255, 255, 0.1);
  color: hsl(var(--foreground));
  box-shadow: none;
}

.prechat-form__footer :deep(.prechat-form__submit:not(:disabled):active) {
  border-color: rgba(255, 255, 255, 0.16);
  background-color: rgba(255, 255, 255, 0.08);
  box-shadow: none;
  transform: none;
}

.prechat-form__footer :deep(.prechat-form__submit:disabled) {
  opacity: 1;
  border-color: rgba(255, 255, 255, 0.06);
  background-color: rgba(255, 255, 255, 0.03);
  color: hsl(var(--muted-foreground) / 0.65);
  box-shadow: none;
  transform: none;
}
</style>
