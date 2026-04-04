import React, { useRef, useEffect } from 'react'
import type maplibregl from 'maplibre-gl'
import styled from 'styled-components'

import 'maplibre-gl/dist/maplibre-gl.css'
import rtlTextPluginUrl from '@mapbox/mapbox-gl-rtl-text/mapbox-gl-rtl-text.js?url'
import { isDarkMode } from '../../theme'

import geolocateIcon from './icons/geolocate.svg'
import geolocateBgIcon from './icons/geolocate-background.svg'
import globeEnabledIcon from './icons/globe-enabled.svg'

const BRAND_ACCENT = '#ff5d43'

const recolorIcon = (iconUrl: string) => `
  background-image: none !important;
  background-color: ${BRAND_ACCENT};
  mask-image: url(${iconUrl});
  mask-size: contain;
  mask-repeat: no-repeat;
  mask-position: center;
`

const MapContainer = styled.div`
  width: 100%;
  height: 100%;
  border-radius: 12px;
  overflow: hidden;

  /* Override MapLibre accent color for geolocate active states */
  .maplibregl-ctrl button.maplibregl-ctrl-geolocate.maplibregl-ctrl-geolocate-active .maplibregl-ctrl-icon {
    ${recolorIcon(geolocateIcon)}
  }
  .maplibregl-ctrl button.maplibregl-ctrl-geolocate.maplibregl-ctrl-geolocate-background .maplibregl-ctrl-icon {
    ${recolorIcon(geolocateBgIcon)}
  }

  /* Override MapLibre accent color for globe enabled state */
  .maplibregl-ctrl button.maplibregl-ctrl-globe-enabled .maplibregl-ctrl-icon {
    ${recolorIcon(globeEnabledIcon)}
  }

  /* Geolocate accuracy circle */
  .maplibregl-user-location-accuracy-circle {
    background-color: rgba(255, 93, 67, 0.2) !important;
  }
  .maplibregl-user-location-dot,
  .maplibregl-user-location-dot::before {
    background-color: #ff5d43 !important;
  }
`

/**
 * Localize map label layers by setting their text-field to prefer
 * the `name:<locale>` property, falling back to `name`.
 * OpenFreemap styles default to name_en; this overrides that.
 */
function localizeLabels(map: maplibregl.Map, locale: string) {
  const style = map.getStyle()
  if (!style?.layers) return

  const nameExpr: maplibregl.ExpressionSpecification = [
    'coalesce',
    ['get', `name:${locale}`],
    ['get', 'name'],
  ]

  for (const layer of style.layers) {
    if (layer.type !== 'symbol') continue
    const tf = (layer.layout as any)?.['text-field']
    if (!tf) continue
    map.setLayoutProperty(layer.id, 'text-field', nameExpr)
  }
}

function getStyleUrl() {
  return isDarkMode()
    ? 'https://tiles.openfreemap.org/styles/dark'
    : 'https://tiles.openfreemap.org/styles/positron'
}

type MaplibreMapProps = {
  configureMap(map: maplibregl.Map, maplibreLibrary: typeof maplibregl): void
  mapOptions?: Partial<maplibregl.MapOptions>
  onStyleLoad?: (map: maplibregl.Map, maplibreLibrary: typeof maplibregl) => void
  locale?: string
}

const useMaplibreMap = ({
  configureMap,
  mapOptions,
  onStyleLoad,
  locale,
}: MaplibreMapProps) => {
  const mapContainer = useRef<HTMLDivElement | null>(null)
  const map = useRef<maplibregl.Map | null>(null)
  const maplibreRef = useRef<typeof maplibregl | null>(null)
  const onStyleLoadRef = useRef(onStyleLoad)
  onStyleLoadRef.current = onStyleLoad
  const localeRef = useRef(locale)
  localeRef.current = locale

  const configureMapRef = useRef(configureMap)
  configureMapRef.current = configureMap

  const mapOptionsRef = useRef(mapOptions)
  mapOptionsRef.current = mapOptions

  const [, setReady] = React.useState(false)

  useEffect(() => {
    let cancelled = false

    async function init() {
      if (mapContainer.current == null || map.current != null) return

      const maplibre = (await import('maplibre-gl')).default
      if (cancelled) return

      maplibreRef.current = maplibre

      if (maplibre.getRTLTextPluginStatus() === 'unavailable') {
        maplibre.setRTLTextPlugin(rtlTextPluginUrl, true)
      }

      const m = new maplibre.Map({
        container: mapContainer.current,
        style: getStyleUrl(),
        zoom: 2,
        ...mapOptionsRef.current,
      })

      map.current = m

      m.on('style.load', () => {
        m.setProjection({ type: 'globe' })
        if (localeRef.current) {
          localizeLabels(m, localeRef.current)
        }
        onStyleLoadRef.current?.(m, maplibre)
      })

      configureMapRef.current(m, maplibre)
      setReady(true)

      // Watch for dark/light mode changes
      const observer = new MutationObserver(() => {
        if (map.current) {
          map.current.setStyle(getStyleUrl())
        }
      })

      observer.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class'],
      })

      // Store observer for cleanup
      ;(m as any)._themeObserver = observer
    }

    init()

    return () => {
      cancelled = true
      if (map.current) {
        const observer = (map.current as any)._themeObserver as
          | MutationObserver
          | undefined
        observer?.disconnect()
        map.current.remove()
        map.current = null
      }
    }
  }, [])

  // Re-localize when locale changes while the map is mounted
  useEffect(() => {
    if (map.current && locale) {
      localizeLabels(map.current, locale)
    }
  }, [locale])

  return {
    mapContainer: <MapContainer ref={mapContainer}></MapContainer>,
    maplibreMap: map.current,
    maplibreLibrary: maplibreRef.current ?? undefined,
  }
}

export default useMaplibreMap
