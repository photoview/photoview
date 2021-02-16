import PropTypes from 'prop-types'
import React from 'react'
import styled from 'styled-components'

const FaceBoxStyle = styled.div`
  box-shadow: inset 0 0 2px 1px rgba(0, 0, 0, 0.3), 0 0 0 1px rgb(255, 255, 255);
  border-radius: 50%;
  position: absolute;
  top: ${({ minY }) => minY * 100}%;
  bottom: ${({ maxY }) => (1 - maxY) * 100}%;
  left: ${({ minX }) => minX * 100}%;
  right: ${({ maxX }) => (1 - maxX) * 100}%;
`

const FaceBox = ({ face /*media*/ }) => {
  return <FaceBoxStyle {...face.rectangle}></FaceBoxStyle>
}

FaceBox.propTypes = {
  face: PropTypes.object.isRequired,
  media: PropTypes.object.isRequired,
}

const SidebarFacesOverlayWrapper = styled.div`
  position: absolute;
  width: ${({ width }) => width * 100}%;
  left: ${({ width }) => (100 - width * 100) / 2}%;
  height: 100%;
  top: 0;
  opacity: 0;

  user-select: none;
  transition: opacity ease 200ms;

  &:hover {
    opacity: 1;
  }
`

export const SidebarFacesOverlay = ({ media }) => {
  if (media.type != 'photo') return null

  const faceBoxes = media.faces?.map(face => (
    <FaceBox key={face.id} face={face} media={media} />
  ))

  let wrapperWidth = 1
  if (media.thumbnail.width * 0.75 < media.thumbnail.height) {
    wrapperWidth = (media.thumbnail.width * 0.75) / media.thumbnail.height
  }

  return (
    <SidebarFacesOverlayWrapper width={wrapperWidth}>
      {faceBoxes}
    </SidebarFacesOverlayWrapper>
  )
}

SidebarFacesOverlay.propTypes = {
  media: PropTypes.object.isRequired,
}
