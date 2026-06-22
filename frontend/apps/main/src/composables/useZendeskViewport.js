import { watch, onMounted, onUnmounted } from 'vue'
import { useUiLayout, UI_LAYOUT_ZENDESK } from './useUiLayout'

const DEFAULT_VIEWPORT = 'width=1280, initial-scale=0.29, shrink-to-fit=no'
const ZENDESK_VIEWPORT = 'width=device-width, initial-scale=1'

function setViewport (content) {
  let meta = document.querySelector('meta[name="viewport"]')
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = 'viewport'
    document.head.appendChild(meta)
  }
  meta.content = content
}

/** Use full device width in Zendesk layout; restore scaled viewport in default layout. */
export function useZendeskViewport () {
  const { layout } = useUiLayout()

  const sync = (value) => {
    document.documentElement.classList.toggle('zendesk-layout', value === UI_LAYOUT_ZENDESK)
    setViewport(value === UI_LAYOUT_ZENDESK ? ZENDESK_VIEWPORT : DEFAULT_VIEWPORT)
  }

  onMounted(() => sync(layout.value))
  onUnmounted(() => {
    document.documentElement.classList.remove('zendesk-layout')
    setViewport(DEFAULT_VIEWPORT)
  })

  watch(layout, sync)
}
