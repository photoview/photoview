import React from 'react'
import styled from 'styled-components'
import { Link } from 'react-router-dom'
import ProtectedImage from '../../components/photoGallery/ProtectedImage'
import { Icon } from 'semantic-ui-react'

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

export const AlbumBox = ({ album, ...props }) => {
  if (!album) {
    return (
      <AlbumBoxLink {...props} to="#">
        <AlbumBoxImage />
      </AlbumBoxLink>
    )
  }

  return (
    <AlbumBoxLink {...props} to={`/album/${album.id}`}>
      <AlbumBoxImage
        src={
          album.photos[0] &&
          album.photos[0].thumbnail &&
          album.photos[0].thumbnail.url
        }
      />
      <p>{album.title}</p>
    </AlbumBoxLink>
  )
}
