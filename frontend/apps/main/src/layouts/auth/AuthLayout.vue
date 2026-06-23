<template>
  <div
    class="auth-layout min-h-dvh flex flex-col bg-secondary px-4 sm:px-6 pt-[max(1rem,env(safe-area-inset-top))] pb-[max(1rem,env(safe-area-inset-bottom))]"
  >
    <main class="flex flex-1 items-center justify-center py-6 sm:py-10">
      <div class="w-full max-w-[22rem] sm:max-w-sm">
        <slot />
      </div>
    </main>
    <footer class="shrink-0 pb-2 text-center">
      <a
        href="https://libredesk.io/"
        class="text-xs text-muted-foreground hover:text-foreground transition-colors"
        target="_blank"
        rel="noopener noreferrer"
      >
        Powered by libredesk
      </a>
    </footer>
  </div>
</template>

<script setup></script>

<style scoped>
/* iOS Safari auto-zooms when font-size < 16px on focused inputs. */
.auth-layout :deep(input),
.auth-layout :deep(select),
.auth-layout :deep(textarea) {
  font-size: 16px;
  min-height: 2.75rem;
}

.auth-layout :deep(button[type='submit']) {
  min-height: 2.75rem;
}
</style>
