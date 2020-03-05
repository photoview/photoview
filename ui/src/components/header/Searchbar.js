import React, { useState } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { useLazyQuery } from '@apollo/react-hooks'
import gql from 'graphql-tag'
import debounce from 'lodash/debounce'
import ProtectedImage from '../photoGallery/ProtectedImage'
import { NavLink } from 'react-router-dom'

const Container = styled.div`
  height: 60px;
  width: 350px;
  margin: 0 12px;
  padding: 12px 0;
  position: relative;
`

const SearchField = styled.input`
  height: 100%;
  width: 100%;
  border: 1px solid #eee;
  border-radius: 4px;
  padding: 0 8px;
  font-size: 16px;
  font-family: Lato, 'Helvetica Neue', Arial, Helvetica, sans-serif;

  &:focus {
    box-shadow: 0 0 4px #eee;
    border-color: #3d82c6;
  }
`

const Results = styled.div`
  display: ${({ show }) => (show ? 'block' : 'none')};
  position: absolute;
  width: 100%;
  min-height: 40px;
  max-height: calc(100vh - 100px);
  overflow-y: scroll;
  padding: 28px 4px 32px;
  background-color: white;
  box-shadow: 0 0 4px #eee;
  border: 1px solid #ccc;
  border-radius: 4px;
  top: 50%;
  z-index: -1;

  ${SearchField}:not(:focus) ~ & {
    display: none;
  }
`

const SEARCH_QUERY = gql`
  query searchQuery($query: String!) {
    search(query: $query) {
      albums {
        id
        title
        thumbnail {
          thumbnail {
            url
          }
        }
      }
      photos {
        id
        title
        thumbnail {
          url
        }
      }
    }
  }
`

const SearchBar = () => {
  const [fetchSearches, fetchResult] = useLazyQuery(SEARCH_QUERY)
  const [query, setQuery] = useState('')
  const [fetched, setFetched] = useState(false)

  let debouncedFetch = null
  const fetchEvent = e => {
    e.persist()

    if (!debouncedFetch) {
      debouncedFetch = debounce(() => {
        console.log('searching', e.target.value.trim())
        fetchSearches({ variables: { query: e.target.value.trim() } })
        setFetched(true)
      }, 250)
    }

    setQuery(e.target.value)
    if (e.target.value.trim() != '') {
      debouncedFetch()
    } else {
      setFetched(false)
    }
  }

  let results = null
  if (query.trim().length > 0 && fetched) {
    results = <SearchResults result={fetchResult} />
  }

  return (
    <Container>
      <SearchField type="search" placeholder="Search" onChange={fetchEvent} />
      {results}
    </Container>
  )
}

const ResultTitle = styled.h1`
  font-size: 20px;
  margin: 12px 0 4px;
`

const SearchResults = ({ result }) => {
  const { data, loading } = result

  const photos = (data && data.search.photos) || []
  const albums = (data && data.search.albums) || []

  let message = null
  if (loading) message = 'Loading results...'
  else if (data && photos.length == 0 && albums.length == 0)
    message = 'No results found'

  const albumElements = albums.map(album => (
    <AlbumRow key={album.id} {...album} />
  ))

  const photoElements = photos.map(photo => (
    <PhotoRow key={photo.id} {...photo} />
  ))

  return (
    <Results
      onMouseDown={e => {
        // Prevent input blur event
        e.preventDefault()
      }}
      show={data}
    >
      {message}
      <ResultTitle>Albums</ResultTitle>
      {albumElements}
      <ResultTitle>Photos</ResultTitle>
      {photoElements}
    </Results>
  )
}

const RowLink = styled(NavLink)`
  display: flex;
  align-items: center;
  border-bottom: 1px solid #eee;
  color: black;
`

const PhotoSearchThumbnail = styled(ProtectedImage)`
  width: 50px;
  height: 50px;
  object-fit: contain;
`

const AlbumSearchThumbnail = styled(ProtectedImage)`
  width: 50px;
  height: 50px;
  margin: 4px 0;
  border-radius: 4px;
  border: 1px solid #888;
  object-fit: cover;
`

const RowTitle = styled.span`
  flex-grow: 1;
  padding-left: 8px;
`

const PhotoRow = photo => (
  <RowLink to="/">
    <PhotoSearchThumbnail src={photo.thumbnail.url} />
    <RowTitle>{photo.title}</RowTitle>
  </RowLink>
)

const AlbumRow = album => (
  <RowLink to={`/album/${album.id}`}>
    <AlbumSearchThumbnail src={album.thumbnail.thumbnail.url} />
    <RowTitle>{album.title}</RowTitle>
  </RowLink>
)

SearchResults.propTypes = {
  result: PropTypes.object,
}

export default SearchBar
