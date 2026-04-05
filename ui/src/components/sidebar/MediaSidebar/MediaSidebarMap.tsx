import React from 'react'
import { useTranslation } from 'react-i18next'
import useMaplibreMap from '../../maplibre/MaplibreMap'
import useMapStyles from '../../maplibre/useMapStyles'
import { SidebarSection, SidebarSectionTitle } from '../SidebarComponents'
import { sidebarMediaQuery_media_exif_coordinates } from './__generated__/sidebarMediaQuery'

type MediaSidebarMapProps = {
  coordinates: sidebarMediaQuery_media_exif_coordinates
}

const MediaSidebarMap = ({ coordinates }: MediaSidebarMapProps) => {
  const { t, i18n } = useTranslation()
  const { mapStyleLight, mapStyleDark } = useMapStyles()

  const { mapContainer } = useMaplibreMap({
    locale: i18n.language,
    mapStyleLight,
    mapStyleDark,
    mapOptions: {
      interactive: false,
      zoom: 12,
      center: {
        lat: coordinates.latitude,
        lng: coordinates.longitude,
      },
    },
    configureMap: (map, maplibreLibrary) => {
      const centerMarker = new maplibreLibrary.Marker({
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
