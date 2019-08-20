import React, { useState } from 'react'
import PropTypes from 'prop-types'
import RouterPropTypes from 'react-router-prop-types'
import { Switch, Route } from 'react-router-dom'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'

const AlbumSharePage = ({ album, match }) => {
  const SubAlbumRoute = subProps => {
    const subAlbumId = subProps.match.params.subAlbum
    const subAlbum = album.subAlbums.find(x => x.id == subAlbumId)

    if (!subAlbum) {
      return <div>Subalbum was not found</div>
    }

    return <AlbumSharePage album={subAlbum} {...subProps} />
  }

  SubAlbumRoute.propTypes = {
    ...RouterPropTypes,
  }

  const customAlbumLink = albumId => {
    return `${match.url}/${albumId}`
  }
  return (
    <Switch>
      <Route path={`${match.url}/:subAlbum`} component={SubAlbumRoute} />
      <Route path="/">
        <AlbumGallery album={album} customAlbumLink={customAlbumLink} />
      </Route>
    </Switch>
  )
}

AlbumSharePage.propTypes = {
  album: PropTypes.object.isRequired,
  match: RouterPropTypes.match,
}

export default AlbumSharePage
