<template>
  <div class="flex flex-1 min-h-0 gap-8">
    <div class="min-w-0 flex-1 relative">
      <slot name="content" />
    </div>
    <aside
      v-if="$slots.help"
      class="hidden lg:block w-[17rem] shrink-0 space-y-4 text-sm text-muted-foreground"
    >
      <slot name="help" />
    </aside>
  </div>
</template>
