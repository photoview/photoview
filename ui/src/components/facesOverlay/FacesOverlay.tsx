import React from 'react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { MediaType } from '../../__generated__/globalTypes'
import { MediaSidebarMedia } from '../sidebar/MediaSidebar/MediaSidebar'
import { sidebarMediaQuery_media_faces } from '../sidebar/MediaSidebar/__generated__/sidebarMediaQuery'

interface FaceBoxStyleProps {
  $minY: number
  $maxY: number
  $minX: number
  $maxX: number
}

const FaceBoxStyle = styled(Link)`
  box-shadow: inset 0 0 2px 1px rgba(0, 0, 0, 0.3), 0 0 0 1px rgb(255, 255, 255);
  border-radius: 50%;
  position: absolute;
  top: ${({ $minY }: FaceBoxStyleProps) => $minY * 100}%;
  bottom: ${({ $maxY }: FaceBoxStyleProps) => (1 - $maxY) * 100}%;
  left: ${({ $minX }: FaceBoxStyleProps) => $minX * 100}%;
  right: ${({ $maxX }: FaceBoxStyleProps) => (1 - $maxX) * 100}%;
`

type FaceBoxProps = {
  face: sidebarMediaQuery_media_faces
}

const FaceBox = ({ face /*media*/ }: FaceBoxProps) => {
  return (
    <FaceBoxStyle
      to={`/people/${face.faceGroup.id}`}
      $minX={face.rectangle.minX}
      $maxX={face.rectangle.maxX}
      $minY={face.rectangle.minY}
      $maxY={face.rectangle.maxY}
    ></FaceBoxStyle>
  )
}

const SidebarFacesOverlayWrapper = styled.div<{ width: number }>`
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

type SidebarFaceOverlayProps = {
  media: MediaSidebarMedia
}

export const SidebarFacesOverlay = ({ media }: SidebarFaceOverlayProps) => {
  if (media.type != MediaType.Photo) return null
  if (media.thumbnail == null) return null

  const faceBoxes = media.faces?.map(face => (
    <FaceBox key={face.id} face={face} />
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
