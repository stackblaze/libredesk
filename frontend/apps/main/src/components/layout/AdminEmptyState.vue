<template>
  <div
    class="flex min-h-[240px] flex-col items-center justify-center rounded-lg border border-dashed px-6 py-12 text-center"
  >
    <component
      :is="icon"
      class="size-10 text-muted-foreground"
      :stroke-width="1.5"
    />
    <p class="mt-3 text-sm font-medium text-foreground">{{ title }}</p>
    <p v-if="description" class="mt-1 max-w-sm text-sm text-muted-foreground">
      {{ description }}
    </p>
    <div v-if="$slots.action" class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup>
import { Inbox } from 'lucide-vue-next'

defineProps({
  icon: { type: [Object, Function], default: () => Inbox },
  title: { type: String, required: true },
  description: { type: String, default: '' }
})
</script>
