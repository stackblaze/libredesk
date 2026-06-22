import { useStorage } from '@vueuse/core'
import { applyUiLayoutViewport } from './zendeskViewport'

export const UI_LAYOUT_DEFAULT = 'default'
export const UI_LAYOUT_ZENDESK = 'zendesk'

/** Per-browser UI layout preference (default | zendesk). */
export function useUiLayout () {
  const layout = useStorage('libredesk_ui_layout', UI_LAYOUT_DEFAULT)

  const isZendesk = () => layout.value === UI_LAYOUT_ZENDESK

  const setLayout = (value) => {
    const next = value === UI_LAYOUT_ZENDESK ? UI_LAYOUT_ZENDESK : UI_LAYOUT_DEFAULT
    if (layout.value === next) return
    layout.value = next
    applyUiLayoutViewport(next)
    // Viewport width is fixed at first paint; reload so Zendesk fills the screen.
    window.location.reload()
  }

  return { layout, isZendesk, setLayout }
}
