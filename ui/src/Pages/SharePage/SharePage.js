import PropTypes from 'prop-types'
import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Route, Switch } from 'react-router-dom'
import RouterProps from 'react-router-prop-types'
import { Form, Header, Icon, Input, Message } from 'semantic-ui-react'
import styled from 'styled-components'
import {
  getSharePassword,
  saveSharePassword,
} from '../../helpers/authentication'
import AlbumSharePage from './AlbumSharePage'
import MediaSharePage from './MediaSharePage'
import { useTranslation } from 'react-i18next'

export const SHARE_TOKEN_QUERY = gql`
  query SharePageToken($token: String!, $password: String) {
    shareToken(credentials: { token: $token, password: $password }) {
      token
      album {
        id
      }
      media {
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
          width
          height
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
    }
  }
`

export const VALIDATE_TOKEN_PASSWORD_QUERY = gql`
  query ShareTokenValidatePassword($token: String!, $password: String) {
    shareTokenValidatePassword(
      credentials: { token: $token, password: $password }
    )
  }
`

const AuthorizedTokenRoute = ({ match }) => {
  const { t } = useTranslation()

  const token = match.params.token
  const password = getSharePassword(token)

  const { loading, error, data } = useQuery(SHARE_TOKEN_QUERY, {
    variables: {
      token,
      password,
    },
  })

  if (error) return error.message
  if (loading) return 'Loading...'

  if (data.shareToken.album) {
    const SharedSubAlbumPage = ({ match }) => {
      return (
        <AlbumSharePage
          albumID={match.params.subAlbum}
          token={token}
          password={password}
        />
      )
    }

    SharedSubAlbumPage.propTypes = {
      match: PropTypes.any,
    }

    return (
      <Switch>
        <Route
          exact
          path={`${match.path}/:subAlbum`}
          component={SharedSubAlbumPage}
        />
        <Route exact path={match.path}>
          <AlbumSharePage
            albumID={data.shareToken.album.id}
            token={token}
            password={password}
          />
        </Route>
      </Switch>
    )
  }

  if (data.shareToken.media) {
    return <MediaSharePage media={data.shareToken.media} />
  }

  return <h1>{t('share_page.share_not_found', 'Share not found')}</h1>
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
  const { t } = useTranslation()

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
        <Message.Content>
          {t('share_page.wrong_password', 'Wrong password, please try again.')}
        </Message.Content>
      </Message>
    )
  }

  return (
    <MessageContainer>
      <Header as="h1" style={{ fontWeight: 400 }}>
        {t('share_page.protected_share.title', 'Protected share')}
      </Header>
      <p>
        {t(
          'share_page.protected_share.description',
          'This share is protected with a password.'
        )}
      </p>
      <Form>
        <Form.Field>
          <label>{t('login_page.field.password', 'Password')}</label>
          <Input
            loading={loading}
            disabled={loading}
            onKeyUp={event => event.key == 'Enter' && onSubmit()}
            onChange={e => setPasswordValue(e.target.value)}
            placeholder={t('login_page.field.password', 'Password')}
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
  const { t } = useTranslation()

  const token = match.params.token

  const { loading, error, data, refetch } = useQuery(
    VALIDATE_TOKEN_PASSWORD_QUERY,
    {
      notifyOnNetworkStatusChange: true,
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
          <h1>{t('share_page.share_not_found', 'Share not found')}</h1>
          <p>
            {t(
              'share_page.share_not_found_description',
              'Maybe the share has expired or has been deleted.'
            )}
          </p>
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
          refetch({ token, password })
        }}
        loading={loading}
      />
    )
  }

  if (loading) return t('general.loading.default', 'Loading...')

  return <AuthorizedTokenRoute match={match} />
}

TokenRoute.propTypes = {
  match: PropTypes.object.isRequired,
}

const SharePage = ({ match }) => {
  const { t } = useTranslation()

  return (
    <Switch>
      <Route path={`${match.url}/:token`}>
        {({ match }) => {
          return <TokenRoute match={match} />
        }}
      </Route>
      <Route path="/">{t('routes.page_not_found', 'Page not found')}</Route>
    </Switch>
  )
}

SharePage.propTypes = {
  ...RouterProps,
}

export default SharePage
