import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { AlbumBox } from './AlbumBox'

const Container = styled.div`
  margin: 20px -10px;
  position: relative;
`

const AlbumGallery = ({ loading, error, albums }) => {
  if (error) return <div>Error {error.message}</div>

  let albumElements = []

  if (albums) {
    albumElements = albums.map(album => (
      <AlbumBox key={album.id} album={album} />
    ))
  } else {
    for (let i = 0; i < 8; i++) {
      albumElements.push(<AlbumBox key={i} />)
    }
  }

  return (
    <Container>
      <Loader active={loading}>Loading albums</Loader>
      {albumElements}
    </Container>
  )
}

AlbumGallery.propTypes = {
  loading: PropTypes.bool.isRequired,
  error: PropTypes.object,
  albums: PropTypes.array,
}

export default AlbumGallery
