import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { ProtectedImage } from '../../components/photoGallery/ProtectedMedia'

const FaceImage = styled(ProtectedImage)`
  position: absolute;
  transform-origin: ${({ $origin }) =>
    `${$origin.x * 100}% ${$origin.y * 100}%`};
  object-fit: cover;

  transition: transform 250ms ease-out;
`

const FaceImagePortrait = styled(FaceImage)`
  width: 100%;
  top: 50%;
  transform: translateY(-50%)
    ${({ $origin, $scale }) =>
      `translate(${(0.5 - $origin.x) * 100}%, ${
        (0.5 - $origin.y) * 100
      }%) scale(${Math.max($scale * 0.8, 1)})`};

  ${({ $selectable, $origin, $scale }) =>
    $selectable
      ? `
    &:hover {
      transform: translateY(-50%) translate(${(0.5 - $origin.x) * 100}%, ${
          (0.5 - $origin.y) * 100
        }%) scale(${Math.max($scale * 0.85, 1)})
      `
      : ''}
`

const FaceImageLandscape = styled(FaceImage)`
  height: 100%;
  left: 50%;
  transform: translateX(-50%)
    ${({ $origin, $scale }) =>
      `translate(${(0.5 - $origin.x) * 100}%, ${
        (0.5 - $origin.y) * 100
      }%) scale(${Math.max($scale * 0.8, 1)})`};

  ${({ $selectable, $origin, $scale }) =>
    $selectable
      ? `
    &:hover {
      transform: translateX(-50%) translate(${(0.5 - $origin.x) * 100}%, ${
          (0.5 - $origin.y) * 100
        }%) scale(${Math.max($scale * 0.85, 1)})
      `
      : ''}
`

const CircleImageWrapper = styled.div`
  background-color: #eee;
  position: relative;
  border-radius: 50%;
  width: ${({ size }) => size};
  height: ${({ size }) => size};
  object-fit: fill;
  overflow: hidden;
`

const FaceCircleImage = ({ imageFace, selectable, size = '150px' }) => {
  const rect = imageFace.rectangle

  let scale = Math.min(1 / (rect.maxX - rect.minX), 1 / (rect.maxY - rect.minY))

  let origin = {
    x: (rect.minX + rect.maxX) / 2,
    y: (rect.minY + rect.maxY) / 2,
  }

  const SpecificFaceImage =
    imageFace.media.thumbnail.width > imageFace.media.thumbnail.height
      ? FaceImageLandscape
      : FaceImagePortrait
  return (
    <CircleImageWrapper size={size}>
      <SpecificFaceImage
        $selectable={selectable}
        $scale={scale}
        $origin={origin}
        src={imageFace.media.thumbnail.url}
      />
    </CircleImageWrapper>
  )
}

FaceCircleImage.propTypes = {
  imageFace: PropTypes.object.isRequired,
  selectable: PropTypes.bool,
  size: PropTypes.string,
}

export default FaceCircleImage
