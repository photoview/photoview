import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import styled from 'styled-components'
import { Link } from 'react-router-dom'
import { Dimmer, Loader } from 'semantic-ui-react'

const getAlbumsQuery = gql`
  query getMyAlbums {
    myAlbums {
      id
      title
      photos {
        thumbnail {
          path
        }
      }
    }
  }
`

const AlbumBoxLink = styled(Link)`
  width: 240px;
  height: 240px;
  display: inline-block;
  text-align: center;
  color: #222;
`

const AlbumBoxImage = styled.div`
  width: 220px;
  height: 220px;
  margin: auto;
  border-radius: 4%;
  background-image: url('${props => props.image}');
  background-color: #eee;
  background-size: cover;
  background-position: center;
`

class Albums extends Component {
  render() {
    return (
      <div style={{ position: 'relative', minHeight: '500px' }}>
        <Query query={getAlbumsQuery}>
          {({ loading, error, data }) => {
            // if (loading) return <Loader active />
            if (error) return <div>Error {error.message}</div>

            let albums

            if (data && data.myAlbums) {
              albums = data.myAlbums.map(album => (
                <AlbumBoxLink key={album.id} to={`/album/${album.id}`}>
                  <AlbumBoxImage image={album.photos[0].thumbnail.path} />
                  <p>{album.title}</p>
                </AlbumBoxLink>
              ))
            } else {
              albums = []
              for (let i = 0; i < 8; i++) {
                albums.push(
                  <AlbumBoxLink key={i} to="#">
                    <AlbumBoxImage />
                  </AlbumBoxLink>
                )
              }
            }

            return (
              <div>
                {' '}
                <Loader active={loading}>Loading images</Loader>
                {albums}
              </div>
            )
          }}
        </Query>
      </div>
    )
  }
}

export default Albums
