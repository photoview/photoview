import React from 'react'
import PropTypes from 'prop-types'
import { useQuery, gql } from '@apollo/client'
import SidebarShare from './Sharing'
import { useTranslation } from 'react-i18next'

const albumQuery = gql`
  query getAlbumSidebar($id: ID!) {
    album(id: $id) {
      id
      title
    }
  }
`

const AlbumSidebar = ({ albumId }) => {
  const { t } = useTranslation()
  const { loading, error, data } = useQuery(albumQuery, {
    variables: { id: albumId },
  })

  if (loading) return <div>{t('general.loading.default', 'Loading...')}</div>
  if (error) return <div>{error.message}</div>

  return (
    <div>
      <p>{t('sidebar.album.title', 'Album options')}</p>
      <div>
        <h1>{data.album.title}</h1>
        <SidebarShare album={data.album} />
      </div>
    </div>
  )
}

AlbumSidebar.propTypes = {
  albumId: PropTypes.string.isRequired,
}

export default AlbumSidebar
