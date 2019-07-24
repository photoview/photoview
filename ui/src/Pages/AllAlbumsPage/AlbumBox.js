import React from 'react'
import styled from 'styled-components'
import { Link } from 'react-router-dom'

const AlbumBoxLink = styled(Link)`
  width: 240px;
  height: 240px;
  display: inline-block;
  text-align: center;
  color: #222;
`

const AlbumBoxImage = styled.div`
  width: 220px;
  height: 220px;
  margin: auto;
  border-radius: 4%;
  background-image: url('${props => props.image}');
  background-color: #eee;
  background-size: cover;
  background-position: center;
`

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
        image={
          album.photos[0] &&
          album.photos[0].thumbnail &&
          album.photos[0].thumbnail.url
        }
      />
      <p>{album.title}</p>
    </AlbumBoxLink>
  )
}
