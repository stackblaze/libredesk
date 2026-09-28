<template>
  <div class="flex gap-1">
    <Button
      v-for="option in [7, 30, 90]"
      :key="option"
      size="sm"
      :variant="modelValue === option ? 'default' : 'outline'"
      @click="emit('update:modelValue', option)"
    >
      {{ t('globals.messages.nDays', { days: option }) }}
    </Button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { Button } from '@shared-ui/components/ui/button'

defineProps({
  modelValue: { type: Number, required: true }
})
const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()
</script>
