import React from 'react'
import styled from 'styled-components'
import {
  ProtectedImage,
  ProtectedImageProps,
} from '../../components/photoGallery/ProtectedMedia'
import {
  myFaces_myFaceGroups_imageFaces_media,
  myFaces_myFaceGroups_imageFaces_rectangle,
} from './__generated__/myFaces'

type FaceImageProps = ProtectedImageProps & {
  origin: { x: number; y: number }
  selectable: boolean
  scale: number
}

const FaceImage = styled(
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  ({ origin, selectable, scale, ...rest }: FaceImageProps) => (
    <ProtectedImage {...rest} />
  )
)`
  position: absolute;
  transform-origin: ${({ origin }) => `${origin.x * 100}% ${origin.y * 100}%`};
  object-fit: cover;

  transition: transform 250ms ease-out;
`

const FaceImagePortrait = styled(FaceImage)`
  width: 100%;
  top: 50%;
  transform: translateY(-50%)
    ${({ origin, scale }) =>
      `translate(${(0.5 - origin.x) * 100}%, ${
        (0.5 - origin.y) * 100
      }%) scale(${Math.max(scale * 0.8, 1)})`};

  ${({ selectable, origin, scale }) =>
    selectable
      ? `
    &:hover {
      transform: translateY(-50%) translate(${(0.5 - origin.x) * 100}%, ${
          (0.5 - origin.y) * 100
        }%) scale(${Math.max(scale * 0.85, 1)})
      `
      : ''}
`

const FaceImageLandscape = styled(FaceImage)`
  height: 100%;
  left: 50%;
  transform: translateX(-50%)
    ${({ origin, scale }) =>
      `translate(${(0.5 - origin.x) * 100}%, ${
        (0.5 - origin.y) * 100
      }%) scale(${Math.max(scale * 0.8, 1)})`};

  ${({ selectable, origin, scale }) =>
    selectable
      ? `
    &:hover {
      transform: translateX(-50%) translate(${(0.5 - origin.x) * 100}%, ${
          (0.5 - origin.y) * 100
        }%) scale(${Math.max(scale * 0.85, 1)})
      `
      : ''}
`

const CircleImageWrapper = styled.div<{ size: string }>`
  background-color: #eee;
  position: relative;
  border-radius: 50%;
  width: ${({ size }) => size};
  height: ${({ size }) => size};
  object-fit: fill;
  overflow: hidden;
`

type FaceCircleImageFace = {
  __typename: 'ImageFace'
  id: string
  rectangle: myFaces_myFaceGroups_imageFaces_rectangle
  media: myFaces_myFaceGroups_imageFaces_media
}

type FaceCircleImageProps = {
  imageFace: FaceCircleImageFace
  selectable: boolean
  size?: string
}

const FaceCircleImage = ({
  imageFace,
  selectable,
  size = '150px',
}: FaceCircleImageProps) => {
  if (!imageFace) {
    return null
  }

  const rect = imageFace.rectangle

  const scale = Math.min(
    1 / (rect.maxX - rect.minX),
    1 / (rect.maxY - rect.minY)
  )

  const origin = {
    x: (rect.minX + rect.maxX) / 2,
    y: (rect.minY + rect.maxY) / 2,
  }

  let SpecificFaceImage: typeof FaceImageLandscape | typeof FaceImagePortrait =
    FaceImageLandscape
  if (imageFace.media.thumbnail) {
    SpecificFaceImage =
      imageFace.media.thumbnail.width > imageFace.media.thumbnail.height
        ? FaceImageLandscape
        : FaceImagePortrait
  }

  return (
    <CircleImageWrapper size={size}>
      <SpecificFaceImage
        selectable={selectable}
        scale={scale}
        origin={origin}
        src={imageFace.media.thumbnail?.url}
      />
    </CircleImageWrapper>
  )
}

export default FaceCircleImage
