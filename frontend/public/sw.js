// Libredesk service worker.
//
// Goals:
//  - Make Libredesk installable as a standalone PWA (desktop + mobile).
//  - Provide lightweight runtime caching for content-hashed build assets and
//    static images so the app shell loads fast and survives flaky networks.
//
// Deliberately conservative: navigations, the API (/api), websockets (/ws) and
// everything else always hit the network so that auth-gated, dynamic content is
// never served stale.

const CACHE = 'libredesk-static-v1'

// Path prefixes that are safe to cache. /assets/ holds content-hashed build
// output (filenames change on every deploy) and /images/ holds static images.
const CACHEABLE_PREFIXES = ['/assets/', '/images/']

// Do NOT skipWaiting on install. A new worker stays in "waiting" until the app
// shows an "update available" toast and the user opts in (which posts
// SKIP_WAITING below). This avoids surprise reloads / version mismatches.
self.addEventListener('install', () => {})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      const keys = await caches.keys()
      await Promise.all(keys.filter((key) => key !== CACHE).map((key) => caches.delete(key)))
    })()
  )
})

// The page posts SKIP_WAITING when the user accepts an update; activating the
// waiting worker triggers a controllerchange in the page, which reloads.
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

  // Stale-while-revalidate: serve from cache immediately, refresh in background.
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
