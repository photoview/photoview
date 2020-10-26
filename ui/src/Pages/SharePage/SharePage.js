import PropTypes from 'prop-types'
import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Route, Switch } from 'react-router-dom'
import RouterProps from 'react-router-prop-types'
import { Form, Header, Icon, Input, Message } from 'semantic-ui-react'
import styled from 'styled-components'
import { getSharePassword, saveSharePassword } from '../../authentication'
import AlbumSharePage from './AlbumSharePage'
import MediaSharePage from './MediaSharePage'

export const SHARE_TOKEN_QUERY = gql`
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
    type
    thumbnail {
      url
      width
      height
    }
    downloads {
      title
      mediaUrl {
        url
        width
        height
        fileSize
      }
    }
    highRes {
      url
    }
    videoWeb {
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

export const VALIDATE_TOKEN_PASSWORD_QUERY = gql`
  query ShareTokenValidatePassword($token: String!, $password: String) {
    shareTokenValidatePassword(token: $token, password: $password)
  }
`

const AuthorizedTokenRoute = ({ match }) => {
  const token = match.params.token

  const { loading, error, data } = useQuery(SHARE_TOKEN_QUERY, {
    variables: {
      token,
      password: getSharePassword(token),
    },
  })

  if (error) return error.message
  if (loading) return 'Loading...'

  if (data.shareToken.album) {
    return <AlbumSharePage album={data.shareToken.album} match={match} />
  }

  if (data.shareToken.media) {
    return <MediaSharePage media={data.shareToken.media} />
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
  refetchWithPassword: PropTypes.func.isRequired,
  loading: PropTypes.bool,
}

const TokenRoute = ({ match }) => {
  const token = match.params.token

  const { loading, error, data, refetch } = useQuery(
    VALIDATE_TOKEN_PASSWORD_QUERY,
    {
      variables: {
        token: match.params.token,
        password: getSharePassword(match.params.token),
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
          saveSharePassword(token, password)
          refetch({ variables: { password } })
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

const SharePage = ({ match }) => (
  <Switch>
    <Route path={`${match.url}/:token`}>
      {({ match }) => {
        return <TokenRoute match={match} />
      }}
    </Route>
    <Route path="/">Route not found</Route>
  </Switch>
)

SharePage.propTypes = {
  ...RouterProps,
}

export default SharePage
