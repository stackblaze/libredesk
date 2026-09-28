<script setup>
import { cn } from '../../../lib/utils';

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <ul
    data-sidebar="menu-badge"
    :class="
      cn(
        'ml-7 flex min-w-0 flex-col gap-px border-l border-border py-0.5 pl-2',
        'group-data-[collapsible=icon]:hidden',
        props.class,
      )
    "
  >
    <slot />
  </ul>
</template>
