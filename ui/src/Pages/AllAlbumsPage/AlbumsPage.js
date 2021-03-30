import React, { useEffect } from 'react'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../Layout'
import { useQuery, gql } from '@apollo/client'
import lazyLoad from '../../helpers/LazyLoad'

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
  const { loading, error, data } = useQuery(getAlbumsQuery)

  useEffect(() => {
    if (!loading) {
      lazyLoad(document.querySelectorAll('img[data-src]'))
    }
  }, [loading])

  return (
    <Layout title="Albums">
      <h1>Albums</h1>
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
