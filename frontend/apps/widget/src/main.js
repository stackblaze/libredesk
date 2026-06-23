import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { toast } from 'vue-sonner'
import App from './App.vue'
import api from './api/index.js'
import '@shared-ui/assets/styles/main.scss'

async function initWidget () {
    try {
        // Get `inbox_id` from URL params
        const urlParams = new URLSearchParams(window.location.search)
        const inboxID = urlParams.get('inbox_id')

        if (!inboxID) {
            throw new Error('`inbox_id` is missing in query parameters')
        }

        // Fetch widget settings to get language config
        const widgetSettingsResponse = await api.getWidgetSettings(inboxID)
        const widgetConfig = widgetSettingsResponse.data.data

        // Resolve language: auto-detect from browser or use admin-configured language.
        let lang
        const fallbackLang = widgetConfig.fallback_language || 'en-US'
        if (widgetConfig.language === 'auto') {
            const browserLang = navigator.language || navigator.languages?.[0] || ''
            const availableResp = await api.getAvailableLanguages()
            const availableCodes = availableResp.data.data.map(l => l.code)
            lang = availableCodes.includes(browserLang) ? browserLang : fallbackLang
        } else {
            lang = widgetConfig.language || fallbackLang
        }

        // Fetch language messages
        const langMessages = await api.getLanguage(lang)

        // Initialize i18n
        const i18nConfig = {
            legacy: false,
            locale: lang,
            fallbackLocale: fallbackLang,
            messages: {
                [lang]: langMessages.data
            }
        }

        const i18n = createI18n(i18nConfig)
        const app = createApp(App)
        const pinia = createPinia()

        app.use(pinia)
        app.use(i18n)
        // Store widget config globally for access in App.vue
        app.config.globalProperties.$widgetConfig = widgetConfig
        app.mount('#app')
    } catch (error) {
        console.error('Error initializing widget:', error)
    }
}

initWidget()

// Register the widget service worker (PWA install + static asset caching).
// Scoped to /widget so it never interferes with the main app. Production only.
if (import.meta.env.PROD && 'serviceWorker' in navigator) {
  window.addEventListener('load', async () => {
    try {
      const registration = await navigator.serviceWorker.register('/widget/sw.js', { scope: '/widget' })

      const promptUpdate = (worker) => {
        if (!worker) return
        toast('A new version of Libredesk Chat is available.', {
          duration: Infinity,
          action: {
            label: 'Reload',
            onClick: () => worker.postMessage('SKIP_WAITING')
          }
        })
      }

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

      let refreshing = false
      navigator.serviceWorker.addEventListener('controllerchange', () => {
        if (refreshing) return
        refreshing = true
        window.location.reload()
      })
    } catch (error) {
      console.error('Widget service worker registration failed: ', error)
    }
  })
}
