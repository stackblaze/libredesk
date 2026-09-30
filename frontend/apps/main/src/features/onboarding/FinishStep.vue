<template>
  <div class="space-y-6 max-w-xl">
    <ul class="space-y-2">
      <li v-for="step in steps" :key="step.key" class="flex items-center gap-3 text-sm">
        <CircleCheck v-if="step.done" class="size-5 text-green-600" />
        <Circle v-else class="size-5 text-muted-foreground" />
        <span :class="step.done ? '' : 'text-muted-foreground'">{{ $t(`onboarding.steps.${step.key}.title`) }}</span>
        <span v-if="!step.done && step.required" class="text-xs text-amber-600">{{ $t('onboarding.finish.recommended') }}</span>
      </li>
    </ul>
    <p class="text-sm text-muted-foreground">{{ $t('onboarding.finish.reopen') }}</p>
    <Button @click="$emit('finish')">{{ $t('onboarding.finish.goToInbox') }}</Button>
  </div>
</template>

<script setup>
import { Circle, CircleCheck } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'

defineProps({ steps: { type: Array, default: () => [] } })
defineEmits(['finish', 'saved'])
</script>
