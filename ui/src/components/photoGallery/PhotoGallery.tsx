import React, { useContext } from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { MediaThumbnail, PhotoThumbnail } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import { SidebarContext } from '../sidebar/Sidebar'
import { useTranslation } from 'react-i18next'
import { PresentMediaProps_Media } from './presentView/PresentMedia'
import { sidebarPhoto_media_thumbnail } from '../sidebar/__generated__/sidebarPhoto'
import {
  PhotoGalleryAction,
  PhotoGalleryState,
  selectImageAction,
} from './photoGalleryReducer'
import { gql, useMutation } from '@apollo/client'
import {
  markMediaFavorite,
  markMediaFavoriteVariables,
} from './__generated__/markMediaFavorite'

const markFavoriteMutation = gql`
  mutation markMediaFavorite($mediaId: ID!, $favorite: Boolean!) {
    favoriteMedia(mediaId: $mediaId, favorite: $favorite) {
      id
      favorite
    }
  }
`

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  min-height: 200px;
  position: relative;
  margin: -4px;

  @media (max-width: 1000px) {
    /* Compensate for tab bar on mobile */
    margin-bottom: 76px;
  }
`

const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

const ClearWrap = styled.div`
  clear: both;
`

export interface PhotoGalleryProps_Media extends PresentMediaProps_Media {
  thumbnail: sidebarPhoto_media_thumbnail | null
  favorite?: boolean
}

type PhotoGalleryProps = {
  loading: boolean
  mediaState: PhotoGalleryState
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
}

const PhotoGallery = ({
  mediaState,
  loading,
  dispatchMedia,
}: PhotoGalleryProps) => {
  const { t } = useTranslation()

  const { updateSidebar } = useContext(SidebarContext)

  const [markFavorite] = useMutation<
    markMediaFavorite,
    markMediaFavoriteVariables
  >(markFavoriteMutation)

  const { media, activeIndex, presenting } = mediaState

  let photoElements = []
  if (media) {
    photoElements = media.map((media, index) => {
      const active = activeIndex == index

      return (
        <MediaThumbnail
          key={media.id}
          media={media}
          active={active}
          selectImage={() => {
            selectImageAction({
              index,
              mediaState,
              dispatchMedia,
              updateSidebar,
            })
          }}
          clickFavorite={() => {
            markFavorite({
              variables: {
                mediaId: media.id,
                favorite: !media.favorite,
              },
              optimisticResponse: {
                favoriteMedia: {
                  id: media.id,
                  favorite: !media.favorite,
                  __typename: 'Media',
                },
              },
            })
          }}
          clickPresent={() => {
            dispatchMedia({
              type: 'setPresenting',
              presenting: true,
            })
          }}
        />
      )
    })
  } else {
    for (let i = 0; i < 6; i++) {
      photoElements.push(<PhotoThumbnail key={i} />)
    }
  }

  return (
    <ClearWrap>
      <Gallery>
        <Loader active={loading}>
          {t('general.loading.media', 'Loading media')}
        </Loader>
        {photoElements}
        <PhotoFiller />
      </Gallery>
      {presenting && (
        <PresentView mediaState={mediaState} dispatchMedia={dispatchMedia} />
      )}
    </ClearWrap>
  )
}

export default PhotoGallery
