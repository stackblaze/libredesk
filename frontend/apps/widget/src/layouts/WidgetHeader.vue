<template>
  <div class="flex items-center justify-center py-2.5 px-4 border-b">
    <h3 class="text-sm font-semibold text-foreground">{{ title }}</h3>
  </div>
</template>
<script setup>
defineProps(['title'])
</script>
