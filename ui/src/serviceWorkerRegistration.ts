// This optional code is used to register a service worker.
// register() is not called by default.

// This lets the app load faster on subsequent visits in production, and gives
// it offline capabilities. However, it also means that developers (and users)
// will only see deployed updates on subsequent visits to a page, after all the
// existing tabs open on the page have been closed, since previously cached
// resources are updated in the background.

// To learn more about the benefits of this model and instructions on how to
// opt-in, read https://cra.link/PWA

const isLocalhost = Boolean(
  window.location.hostname === 'localhost' ||
  // [::1] is the IPv6 localhost address.
  window.location.hostname === '[::1]' ||
  // 127.0.0.0/8 are considered localhost for IPv4.
  window.location.hostname.match(
    /^127(?:\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$/
  )
)

type Config = {
  onSuccess?: (registration: ServiceWorkerRegistration) => void
  onUpdate?: (registration: ServiceWorkerRegistration) => void
}

/**
 * Registers a service worker to enable offline capabilities and faster loading in production environments.
 *
 * If running on localhost, validates the service worker before registration and provides additional warnings for developers. In production, registers the service worker directly if supported and the public URL matches the current origin.
 *
 * @param config - Optional callbacks for service worker registration success or update events.
 */
export function register(config?: Config) {
  if (import.meta.env.PROD && 'serviceWorker' in navigator) {
    // The URL constructor is available in all browsers that support SW.
    const publicUrl = new URL(import.meta.env.BASE_URL, window.location.href)
    if (publicUrl.origin !== window.location.origin) {
      // Our service worker won't work if PUBLIC_URL is on a different origin
      // from what our page is served on. This might happen if a CDN is used to
      // serve assets; see https://github.com/facebook/create-react-app/issues/2374
      return
    }

    window.addEventListener('load', () => {
      const swUrl = `${import.meta.env.BASE_URL}service-worker.js`

      if (isLocalhost) {
        // This is running on localhost. Let's check if a service worker still exists or not.
        checkValidServiceWorker(swUrl, config)

        // Add some additional logging to localhost, pointing developers to the
        // service worker/PWA documentation.
        navigator.serviceWorker.ready
          .catch(() => {
            console.warn('[Service Worker] Failed to load service worker')
          })
      } else {
        // Is not localhost. Just register service worker
        registerValidSW(swUrl, config)
      }
    })
  }
}

/**
 * Registers a service worker at the specified URL and manages update and success callbacks.
 *
 * If a new service worker is found and installed, invokes the {@link Config.onUpdate} callback if a previous service worker controls the page, or {@link Config.onSuccess} if this is the initial installation.
 *
 * @param swUrl - The URL of the service worker script to register.
 * @param config - Optional configuration with callbacks for update and success events.
 */
function registerValidSW(swUrl: string, config?: Config) {
  navigator.serviceWorker
    .register(swUrl)
    .then(registration => {
      registration.onupdatefound = () => {
        const installingWorker = registration.installing
        if (installingWorker == null) {
          return
        }
        installingWorker.onstatechange = () => {
          if (installingWorker.state === 'installed') {
            if (navigator.serviceWorker.controller) {
              // At this point, the updated precached content has been fetched,
              // but the previous service worker will still serve the older
              // content until all client tabs are closed.
              // Execute callback
              if (config && config.onUpdate) {
                config.onUpdate(registration)
              }
              // At this point, everything has been precached.
              // Execute callback
            } else if (config?.onSuccess) {
              config.onSuccess(registration)
            }
          }
        }
      }
    })
    .catch(error => {
      console.error('Error during service worker registration:', error)
    })
}

/**
 * Validates the existence and type of the service worker script at the specified URL.
 *
 * If the script is missing or not JavaScript, unregisters any existing service worker and reloads the page. Otherwise, proceeds to register the service worker. Logs a warning if the network is unavailable.
 *
 * @param swUrl - The URL of the service worker script to validate.
 * @param config - Optional configuration with callbacks for registration events.
 */
function checkValidServiceWorker(swUrl: string, config?: Config) {
  // Check if the service worker can be found. If it can't reload the page.
  fetch(swUrl, {
    headers: { 'Service-Worker': 'script' },
  })
    .then(response => {
      // Ensure service worker exists, and that we really are getting a JS file.
      const contentType = response.headers.get('content-type')
      if (
        response.status === 404 ||
        (contentType != null && contentType.indexOf('javascript') === -1)
      ) {
        // No service worker found. Probably a different app. Reload the page.
        navigator.serviceWorker.ready.then(registration => {
          registration.unregister().then(() => {
            window.location.reload()
          })
        })
      } else {
        // Service worker found. Proceed as normal.
        registerValidSW(swUrl, config)
      }
    })
    .catch(() => {
      console.warn(
        '[Service Worker] No internet connection found. App is running in offline mode.'
      )
    })
}

export function unregister() {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.ready
      .then(registration => {
        registration.unregister()
      })
      .catch((error: Error) => {
        console.error(error.message)
      })
  }
}
