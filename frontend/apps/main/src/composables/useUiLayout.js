import { useStorage } from '@vueuse/core'

export const UI_LAYOUT_DEFAULT = 'default'
export const UI_LAYOUT_ZENDESK = 'zendesk'

/** Per-browser UI layout preference (default | zendesk). */
export function useUiLayout () {
  const layout = useStorage('libredesk_ui_layout', UI_LAYOUT_DEFAULT)

  const isZendesk = () => layout.value === UI_LAYOUT_ZENDESK

  const setLayout = (value) => {
    layout.value = value === UI_LAYOUT_ZENDESK ? UI_LAYOUT_ZENDESK : UI_LAYOUT_DEFAULT
  }

  return { layout, isZendesk, setLayout }
}
