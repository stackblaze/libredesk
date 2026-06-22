export const UI_LAYOUT_STORAGE_KEY = 'libredesk_ui_layout'
export const UI_LAYOUT_ZENDESK = 'zendesk'

export const DEFAULT_VIEWPORT = 'width=1280, initial-scale=0.29, shrink-to-fit=no'
export const ZENDESK_VIEWPORT = 'width=device-width, initial-scale=1'

/**
 * Read layout preference from localStorage.
 * VueUse useStorage with a string default stores the value raw ("zendesk"),
 * but tolerate a JSON-quoted form ('"zendesk"') too.
 */
export function readStoredUiLayout () {
  try {
    const raw = localStorage.getItem(UI_LAYOUT_STORAGE_KEY)
    if (!raw) return null
    try {
      return JSON.parse(raw)
    } catch {
      return raw
    }
  } catch {
    return null
  }
}

export function setViewportContent (content) {
  let meta = document.querySelector('meta[name="viewport"]')
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = 'viewport'
    document.head.appendChild(meta)
  }
  meta.content = content
}

/** Apply viewport + html class for the given layout mode. Safe to call before Vue mounts. */
export function applyUiLayoutViewport (layout) {
  const isZendesk = layout === UI_LAYOUT_ZENDESK
  document.documentElement.classList.toggle('zendesk-layout', isZendesk)
  setViewportContent(isZendesk ? ZENDESK_VIEWPORT : DEFAULT_VIEWPORT)
}

/** Apply viewport from localStorage if Zendesk mode is already selected. */
export function applyStoredUiLayoutViewport () {
  applyUiLayoutViewport(readStoredUiLayout())
}
