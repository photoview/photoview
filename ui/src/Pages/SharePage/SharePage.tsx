import PropTypes from 'prop-types'
import React from 'react'
import { useQuery, gql } from '@apollo/client'
import { match as MatchType, Route, Switch } from 'react-router-dom'
import styled from 'styled-components'
import {
  getSharePassword,
  saveSharePassword,
} from '../../helpers/authentication'
import AlbumSharePage from './AlbumSharePage'
import MediaSharePage from './MediaSharePage'
import { useTranslation } from 'react-i18next'
import PasswordProtectedShare from './PasswordProtectedShare'

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
          width
          height
        }
        exif {
          id
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
          coordinates {
            longitude
            latitude
          }
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

const AuthorizedTokenRoute = ({ match }: MatchProps<TokenRouteMatch>) => {
  const { t } = useTranslation()

  const token = match.params.token
  const password = getSharePassword(token)

  const { loading, error, data } = useQuery(SHARE_TOKEN_QUERY, {
    variables: {
      token,
      password,
    },
  })

  if (error) return <div>{error.message}</div>
  if (loading) return <div>{t('general.loading.default', 'Loading...')}</div>

  if (data.shareToken.album) {
    const SharedSubAlbumPage = ({ match }: MatchProps<SubalbumRouteMatch>) => {
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

export const MessageContainer = styled.div`
  max-width: 400px;
  margin: 100px auto 0;
`

interface TokenRouteMatch {
  token: string
}

interface SubalbumRouteMatch extends TokenRouteMatch {
  subAlbum: string
}

interface MatchProps<Route> {
  match: MatchType<Route>
}

const TokenRoute = ({ match }: MatchProps<TokenRouteMatch>) => {
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

    return <div>{error.message}</div>
  }

  if (data && data.shareTokenValidatePassword == false) {
    return (
      <PasswordProtectedShare
        refetchWithPassword={password => {
          saveSharePassword(token, password)
          refetch({ token, password })
        }}
        loading={loading}
      />
    )
  }

  if (loading) return <div>{t('general.loading.default', 'Loading...')}</div>

  return <AuthorizedTokenRoute match={match} />
}

const SharePage = ({ match }: { match: MatchType }) => {
  const { t } = useTranslation()

  return (
    <Switch>
      <Route path={`${match.url}/:token`}>
        {({ match }: { match: MatchType<TokenRouteMatch> }) => {
          return <TokenRoute match={match} />
        }}
      </Route>
      <Route path="/">{t('routes.page_not_found', 'Page not found')}</Route>
    </Switch>
  )
}

export default SharePage
