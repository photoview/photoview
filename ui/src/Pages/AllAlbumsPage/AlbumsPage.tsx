import React, { useEffect } from 'react'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../components/layout/Layout'
import { useQuery, gql } from '@apollo/client'
import LazyLoad from '../../helpers/LazyLoad'
import { getMyAlbums } from './__generated__/getMyAlbums'

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
  const { loading, error, data } = useQuery<getMyAlbums>(getAlbumsQuery)

  useEffect(() => {
    return () => LazyLoad.disconnect()
  }, [])

  useEffect(() => {
    !loading && LazyLoad.loadImages(document.querySelectorAll('img[data-src]'))
  }, [loading])

  return (
    <Layout title="Albums">
      <AlbumBoxes error={error} albums={data?.myAlbums} />
    </Layout>
  )
}

export default AlbumsPage
