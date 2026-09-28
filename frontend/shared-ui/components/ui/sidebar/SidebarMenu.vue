<script setup>
import { cn } from '../../../lib/utils';

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <ul
    data-sidebar="menu"
    :class="cn('flex w-full min-w-0 flex-col gap-px', props.class)"
  >
    <slot />
  </ul>
</template>
