import React from 'react'
import { useQuery, gql } from '@apollo/client'
import { Route, Routes, useParams } from 'react-router-dom'
import styled from 'styled-components'
import {
  getSharePassword,
  saveSharePassword,
} from '../../helpers/authentication'
import AlbumSharePage from './AlbumSharePage'
import MediaSharePage from './MediaSharePage'
import { useTranslation } from 'react-i18next'
import PasswordProtectedShare from './PasswordProtectedShare'
import { isNil } from '../../helpers/utils'
import {
  SharePageToken,
  SharePageTokenVariables,
} from './__generated__/SharePageToken'
import {
  ShareTokenValidatePassword,
  ShareTokenValidatePasswordVariables,
} from './__generated__/ShareTokenValidatePassword'

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
          description
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

const tokenFromParams = () => {
  const { token } = useParams()
  if (isNil(token)) throw new Error('Expected `token` param to be defined')
  return token
}

const AuthorizedTokenRoute = () => {
  const { t } = useTranslation()

  const token = tokenFromParams()
  const password = getSharePassword(token)

  const { loading, error, data } = useQuery<
    SharePageToken,
    SharePageTokenVariables
  >(SHARE_TOKEN_QUERY, {
    variables: {
      token,
      password,
    },
  })

  if (!isNil(error)) return <div>{error.message}</div>
  if (loading) return <div>{t('general.loading.default', 'Loading...')}</div>

  if (data?.shareToken.album) {
    const SharedSubAlbumPage = () => {
      const { subAlbum } = useParams()
      if (isNil(subAlbum))
        throw new Error('Expected `subAlbum` param to be defined')

      return (
        <AlbumSharePage albumID={subAlbum} token={token} password={password} />
      )
    }

    return (
      <Routes>
        <Route path=":subAlbum" element={<SharedSubAlbumPage />} />
        <Route
          index
          element={
            <AlbumSharePage
              albumID={data.shareToken.album.id}
              token={token}
              password={password}
            />
          }
        />
      </Routes>
    )
  }

  if (data?.shareToken.media) {
    return <MediaSharePage media={data.shareToken.media} />
  }

  return <h1>{t('share_page.share_not_found', 'Share not found')}</h1>
}

export const MessageContainer = styled.div`
  max-width: 400px;
  margin: 100px auto 0;
`

export const TokenRoute = () => {
  const { t } = useTranslation()
  const token = tokenFromParams()

  const { loading, error, data, refetch } = useQuery<
    ShareTokenValidatePassword,
    ShareTokenValidatePasswordVariables
  >(VALIDATE_TOKEN_PASSWORD_QUERY, {
    notifyOnNetworkStatusChange: true,
    variables: {
      token: token,
      password: getSharePassword(token),
    },
  })

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

  return <AuthorizedTokenRoute />
}
