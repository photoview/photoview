import React from 'react'
import styled from 'styled-components'
import imagePopupSrc from './image-popup.svg'
import { MediaMarker } from './MapPresentMarker'
import { PlacesAction } from './placesReducer'

const Wrapper = styled.div`
  width: 56px;
  height: 68px;
  position: relative;
  margin-top: -54px;
  cursor: pointer;
`

const ThumbnailImage = styled.img`
  position: absolute;
  width: 48px;
  height: 48px;
  top: 4px;
  left: 4px;
  border-radius: 2px;
  object-fit: cover;
`

const PopupImage = styled.img`
  width: 100%;
  height: 100%;
`

const PointCountCircle = styled.div`
  position: absolute;
  top: -10px;
  right: -10px;
  width: 24px;
  height: 24px;
  background-color: #00b3dc;
  border-radius: 50%;
  color: white;
  text-align: center;
  padding-top: 2px;
`

type MapClusterMarkerProps = {
  dispatchMarkerMedia: React.Dispatch<PlacesAction>
  marker: MediaMarker
}

const MapClusterMarker = ({
  marker,
  dispatchMarkerMedia,
}: MapClusterMarkerProps) => {
  const thumbnail = JSON.parse(marker.thumbnail) as { url: string }

  const presentMedia = () => {
    dispatchMarkerMedia({
      type: 'replacePresentMarker',
      marker: {
        cluster: !!marker.cluster,
        id: marker.cluster ? marker.cluster_id : marker.media_id,
      },
    })
  }

  return (
    <Wrapper onClick={presentMedia}>
      <PopupImage src={imagePopupSrc} />
      <ThumbnailImage src={thumbnail.url} />
      {marker.cluster && (
        <PointCountCircle>{marker.point_count_abbreviated}</PointCountCircle>
      )}
    </Wrapper>
  )
}

export default MapClusterMarker
