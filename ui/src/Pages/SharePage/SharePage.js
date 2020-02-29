import React from 'react'
import RouterProps from 'react-router-prop-types'
import { Route, Switch } from 'react-router-dom'
import AlbumSharePage from './AlbumSharePage'
import PhotoSharePage from './PhotoSharePage'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'

const tokenQuery = gql`
  query SharePageToken($token: String!) {
    shareToken(token: $token) {
      token
      album {
        ...AlbumProps
        subAlbums {
          ...AlbumProps
          subAlbums {
            ...AlbumProps
          }
        }
      }
      photo {
        ...PhotoProps
      }
    }
  }

  fragment AlbumProps on Album {
    id
    title
    thumbnail {
      thumbnail {
        url
      }
    }
    photos(filter: { order_by: "title", order_direction: DESC }) {
      ...PhotoProps
    }
  }

  fragment PhotoProps on Photo {
    id
    title
    thumbnail {
      url
      width
      height
    }
    downloads {
      title
      url
      width
      height
    }
    highRes {
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
`

const SharePage = ({ match }) => {
  return (
    <Switch>
      <Route path={`${match.url}/:token`}>
        {({ match }) => (
          <Query query={tokenQuery} variables={{ token: match.params.token }}>
            {({ loading, error, data }) => {
              if (error) return error.message
              if (loading) return 'Loading...'

              if (data.shareToken.album) {
                return (
                  <AlbumSharePage album={data.shareToken.album} match={match} />
                )
              }

              if (data.shareToken.photo) {
                return <PhotoSharePage photo={data.shareToken.photo} />
              }

              return <h1>Share not found</h1>
            }}
          </Query>
        )}
      </Route>
      <Route path="/">Share not found</Route>
    </Switch>
  )
}

SharePage.propTypes = {
  ...RouterProps,
}

export default SharePage
