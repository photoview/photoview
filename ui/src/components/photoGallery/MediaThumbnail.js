import React, { useState } from 'react'
import { useMutation, gql } from '@apollo/client'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import LazyLoad from 'react-lazyload'
import { Icon } from 'semantic-ui-react'
import { ProtectedImage } from './ProtectedMedia'

const markFavoriteMutation = gql`
  mutation markMediaFavorite($mediaId: ID!, $favorite: Boolean!) {
    favoriteMedia(mediaId: $mediaId, favorite: $favorite) {
      id
      favorite
    }
  }
`

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
  opacity: ${({ loaded }) => (loaded ? 1 : 0)};

  transition: opacity 300ms;
`

const PhotoImg = photoProps => {
  const [loaded, setLoaded] = useState(false)

  return (
    <StyledPhoto
      {...photoProps}
      loaded={loaded ? 1 : 0}
      onLoad={() => {
        setLoaded(true)
      }}
    />
  )
}

const LazyPhoto = React.memo(
  props => {
    return (
      <LazyLoad scrollContainer="#layout-content">
        <PhotoImg {...props} />
      </LazyLoad>
    )
  },
  (prevProps, nextProps) => prevProps.src === nextProps.src
)

LazyPhoto.propTypes = {
  src: PropTypes.string,
}

const PhotoOverlay = styled.div`
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;

  ${props =>
    props.active &&
    `
      border: 4px solid rgba(65, 131, 196, 0.6);

      & ${HoverIcon} {
        top: -4px !important;
        left: -4px !important;
      }
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

const VideoThumbnailIcon = styled(Icon)`
  color: rgba(255, 255, 255, 0.8);
  position: absolute;
  left: calc(50% - 16px);
  top: calc(50% - 13px);
`

export const MediaThumbnail = ({
  media,
  onSelectImage,
  minWidth,
  index,
  active,
  setPresenting,
  onFavorite,
}) => {
  const [markFavorite] = useMutation(markFavoriteMutation)

  let heartIcon = null
  if (typeof media.favorite == 'boolean') {
    heartIcon = (
      <FavoriteIcon
        favorite={media.favorite.toString()}
        name={media.favorite ? 'heart' : 'heart outline'}
        onClick={event => {
          event.stopPropagation()
          const favorite = !media.favorite
          markFavorite({
            variables: {
              mediaId: media.id,
              favorite: favorite,
            },
            optimisticResponse: {
              favoriteMedia: {
                id: media.id,
                favorite: favorite,
                __typename: 'Media',
              },
            },
          })
          onFavorite && onFavorite()
        }}
      />
    )
  }

  let videoIcon = null
  if (media.type == 'video') {
    videoIcon = <VideoThumbnailIcon name="play" size="big" />
  }

  return (
    <MediaContainer
      key={media.id}
      style={{
        cursor: onSelectImage ? 'pointer' : null,
        minWidth: `clamp(124px, ${minWidth}px, 100% - 8px)`,
      }}
      onClick={() => {
        onSelectImage && onSelectImage(index)
      }}
    >
      <LazyPhoto src={media.thumbnail && media.thumbnail.url} />
      <PhotoOverlay active={active}>
        {videoIcon}
        <HoverIcon
          name="expand"
          onClick={() => {
            setPresenting(true)
          }}
        />
        {heartIcon}
      </PhotoOverlay>
    </MediaContainer>
  )
}

MediaThumbnail.propTypes = {
  media: PropTypes.object.isRequired,
  onSelectImage: PropTypes.func,
  minWidth: PropTypes.number.isRequired,
  index: PropTypes.number.isRequired,
  active: PropTypes.bool.isRequired,
  setPresenting: PropTypes.func.isRequired,
  onFavorite: PropTypes.func,
}

export const PhotoThumbnail = styled.div`
  flex-grow: 1;
  height: 200px;
  width: 300px;
  margin: 4px;
  background-color: #eee;
  position: relative;
`
