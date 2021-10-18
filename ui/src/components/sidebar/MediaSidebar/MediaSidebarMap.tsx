import React from 'react'
import { useTranslation } from 'react-i18next'
import { isNil } from '../../../helpers/utils'
import useMapboxMap from '../../mapbox/MapboxMap'
import { SidebarSection, SidebarSectionTitle } from '../SidebarComponents'
import { sidebarMediaQuery_media_exif_coordinates } from './__generated__/sidebarMediaQuery'

type MediaSidebarMapProps = {
  coordinates: sidebarMediaQuery_media_exif_coordinates
}

const MediaSidebarMap = ({ coordinates }: MediaSidebarMapProps) => {
  const { t } = useTranslation()

  const { mapContainer, mapboxToken } = useMapboxMap({
    mapboxOptions: {
      interactive: false,
      zoom: 12,
      center: {
        lat: coordinates.latitude,
        lng: coordinates.longitude,
      },
    },
    configureMapbox: (map, mapboxLibrary) => {
      // todo
      map.addControl(
        new mapboxLibrary.NavigationControl({ showCompass: false })
      )

      const centerMarker = new mapboxLibrary.Marker({
        color: 'red',
        scale: 0.8,
      })
      centerMarker.setLngLat({
        lat: coordinates.latitude,
        lng: coordinates.longitude,
      })
      centerMarker.addTo(map)
    },
  })

  if (isNil(mapboxToken)) {
    return null
  }

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.location.title', 'Location')}
      </SidebarSectionTitle>
      <div className="w-full h-64">{mapContainer}</div>
    </SidebarSection>
  )
}

export default MediaSidebarMap
