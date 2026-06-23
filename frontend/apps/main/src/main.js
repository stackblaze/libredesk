import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { toast } from 'vue-sonner'
import { initI18n } from './i18n'
import { useAppSettingsStore } from './stores/appSettings'
import router from './router'
import mitt from 'mitt'
import api from './api'
import { applyStoredUiLayoutViewport } from './composables/zendeskViewport'
import '@shared-ui/assets/styles/main.scss'
import '@main/layouts/zendesk/zendesk.scss'
import '@shared-ui/utils/string.js'
import Root from './Root.vue'

applyStoredUiLayoutViewport()

const setFavicon = (url) => {
  let link = document.createElement("link")
  link.rel = "icon"
  document.head.appendChild(link)
  link.href = url
}

async function initApp () {
  const config = (await api.getConfig()).data.data
  const emitter = mitt()
  const lang = config['app.lang'] || 'en-US'
  const langMessages = await api.getLanguage(lang)

  // Set favicon.
  if (config['app.favicon_url'])
    setFavicon(config['app.favicon_url'])

  // Initialize i18n.
  const i18nConfig = {
    legacy: false,
    locale: lang,
    fallbackLocale: 'en-US',
    messages: {
      [lang]: langMessages.data
    }
  }

  const i18n = initI18n(i18nConfig)
  const app = createApp(Root)
  const pinia = createPinia()
  app.use(pinia)

  // Fetch and store app settings in store (after pinia is initialized)
  const settingsStore = useAppSettingsStore()

  // Store the public config in the store
  settingsStore.setPublicConfig(config)

  try {
    await settingsStore.fetchSettings('general')
  } catch (error) {
    // Pass
  }

  // Add emitter to global properties.
  app.config.globalProperties.emitter = emitter

  app.use(router)
  app.use(i18n)
  app.mount('#app')
}

initApp().catch((error) => {
  console.error('Error initializing app: ', error)
})

// Register the service worker to enable PWA install + runtime caching of static
// assets. Only in production builds to avoid caching surprises during dev.
if (import.meta.env.PROD && 'serviceWorker' in navigator) {
  window.addEventListener('load', async () => {
    try {
      const registration = await navigator.serviceWorker.register('/sw.js', { scope: '/' })

      const promptUpdate = (worker) => {
        if (!worker) return
        toast('A new version of Libredesk is available.', {
          duration: Infinity,
          action: {
            label: 'Reload',
            onClick: () => worker.postMessage('SKIP_WAITING')
          }
        })
      }

      // A worker updated in a previous session may already be waiting.
      if (registration.waiting && navigator.serviceWorker.controller) {
        promptUpdate(registration.waiting)
      }

      registration.addEventListener('updatefound', () => {
        const newWorker = registration.installing
        if (!newWorker) return
        newWorker.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            promptUpdate(newWorker)
          }
        })
      })

      // Reload once the new worker takes control (after the user accepts).
      let refreshing = false
      navigator.serviceWorker.addEventListener('controllerchange', () => {
        if (refreshing) return
        refreshing = true
        window.location.reload()
      })
    } catch (error) {
      console.error('Service worker registration failed: ', error)
    }
  })
}
