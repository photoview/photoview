import React, { useState } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Link } from 'react-router-dom'
import { ProtectedImage } from '../photoGallery/ProtectedMedia'

const AlbumBoxLink = styled(Link)`
  width: 240px;
  height: 240px;
  display: inline-block;
  text-align: center;
  color: #222;
`

const ImageWrapper = styled.div`
  width: 240px;
  height: 220px;
  padding: 0 10px;
  position: relative;
`

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

  ${({ overlap, loaded }) =>
    overlap &&
    `
    position: absolute;
    top: 0;
    left: 10px;
    opacity: ${loaded ? 0 : 1};

    transition: opacity 200ms;
  `}
`

const AlbumBoxImage = ({ src, ...props }) => {
  const [loaded, setLoaded] = useState(false)

  if (src) {
    return (
      <ImageWrapper>
        <Image {...props} onLoad={loaded => setLoaded(loaded)} src={src} />
        <Placeholder overlap loaded={loaded ? 1 : 0} />
      </ImageWrapper>
    )
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

  let thumbnail = album.thumbnail && album.thumbnail.thumbnail.url

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
