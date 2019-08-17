import React from 'react'
import RouterProps from 'react-router-prop-types'
import { Route, Switch } from 'react-router-dom'
import AlbumSharePage from './AlbumSharePage'
import PhotoSharePage from './PhotoSharePage'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'

const tokenQuery = gql`
  query SharePageToken($token: ID!) {
    shareToken(token: $token) {
      token
      album {
        id
        title
        photos(orderBy: title_desc) {
          ...PhotoProps
        }
      }
      photo {
        ...PhotoProps
      }
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
    original {
      url
    }
    exif {
      camera
      maker
      lens
      dateShot {
        formatted
      }
      fileSize
      exposure
      aperture
      iso
      focalLength
      flash
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
                return <AlbumSharePage album={data.shareToken.album} />
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
