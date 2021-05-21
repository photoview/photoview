import React, { useEffect } from 'react'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../components/layout/Layout'
import { useQuery, gql } from '@apollo/client'
import LazyLoad from '../../helpers/LazyLoad'
import { useTranslation } from 'react-i18next'

const getAlbumsQuery = gql`
  query getMyAlbums {
    myAlbums(order: { order_by: "title" }, onlyRoot: true, showEmpty: true) {
      id
      title
      thumbnail {
        thumbnail {
          url
        }
      }
    }
  }
`

const AlbumsPage = () => {
  const { t } = useTranslation()
  const { loading, error, data } = useQuery(getAlbumsQuery)

  useEffect(() => {
    return () => LazyLoad.disconnect()
  }, [])

  useEffect(() => {
    !loading && LazyLoad.loadImages(document.querySelectorAll('img[data-src]'))
  }, [loading])

  return (
    <Layout title="Albums">
      <h1>{t('albums_page.title', 'Albums')}</h1>
      {!loading && (
        <AlbumBoxes
          loading={loading}
          error={error}
          albums={data && data.myAlbums}
        />
      )}
    </Layout>
  )
}

export default AlbumsPage
