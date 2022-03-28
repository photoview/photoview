import React from 'react'
import Layout from '../../components/layout/Layout'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'
import styled from 'styled-components'
import { gql, useQuery } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import useURLParameters from '../../hooks/useURLParameters'
import useOrderingParams from '../../hooks/useOrderingParams'
import { shareAlbumQuery } from './__generated__/shareAlbumQuery'
import useScrollPagination from '../../hooks/useScrollPagination'
import PaginateLoader from '../../components/PaginateLoader'

export const SHARE_ALBUM_QUERY = gql`
  query shareAlbumQuery(
    $id: ID!
    $token: String!
    $password: String
    $mediaOrderBy: String
    $mediaOrderDirection: OrderDirection
    $limit: Int
    $offset: Int
  ) {
    album(id: $id, tokenCredentials: { token: $token, password: $password }) {
      id
      title
      subAlbums(order: { order_by: "title" }) {
        id
        title
        thumbnail {
          id
          thumbnail {
            url
          }
        }
      }
      media(
        paginate: { limit: $limit, offset: $offset }
        order: {
          order_by: $mediaOrderBy
          order_direction: $mediaOrderDirection
        }
      ) {
        id
        title
        type
        blurhash
        thumbnail {
          url
          width
          height
        }
        downloads {
          title
          mediaUrl {
            url
            width
            height
            fileSize
          }
        }
        highRes {
          url
          width
          height
        }
        videoWeb {
          url
        }
        exif {
          id
          description
          camera
          maker
          lens
          dateShot
          exposure
          aperture
          iso
          focalLength
          flash
          exposureProgram
          coordinates {
            latitude
            longitude
          }
        }
      }
    }
  }
`

const AlbumSharePageWrapper = styled.div`
  height: 100%;
`

type AlbumSharePageProps = {
  albumID: string
  token: string
  password: string | null
}

const AlbumSharePage = ({ albumID, token, password }: AlbumSharePageProps) => {
  const { t } = useTranslation()

  const urlParams = useURLParameters()
  const orderParams = useOrderingParams(urlParams)

  const { data, error, loading, fetchMore } = useQuery<shareAlbumQuery>(
    SHARE_ALBUM_QUERY,
    {
      variables: {
        id: albumID,
        token,
        password,
        limit: 200,
        offset: 0,
        mediaOrderBy: orderParams.orderBy,
        mediaOrderDirection: orderParams.orderDirection,
      },
    }
  )

  const { containerElem, finished: finishedLoadingMore } =
    useScrollPagination<shareAlbumQuery>({
      loading,
      fetchMore,
      data,
      getItems: data => data.album.media,
    })

  if (error) {
    return <div>{error.message}</div>
  }

  const album = data?.album

  return (
    <AlbumSharePageWrapper data-testid="AlbumSharePage">
      <Layout
        title={
          album ? album.title : t('general.loading.album', 'Loading album')
        }
      >
        <AlbumGallery
          ref={containerElem}
          album={album}
          customAlbumLink={albumId => `/share/${token}/${albumId}`}
          showFilter
          setOrdering={orderParams.setOrdering}
          ordering={orderParams}
        />
        <PaginateLoader
          active={!finishedLoadingMore && !loading}
          text={t('general.loading.paginate.media', 'Loading more media')}
        />
      </Layout>
    </AlbumSharePageWrapper>
  )
}

export default AlbumSharePage
