import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Link } from 'react-router-dom'
import ProtectedImage from '../photoGallery/ProtectedImage'

const AlbumBoxLink = styled(Link)`
  width: 240px;
  height: 240px;
  display: inline-block;
  text-align: center;
  color: #222;
`

const AlbumBoxImage = ({ src, ...props }) => {
  const Image = styled(ProtectedImage)`
    width: 220px;
    height: 220px;
    margin: auto;
    border-radius: 4%;
    object-fit: cover;
    object-position: center;
  `

  const Placeholder = styled.div`
    width: 220px;
    height: 220px;
    border-radius: 4%;
    margin: auto;
    background: linear-gradient(#f7f7f7 0%, #eee 100%);
  `

  if (src) {
    return <Image {...props} src={src} />
  }

  return <Placeholder />
}

AlbumBoxImage.propTypes = {
  src: PropTypes.string,
}

export const AlbumBox = ({ album, customLink, ...props }) => {
  if (!album) {
    return (
      <AlbumBoxLink {...props} to="#">
        <AlbumBoxImage />
      </AlbumBoxLink>
    )
  }

  let thumbnail =
    album.photos[0] &&
    album.photos[0].thumbnail &&
    album.photos[0].thumbnail.url

  thumbnail =
    thumbnail ||
    (album.subAlbums &&
      album.subAlbums[0] &&
      album.subAlbums[0].photos[0] &&
      album.subAlbums[0].photos[0].thumbnail.url)

  return (
    <AlbumBoxLink {...props} to={customLink || `/album/${album.id}`}>
      <AlbumBoxImage src={thumbnail} />
      <p>{album.title}</p>
    </AlbumBoxLink>
  )
}

AlbumBox.propTypes = {
  album: PropTypes.object,
  customLink: PropTypes.string,
}
