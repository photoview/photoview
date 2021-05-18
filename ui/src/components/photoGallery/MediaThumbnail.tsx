import React, { useCallback, useState } from 'react'
import styled from 'styled-components'
import { Icon } from 'semantic-ui-react'
import { ProtectedImage } from './ProtectedMedia'
import { MediaType } from '../../../__generated__/globalTypes'

const MediaContainer = styled.div`
  flex-grow: 1;
  flex-basis: 0;
  height: 200px;
  margin: 4px;
  background-color: #eee;
  position: relative;
  overflow: hidden;
`

const StyledPhoto = styled(ProtectedImage)<{ loaded: boolean }>`
  height: 200px;
  min-width: 100%;
  position: relative;
  object-fit: cover;
  opacity: ${({ loaded }) => (loaded ? 1 : 0)};

  transition: opacity 300ms;
`

type LazyPhotoProps = {
  src?: string
}

const LazyPhoto = (photoProps: LazyPhotoProps) => {
  const [loaded, setLoaded] = useState(false)
  const onLoad = useCallback(e => {
    !e.target.dataset.src && setLoaded(true)
  }, [])

  return (
    <StyledPhoto {...photoProps} lazyLoading loaded={loaded} onLoad={onLoad} />
  )
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

const HoverIcon = styled(Icon)`
  font-size: 1.5em !important;
  margin: 160px 10px 0 10px !important;
  color: white !important;
  text-shadow: 0 0 4px black;
  opacity: 0 !important;
  position: relative;

  border-radius: 50%;
  width: 34px !important;
  height: 34px !important;
  padding-top: 7px;

  ${MediaContainer}:hover & {
    opacity: 1 !important;
  }

  &:hover {
    background-color: rgba(255, 255, 255, 0.4);
  }

  transition: opacity 100ms, background-color 100ms;
`

const FavoriteIcon = styled(HoverIcon)`
  float: right;
  opacity: ${({ favorite }) => (favorite ? '0.8' : '0.2')} !important;
`

const SidebarIcon = styled(HoverIcon)`
  margin: 10px !important;
  position: absolute;
  top: 0;
  right: 0;
`

const VideoThumbnailIcon = styled(Icon)`
  color: rgba(255, 255, 255, 0.8);
  position: absolute;
  left: calc(50% - 16px);
  top: calc(50% - 13px);
`

type MediaThumbnailProps = {
  media: {
    id: string
    type: MediaType
    favorite?: boolean
    thumbnail: null | {
      url: string
      width: number
      height: number
    }
  }
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
        favorite={media.favorite.toString()}
        name={media.favorite ? 'heart' : 'heart outline'}
        onClick={(event: MouseEvent) => {
          event.stopPropagation()
          clickFavorite()
        }}
      />
    )
  }

  let videoIcon = null
  if (media.type == MediaType.Video) {
    videoIcon = <VideoThumbnailIcon name="play" size="big" />
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
          minWidth: `${minWidth}px`,
          height: `200px`,
        }}
      >
        <LazyPhoto src={media.thumbnail?.url} />
      </div>
      <PhotoOverlay active={active}>
        {videoIcon}
        <SidebarIcon
          name="info"
          onClick={(e: MouseEvent) => {
            e.stopPropagation()
            selectImage()
          }}
        />
        {heartIcon}
      </PhotoOverlay>
    </MediaContainer>
  )
}

export const PhotoThumbnail = styled.div`
  flex-grow: 1;
  height: 200px;
  width: 300px;
  margin: 4px;
  background-color: #eee;
  position: relative;
`
