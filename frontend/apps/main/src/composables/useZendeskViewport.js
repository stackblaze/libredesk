import { watch, onMounted, onUnmounted } from 'vue'
import { useUiLayout, UI_LAYOUT_ZENDESK } from './useUiLayout'
import { applyUiLayoutViewport, DEFAULT_VIEWPORT, setViewportContent } from './zendeskViewport'

/** Keep viewport in sync when layout preference changes at runtime. */
export function useZendeskViewport () {
  const { layout } = useUiLayout()

  const sync = (value) => {
    applyUiLayoutViewport(value)
  }

  onMounted(() => sync(layout.value))
  onUnmounted(() => {
    document.documentElement.classList.remove('zendesk-layout')
    setViewportContent(DEFAULT_VIEWPORT)
  })

  watch(layout, sync)
}

export { UI_LAYOUT_ZENDESK }
