import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'

import imagePopupSrc from './image-popup.svg'

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

const MapClusterMarker = ({
  thumbnail: thumbJson,
  point_count_abbreviated,
  cluster,
  cluster_id,
  media_id,
  setPresentMarker,
}) => {
  const thumbnail = JSON.parse(thumbJson)

  const presentMedia = () => {
    setPresentMarker({
      cluster: !!cluster,
      id: cluster ? cluster_id : media_id,
    })
  }

  return (
    <Wrapper onClick={presentMedia}>
      <PopupImage src={imagePopupSrc} />
      <ThumbnailImage src={thumbnail.url} />
      {cluster && (
        <PointCountCircle>{point_count_abbreviated}</PointCountCircle>
      )}
    </Wrapper>
  )
}

MapClusterMarker.propTypes = {
  thumbnail: PropTypes.string,
  cluster: PropTypes.bool,
  point_count_abbreviated: PropTypes.number,
  cluster_id: PropTypes.number,
  media_id: PropTypes.number,
  setPresentMarker: PropTypes.func,
}

export default MapClusterMarker
