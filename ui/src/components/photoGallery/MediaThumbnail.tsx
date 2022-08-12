import React from 'react'
import styled from 'styled-components'
import { ProtectedImage } from './ProtectedMedia'
import { MediaType } from '../../__generated__/globalTypes'
import { ReactComponent as VideoThumbnailIconSVG } from './icons/videoThumbnailIcon.svg'
import { MediaGalleryFields } from './__generated__/MediaGalleryFields'

const MediaContainer = styled.div`
  flex-grow: 1;
  flex-basis: 0;
  height: 200px;
  margin: 4px;
  background-color: #eee;
  position: relative;
  overflow: hidden;
`

const StyledPhoto = styled(ProtectedImage)`
  height: 200px;
  min-width: 100%;
  position: relative;
  object-fit: cover;

  transition: opacity 300ms;
`

type LazyPhotoProps = {
  src?: string
  blurhash: string | null
}

const LazyPhoto = (photoProps: LazyPhotoProps) => {
  return <StyledPhoto {...photoProps} lazyLoading />
}

const PhotoOverlay = styled.div<{ active: boolean }>`
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;

  ${({ active }) =>
    active &&
    `
      outline: 4px solid rgba(65, 131, 196, 0.6);
      outline-offset: -4px;
    `}
`

const HoverIcon = styled.button`
  font-size: 1.5em;
  margin: 160px 10px 0 10px;
  color: white;
  text-shadow: 0 0 4px black;
  opacity: 0;
  position: relative;

  border-radius: 50%;
  width: 34px;
  height: 34px;

  ${MediaContainer}:hover &, ${MediaContainer}:focus-within & {
    opacity: 1 !important;
  }

  &:hover,
  &:focus {
    background-color: rgba(0, 0, 0, 0.4);
  }

  transition: opacity 100ms, background-color 100ms;
`

type FavoriteIconProps = {
  favorite: boolean
  onClick(e: React.MouseEvent<HTMLButtonElement, MouseEvent>): void
}

const FavoriteIcon = ({ favorite, onClick }: FavoriteIconProps) => {
  return (
    <HoverIcon
      onClick={onClick}
      style={{ opacity: favorite ? '0.75' : undefined }}
    >
      <svg
        className="text-white m-auto mt-1"
        width="19px"
        height="17px"
        viewBox="0 0 19 17"
        version="1.1"
      >
        <path
          d="M13.999086,1 C15.0573371,1 16.0710089,1.43342987 16.8190212,2.20112483 C17.5765039,2.97781012 18,4.03198704 18,5.13009709 C18,6.22820714 17.5765039,7.28238406 16.8188574,8.05923734 L16.8188574,8.05923734 L15.8553647,9.04761889 L9.49975689,15.5674041 L3.14414912,9.04761889 L2.18065643,8.05923735 C1.39216493,7.2503776 0.999999992,6.18971057 1,5.13009711 C1.00000001,4.07048366 1.39216496,3.00981663 2.18065647,2.20095689 C2.95931483,1.40218431 3.97927681,1.00049878 5.00042783,1.00049878 C6.02157882,1.00049878 7.04154078,1.4021843 7.82019912,2.20095684 L7.82019912,2.20095684 L9.4997569,3.92390079 L11.1794784,2.20078881 C11.9271631,1.43342987 12.9408349,1 13.999086,1 L13.999086,1 Z"
          fill={favorite ? 'currentColor' : 'none'}
          stroke="currentColor"
          strokeWidth={favorite ? '0' : '2'}
        ></path>
      </svg>
    </HoverIcon>
  )
}

type SidebarIconProps = {
  onClick(e: React.MouseEvent<HTMLButtonElement, MouseEvent>): void
}

const SidebarIcon = ({ onClick }: SidebarIconProps) => (
  <SidebarIconWrapper onClick={onClick}>
    <svg
      width="20px"
      height="20px"
      viewBox="0 0 20 20"
      version="1.1"
      className="m-auto"
    >
      <path
        d="M10,0 C15.5228475,0 20,4.4771525 20,10 C20,15.5228475 15.5228475,20 10,20 C4.4771525,20 0,15.5228475 0,10 C0,4.4771525 4.4771525,0 10,0 Z M10,9 C9.44771525,9 9,9.44771525 9,10 L9,10 L9,14 L9.00672773,14.1166211 C9.06449284,14.6139598 9.48716416,15 10,15 C10.5522847,15 11,14.5522847 11,14 L11,14 L11,10 L10.9932723,9.88337887 C10.9355072,9.38604019 10.5128358,9 10,9 Z M10.01,5 L9.88337887,5.00672773 C9.38604019,5.06449284 9,5.48716416 9,6 C9,6.55228475 9.44771525,7 10,7 L10,7 L10.1266211,6.99327227 C10.6239598,6.93550716 11.01,6.51283584 11.01,6 C11.01,5.44771525 10.5622847,5 10.01,5 L10.01,5 Z"
        fill="#FFFFFF"
      ></path>
    </svg>
  </SidebarIconWrapper>
)

const SidebarIconWrapper = styled(HoverIcon)`
  margin: 10px !important;
  position: absolute;
  top: 0;
  right: 0;
`

const VideoThumbnailIcon = styled(VideoThumbnailIconSVG)`
  color: rgba(255, 255, 255, 0.8);
  position: absolute;
  left: calc(50% - 17.5px);
  top: calc(50% - 22px);
`

type MediaThumbnailProps = {
  media: MediaGalleryFields
  active: boolean
  selectImage(): void
  clickPresent(): void
  clickFavorite(): void
}

export const MediaThumbnail = ({
  media,
  active,
  selectImage,
  clickPresent,
  clickFavorite,
}: MediaThumbnailProps) => {
  let heartIcon = null
  if (media.favorite !== undefined) {
    heartIcon = (
      <FavoriteIcon
        favorite={media.favorite}
        onClick={e => {
          e.stopPropagation()
          clickFavorite()
        }}
      />
    )
  }

  let videoIcon = null
  if (media.type == MediaType.Video) {
    videoIcon = <VideoThumbnailIcon />
  }

  let minWidth = 100
  if (media.thumbnail) {
    minWidth = Math.floor(
      (media.thumbnail.width / media.thumbnail.height) * 200
    )
  }

  return (
    <MediaContainer
      key={media.id}
      style={{
        cursor: 'pointer',
        minWidth: `clamp(124px, ${minWidth}px, 100% - 8px)`,
      }}
      onClick={() => {
        clickPresent()
      }}
    >
      <div
        style={{
          // minWidth: `min(${minWidth}px, 100%)`,
          minWidth: `${minWidth}px`,
          height: `200px`,
        }}
      >
        <LazyPhoto src={media.thumbnail?.url} blurhash={media.blurhash} />
      </div>
      <PhotoOverlay active={active}>
        {videoIcon}
        <SidebarIcon
          onClick={e => {
            e.stopPropagation()
            selectImage()
          }}
        />
        {heartIcon}
      </PhotoOverlay>
    </MediaContainer>
  )
}

export const MediaPlaceholder = styled.div`
  flex-grow: 1;
  height: 200px;
  width: 300px;
  margin: 4px;
  background-color: #eee;
  position: relative;
`
