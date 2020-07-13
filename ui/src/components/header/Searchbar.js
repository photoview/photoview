import React, { useState, useRef, useEffect } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { useLazyQuery } from '@apollo/react-hooks'
import gql from 'graphql-tag'
import debounce from '../../debounce'
import { ProtectedImage } from '../photoGallery/ProtectedMedia'
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
  font-size: 1rem;
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
      query
      albums {
        id
        title
        thumbnail {
          thumbnail {
            url
          }
        }
      }
      media {
        id
        title
        thumbnail {
          url
        }
        album {
          id
        }
      }
    }
  }
`

const SearchBar = () => {
  const [fetchSearches, fetchResult] = useLazyQuery(SEARCH_QUERY)
  const [query, setQuery] = useState('')
  const [fetched, setFetched] = useState(false)

  let debouncedFetch = useRef(null)
  useEffect(() => {
    debouncedFetch.current = debounce(query => {
      console.log('searching', query)
      fetchSearches({ variables: { query } })
      setFetched(true)
    }, 250)

    return () => {
      debouncedFetch.current.cancel()
    }
  }, [])

  const fetchEvent = e => {
    e.persist()

    setQuery(e.target.value)
    if (e.target.value.trim() != '') {
      debouncedFetch.current(e.target.value.trim())
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
  font-size: 1.25rem;
  margin: 12px 0 0.25rem;
`

const SearchResults = ({ result }) => {
  const { data, loading } = result
  const query = data && data.search.query

  const media = (data && data.search.media) || []
  const albums = (data && data.search.albums) || []

  let message = null
  if (loading) message = 'Loading results...'
  else if (data && media.length == 0 && albums.length == 0)
    message = 'No results found'

  const albumElements = albums.map(album => (
    <AlbumRow key={album.id} query={query} album={album} />
  ))

  const mediaElements = media.map(media => (
    <PhotoRow key={media.id} query={query} media={media} />
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
      {albumElements.length > 0 && <ResultTitle>Albums</ResultTitle>}
      {albumElements}
      {mediaElements.length > 0 && <ResultTitle>Photos</ResultTitle>}
      {mediaElements}
    </Results>
  )
}

SearchResults.propTypes = {
  result: PropTypes.object,
}

const RowLink = styled(NavLink)`
  display: flex;
  align-items: center;
  color: black;
`

const PhotoSearchThumbnail = styled(ProtectedImage)`
  width: 50px;
  height: 50px;
  margin: 2px 0;
  object-fit: contain;
`

const AlbumSearchThumbnail = styled(ProtectedImage)`
  width: 50px;
  height: 50px;
  margin: 4px 0;
  border-radius: 4px;
  /* border: 1px solid #888; */
  object-fit: cover;
`

const RowTitle = styled.span`
  flex-grow: 1;
  padding-left: 8px;
`

const PhotoRow = ({ query, media }) => (
  <RowLink to={`/album/${media.album.id}`}>
    <PhotoSearchThumbnail src={media.thumbnail.url} />
    <RowTitle>{searchHighlighted(query, media.title)}</RowTitle>
  </RowLink>
)

PhotoRow.propTypes = {
  query: PropTypes.string.isRequired,
  media: PropTypes.object.isRequired,
}

const AlbumRow = ({ query, album }) => (
  <RowLink to={`/album/${album.id}`}>
    <AlbumSearchThumbnail src={album.thumbnail.thumbnail.url} />
    <RowTitle>{searchHighlighted(query, album.title)}</RowTitle>
  </RowLink>
)

AlbumRow.propTypes = {
  query: PropTypes.string.isRequired,
  album: PropTypes.object.isRequired,
}

const searchHighlighted = (query, text) => {
  const i = text.toLowerCase().indexOf(query.toLowerCase())

  if (i == -1) {
    return text
  }

  const start = text.substring(0, i)
  const middle = text.substring(i, i + query.length)
  const end = text.substring(i + query.length)

  return (
    <>
      {start}
      <b>{middle}</b>
      {end}
    </>
  )
}

export default SearchBar
