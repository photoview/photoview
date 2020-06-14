import React from 'react'
import PropTypes from 'prop-types'
import RouterProps from 'react-router-prop-types'
import { Route, Switch } from 'react-router-dom'
import AlbumSharePage from './AlbumSharePage'
import PhotoSharePage from './PhotoSharePage'
import { useQuery } from 'react-apollo'
import gql from 'graphql-tag'

const tokenQuery = gql`
  query SharePageToken($token: String!, $password: String) {
    shareToken(token: $token, password: $password) {
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

const tokenPasswordProtectedQuery = gql`
  query ShareTokenRequiresPassword($token: String!) {
    shareTokenRequiresPassword(token: $token)
  }
`

const AuthorizedTokenRoute = ({ match, password }) => {
  const { loading, error, data } = useQuery(tokenQuery, {
    variables: { token: match.params.token, password },
  })

  if (error) return error.message
  if (loading) return 'Loading...'

  if (data.shareToken.album) {
    return <AlbumSharePage album={data.shareToken.album} match={match} />
  }

  if (data.shareToken.photo) {
    return <PhotoSharePage photo={data.shareToken.photo} />
  }

  return <h1>Share not found</h1>
}

AuthorizedTokenRoute.propTypes = {
  match: PropTypes.object.isRequired,
  password: PropTypes.string,
}

const TokenRoute = ({ match }) => {
  const { loading, error, data } = useQuery(tokenPasswordProtectedQuery, {
    variables: { token: match.params.token },
  })

  if (error) return error.message
  if (loading) return 'Loading...'

  if (data.shareTokenRequiresPassword == true) {
    return 'Please provide password'
  }

  return <AuthorizedTokenRoute match={match} />
}

TokenRoute.propTypes = {
  match: PropTypes.object.isRequired,
}

const SharePage = ({ match }) => {
  return (
    <Switch>
      <Route path={`${match.url}/:token`}>
        {({ match }) => <TokenRoute match={match} />}
      </Route>
      <Route path="/">Share not found</Route>
    </Switch>
  )
}

SharePage.propTypes = {
  ...RouterProps,
}

export default SharePage
