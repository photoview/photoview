import React, { useState } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import RouterProps from 'react-router-prop-types'
import { Route, Switch } from 'react-router-dom'
import AlbumSharePage from './AlbumSharePage'
import PhotoSharePage from './PhotoSharePage'
import { useQuery } from 'react-apollo'
import gql from 'graphql-tag'
import {
  Container,
  Header,
  Form,
  Button,
  Input,
  Icon,
  Message,
} from 'semantic-ui-react'

const shareTokenQuery = gql`
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
      media {
        ...MediaProps
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
    media(filter: { order_by: "title", order_direction: DESC }) {
      ...MediaProps
    }
  }

  fragment MediaProps on Media {
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

const validateTokenPasswordQuery = gql`
  query ShareTokenValidatePassword($token: String!, $password: String) {
    shareTokenValidatePassword(token: $token, password: $password)
  }
`

const AuthorizedTokenRoute = ({ match }) => {
  const token = match.params.token

  const { loading, error, data } = useQuery(shareTokenQuery, {
    variables: {
      token,
      password: sessionStorage.getItem(`share-token-pw-${token}`),
    },
  })

  if (error) return error.message
  if (loading) return 'Loading...'

  if (data.shareToken.album) {
    return <AlbumSharePage album={data.shareToken.album} match={match} />
  }

  if (data.shareToken.media) {
    return <PhotoSharePage photo={data.shareToken.media} />
  }

  return <h1>Share not found</h1>
}

AuthorizedTokenRoute.propTypes = {
  match: PropTypes.object.isRequired,
}

const MessageContainer = styled.div`
  max-width: 400px;
  margin: 100px auto 0;
`

const ProtectedTokenEnterPassword = ({
  match,
  refetchWithPassword,
  loading = false,
}) => {
  const [passwordValue, setPasswordValue] = useState('')
  const [invalidPassword, setInvalidPassword] = useState(false)

  const onSubmit = () => {
    refetchWithPassword(passwordValue)
    setInvalidPassword(true)
  }

  let errorMessage = null
  if (invalidPassword && !loading) {
    errorMessage = (
      <Message negative>
        <Message.Content>Wrong password, please try again.</Message.Content>
      </Message>
    )
  }

  return (
    <MessageContainer>
      <Header as="h1" style={{ fontWeight: 400 }}>
        Protected share
      </Header>
      <p>This share is protected with a password.</p>
      <Form>
        <Form.Field>
          <label>Password</label>
          <Input
            loading={loading}
            disabled={loading}
            onKeyUp={event => event.key == 'Enter' && onSubmit()}
            onChange={e => setPasswordValue(e.target.value)}
            placeholder="Password"
            type="password"
            icon={<Icon onClick={onSubmit} link name="arrow right" />}
          />
        </Form.Field>
        {errorMessage}
      </Form>
    </MessageContainer>
  )
}

ProtectedTokenEnterPassword.propTypes = {
  match: PropTypes.object.isRequired,
  refetchWithPassword: PropTypes.func.isRequired,
  loading: PropTypes.bool,
}

const TokenRoute = ({ match }) => {
  const token = match.params.token

  const { loading, error, data, refetch } = useQuery(
    validateTokenPasswordQuery,
    {
      variables: {
        token: match.params.token,
        password: sessionStorage.getItem(`share-token-pw-${token}`),
      },
    }
  )

  if (error) {
    if (error.message == 'GraphQL error: share not found') {
      return (
        <MessageContainer>
          <h1>Share not found</h1>
          <p>Maybe the share has expired or has been deleted.</p>
        </MessageContainer>
      )
    }

    return error.message
  }

  if (data && data.shareTokenValidatePassword == false) {
    return (
      <ProtectedTokenEnterPassword
        match={match}
        refetchWithPassword={password => {
          sessionStorage.setItem(`share-token-pw-${token}`, password)
          refetch({ variables: { password: password } })
        }}
        loading={loading}
      />
    )
  }

  if (loading) return 'Loading...'

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
      <Route path="/">Route not found</Route>
    </Switch>
  )
}

SharePage.propTypes = {
  ...RouterProps,
}

export default SharePage
