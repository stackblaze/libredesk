<template>
  <div class="space-y-6 max-w-xl">
    <div v-if="inboxes.length" class="rounded-lg border">
      <div v-for="inbox in inboxes" :key="inbox.id" class="flex items-center gap-3 border-b px-4 py-3 last:border-b-0">
        <Mail v-if="inbox.channel === 'email'" class="size-4 text-muted-foreground" />
        <MessageCircle v-else class="size-4 text-muted-foreground" />
        <span class="flex-1 text-sm font-medium">{{ inbox.name }}</span>
        <span class="text-xs text-muted-foreground">{{ inbox.channel === 'email' ? 'Email' : 'Live chat' }}</span>
      </div>
    </div>

    <div class="grid gap-4 sm:grid-cols-2">
      <button
        v-for="option in options"
        :key="option.key"
        type="button"
        class="rounded-lg border p-4 text-left transition-colors hover:bg-muted/60"
        @click="createInbox"
      >
        <component :is="option.icon" class="mb-3 size-5" />
        <p class="font-medium">{{ $t(`onboarding.inbox.${option.key}`) }}</p>
        <p class="mt-1 text-sm text-muted-foreground">{{ $t(`onboarding.inbox.${option.key}Description`) }}</p>
      </button>
    </div>
    <p class="text-xs text-muted-foreground">{{ $t('onboarding.inbox.returnNote') }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Mail, MessageCircle } from 'lucide-vue-next'
import api from '@main/api'

defineEmits(['saved'])
const router = useRouter()
const inboxes = ref([])
const options = [
  { key: 'email', icon: Mail },
  { key: 'livechat', icon: MessageCircle }
]

onMounted(async () => {
  try {
    const resp = await api.getInboxes()
    inboxes.value = resp.data.data || []
  } catch {
    inboxes.value = []
  }
})

// The existing inbox wizard handles channel setup and brings the admin back here afterwards.
const createInbox = () => router.push({ name: 'new-inbox', query: { from: 'onboarding' } })
</script>
