import { useStorage } from '@vueuse/core'
import { applyUiLayoutViewport } from './zendeskViewport'

export const UI_LAYOUT_DEFAULT = 'default'
export const UI_LAYOUT_ZENDESK = 'zendesk'

/** UI layout is locked to Zendesk (OpenBooks-styled agent workspace). */
export function useUiLayout () {
  const layout = useStorage('libredesk_ui_layout', UI_LAYOUT_ZENDESK)

  // Migrate any leftover "default" preference from before the lock.
  if (layout.value !== UI_LAYOUT_ZENDESK) {
    layout.value = UI_LAYOUT_ZENDESK
  }

  const isZendesk = () => true

  const setLayout = () => {
    if (layout.value !== UI_LAYOUT_ZENDESK) {
      layout.value = UI_LAYOUT_ZENDESK
      applyUiLayoutViewport(UI_LAYOUT_ZENDESK)
      window.location.reload()
    }
  }

  return { layout, isZendesk, setLayout }
}
