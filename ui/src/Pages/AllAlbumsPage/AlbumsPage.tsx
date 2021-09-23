import React from 'react'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../components/layout/Layout'
import { useQuery, gql } from '@apollo/client'
import { getMyAlbums } from './__generated__/getMyAlbums'

const getAlbumsQuery = gql`
  query getMyAlbums {
    myAlbums(order: { order_by: "title" }, onlyRoot: true, showEmpty: true) {
      id
      title
      thumbnail {
        id
        thumbnail {
          url
        }
      }
    }
  }
`

const AlbumsPage = () => {
  const { error, data } = useQuery<getMyAlbums>(getAlbumsQuery)

  return (
    <Layout title="Albums">
      <AlbumBoxes error={error} albums={data?.myAlbums} />
    </Layout>
  )
}

export default AlbumsPage
