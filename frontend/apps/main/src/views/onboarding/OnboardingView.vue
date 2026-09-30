<template>
  <div class="min-h-screen bg-background text-foreground">
    <header class="flex items-center justify-between border-b px-4 py-3 sm:px-8">
      <div class="flex items-center gap-3">
        <img v-if="logoUrl && !logoFailed" :src="logoUrl" alt="" class="h-7 w-auto" @error="logoFailed = true" />
        <span class="font-semibold">{{ $t('onboarding.title') }}</span>
      </div>
      <Button variant="ghost" size="sm" :disabled="isSaving" @click="dismiss">
        {{ $t('onboarding.skipForNow') }}
      </Button>
    </header>

    <div v-if="isLoading" class="flex justify-center py-24"><Spinner /></div>

    <div v-else class="mx-auto flex max-w-5xl flex-col gap-8 px-4 py-8 md:flex-row sm:px-8">
      <nav class="md:w-64 shrink-0">
        <p class="mb-4 text-sm text-muted-foreground">
          {{ $t('onboarding.progress', { done: doneCount, total: steps.length }) }}
        </p>
        <ol class="space-y-1">
          <li v-for="(step, index) in allSteps" :key="step.key">
            <button
              type="button"
              class="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors"
              :class="index === current ? 'bg-muted font-medium' : 'hover:bg-muted/60'"
              @click="current = index"
            >
              <span
                class="flex size-6 shrink-0 items-center justify-center rounded-full border text-xs"
                :class="isDone(step.key) ? 'border-primary bg-primary text-primary-foreground' : ''"
              >
                <Check v-if="isDone(step.key)" class="size-3.5" />
                <template v-else>{{ index + 1 }}</template>
              </span>
              <span class="flex-1">{{ $t(`onboarding.steps.${step.key}.title`) }}</span>
              <span v-if="step.optional" class="text-xs text-muted-foreground">{{ $t('onboarding.optional') }}</span>
            </button>
          </li>
        </ol>
      </nav>

      <main class="flex-1 min-w-0">
        <h1 class="text-2xl font-semibold">{{ $t(`onboarding.steps.${currentStep.key}.title`) }}</h1>
        <p class="mt-1 mb-6 text-sm text-muted-foreground">{{ $t(`onboarding.steps.${currentStep.key}.description`) }}</p>

        <component
          :is="currentStep.component"
          :key="currentStep.key"
          v-bind="currentStep.key === 'finish' ? { steps } : {}"
          @saved="onSaved"
          @finish="finish"
        />

        <div v-if="currentStep.key !== 'finish'" class="mt-8 flex items-center justify-between border-t pt-4">
          <Button variant="ghost" :disabled="current === 0" @click="current--">{{ $t('onboarding.back') }}</Button>
          <Button variant="outline" @click="current++">
            {{ isDone(currentStep.key) ? $t('onboarding.next') : $t('onboarding.skipStep') }}
          </Button>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, markRaw } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Check } from 'lucide-vue-next'
import { Button } from '@shared-ui/components/ui/button'
import { Spinner } from '@shared-ui/components/ui/spinner'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import api from '@main/api'
import { useUserStore } from '@main/stores/user'
import { useAppSettingsStore } from '@main/stores/appSettings'
import { useEmitter } from '@main/composables/useEmitter'
import { EMITTER_EVENTS } from '@main/constants/emitterEvents.js'
import WorkspaceStep from '@main/features/onboarding/WorkspaceStep.vue'
import EmailStep from '@main/features/onboarding/EmailStep.vue'
import InboxStep from '@main/features/onboarding/InboxStep.vue'
import TeammatesStep from '@main/features/onboarding/TeammatesStep.vue'
import TeamsStep from '@main/features/onboarding/TeamsStep.vue'
import SignInStep from '@main/features/onboarding/SignInStep.vue'
import FinishStep from '@main/features/onboarding/FinishStep.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const appSettingsStore = useAppSettingsStore()
const emitter = useEmitter()

const allSteps = [
  { key: 'workspace', component: markRaw(WorkspaceStep) },
  { key: 'email', component: markRaw(EmailStep) },
  { key: 'inbox', component: markRaw(InboxStep) },
  { key: 'teammates', component: markRaw(TeammatesStep), optional: true },
  { key: 'teams', component: markRaw(TeamsStep), optional: true },
  { key: 'sign_in', component: markRaw(SignInStep), optional: true },
  { key: 'finish', component: markRaw(FinishStep) }
]

const isLoading = ref(true)
const isSaving = ref(false)
const logoFailed = ref(false)
const steps = ref([])
const current = ref(0)

const currentStep = computed(() => allSteps[Math.min(current.value, allSteps.length - 1)])
const doneCount = computed(() => steps.value.filter((s) => s.done).length)
const logoUrl = computed(() => appSettingsStore.public_config?.['app.logo_url'] || '')
const isDone = (key) => steps.value.find((s) => s.key === key)?.done === true

const loadState = async () => {
  const resp = await api.getOnboarding()
  steps.value = resp.data.data.steps
}

onMounted(async () => {
  try {
    if (!userStore.userID) await userStore.getCurrentUser()
    if (!userStore.can('general_settings:manage')) {
      router.replace({ name: 'inboxes' })
      return
    }
    await loadState()
    const requested = allSteps.findIndex((s) => s.key === route.query.step)
    const firstOpen = allSteps.findIndex((s) => s.key !== 'finish' && !isDone(s.key))
    current.value = requested >= 0 ? requested : Math.max(firstOpen, 0)
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  } finally {
    isLoading.value = false
  }
})

// A step saved something: refresh live status and move on.
const onSaved = async ({ advance = true } = {}) => {
  try {
    await loadState()
  } catch {
    // Status refresh is best effort.
  }
  if (advance && current.value < allSteps.length - 1) current.value++
}

const setStatus = async (status) => {
  isSaving.value = true
  try {
    await api.updateOnboarding({ status })
    router.push({ name: 'inboxes' })
  } catch (error) {
    emitter.emit(EMITTER_EVENTS.SHOW_TOAST, { variant: 'destructive', description: handleHTTPError(error).message })
  } finally {
    isSaving.value = false
  }
}

const dismiss = () => setStatus('dismissed')
const finish = () => setStatus('completed')
</script>
