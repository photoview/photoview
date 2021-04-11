import PropTypes from 'prop-types'
import React from 'react'
import Layout from '../../Layout'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'
import styled from 'styled-components'
import { gql, useQuery } from '@apollo/client'
import { useTranslation } from 'react-i18next'

export const SHARE_ALBUM_QUERY = gql`
  query shareAlbumQuery(
    $id: ID!
    $token: String!
    $password: String
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
          thumbnail {
            url
          }
        }
      }
      media(paginate: { limit: $limit, offset: $offset }) {
        id
        title
        type
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
        }
      }
    }
  }
`

const AlbumSharePageWrapper = styled.div`
  height: 100%;
`

const AlbumSharePage = ({ albumID, token, password }) => {
  const { t } = useTranslation()
  const { data, loading, error } = useQuery(SHARE_ALBUM_QUERY, {
    variables: {
      id: albumID,
      token,
      password,
      limit: 200,
      offset: 0,
    },
  })

  if (error) {
    return error.message
  }

  if (loading) {
    return t('general.loading.default', 'Loading...')
  }

  const album = data.album

  return (
    <AlbumSharePageWrapper data-testid="AlbumSharePage">
      <Layout
        title={
          album ? album.title : t('general.loading.album', 'Loading album')
        }
      >
        <AlbumGallery
          album={album}
          customAlbumLink={albumId => `/share/${token}/${albumId}`}
        />
      </Layout>
    </AlbumSharePageWrapper>
  )
}

AlbumSharePage.propTypes = {
  albumID: PropTypes.string.isRequired,
  token: PropTypes.string.isRequired,
  password: PropTypes.string,
}

export default AlbumSharePage
