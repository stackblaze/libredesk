// Libredesk chat widget service worker.
//
// Makes the standalone widget page (/widget) installable as a PWA and provides
// lightweight runtime caching for the widget's hashed build assets and shared
// static images. API (/api) and websocket traffic always hit the network.

const CACHE = 'libredesk-widget-static-v1'

// /widget/assets/ holds content-hashed widget build output; /images/ holds the
// shared static icons used by the manifest.
const CACHEABLE_PREFIXES = ['/widget/assets/', '/images/']

// Do NOT skipWaiting on install. The page shows an update toast and the user
// opts in via SKIP_WAITING (same flow as the main app).
self.addEventListener('install', () => {})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      const keys = await caches.keys()
      await Promise.all(keys.filter((key) => key !== CACHE).map((key) => caches.delete(key)))
    })()
  )
})

self.addEventListener('message', (event) => {
  if (event.data === 'SKIP_WAITING') {
    self.skipWaiting()
  }
})

self.addEventListener('fetch', (event) => {
  const { request } = event

  if (request.method !== 'GET') return

  let url
  try {
    url = new URL(request.url)
  } catch {
    return
  }

  if (url.origin !== self.location.origin) return

  const isCacheable = CACHEABLE_PREFIXES.some((prefix) => url.pathname.startsWith(prefix))
  if (!isCacheable) return

  event.respondWith(
    (async () => {
      const cache = await caches.open(CACHE)
      const cached = await cache.match(request)

      const fromNetwork = fetch(request)
        .then((response) => {
          if (response && response.ok) {
            cache.put(request, response.clone())
          }
          return response
        })
        .catch(() => cached)

      return cached || fromNetwork
    })()
  )
})
