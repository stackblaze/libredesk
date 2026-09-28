<template>
  <div class="space-y-4">
    <div class="flex items-center gap-3">
      <div class="ml-auto flex shrink-0 items-center gap-2">
        <Button @click="newRule">
          <Plus class="size-4" />
          {{ $t('automation.newRule') }}
        </Button>
      </div>
    </div>
    <AutomationTabs v-if="selectedTab" v-model:automationsTab="selectedTab" />
  </div>
</template>

<script setup>
import { Plus } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { useRouter } from 'vue-router'
import { useStorage } from '@vueuse/core'
import AutomationTabs from '@/features/admin/automation/AutomationTabs.vue'

const router = useRouter()
const selectedTab = useStorage('automationsTab', 'new_conversation')
const newRule = () => {
  router.push({ name: 'new-automation', query: { type: selectedTab.value } })
}
</script>
