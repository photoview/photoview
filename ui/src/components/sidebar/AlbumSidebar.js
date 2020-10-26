import React from 'react'
import PropTypes from 'prop-types'
import { Query } from '@apollo/client/react/components'
import { gql } from '@apollo/client'
import SidebarShare from './Sharing'

const albumQuery = gql`
  query getAlbumSidebar($id: Int!) {
    album(id: $id) {
      id
      title
    }
  }
`

const AlbumSidebar = ({ albumId }) => {
  return (
    <div>
      <p>Album options</p>
      <Query query={albumQuery} variables={{ id: albumId }}>
        {({ loading, error, data }) => {
          if (loading) return <div>Loading...</div>
          if (error) return <div>{error.message}</div>

          console.log('ALBUM', data.album)

          return (
            <div>
              <h1>{data.album.title}</h1>
              <SidebarShare album={data.album} />
            </div>
          )
        }}
      </Query>
    </div>
  )
}

AlbumSidebar.propTypes = {
  albumId: PropTypes.number.isRequired,
}

export default AlbumSidebar
