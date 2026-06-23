import { watch, onMounted, onUnmounted } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useUiLayout, UI_LAYOUT_ZENDESK } from './useUiLayout'
import { applyUiLayoutViewport, DEFAULT_VIEWPORT, MOBILE_BREAKPOINT, setViewportContent } from './zendeskViewport'

/** Keep viewport in sync when layout preference or screen size changes. */
export function useZendeskViewport () {
  const { layout } = useUiLayout()
  const isSmallScreen = useMediaQuery(MOBILE_BREAKPOINT)

  const sync = () => {
    applyUiLayoutViewport(layout.value)
  }

  onMounted(() => sync())
  onUnmounted(() => {
    document.documentElement.classList.remove('zendesk-layout', 'mobile-layout')
    setViewportContent(DEFAULT_VIEWPORT)
  })

  watch([layout, isSmallScreen], sync)
}

export { UI_LAYOUT_ZENDESK }
