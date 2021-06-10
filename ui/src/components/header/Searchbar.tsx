import React, { useState, useRef, useEffect } from 'react'
import styled from 'styled-components'
import { useLazyQuery, gql } from '@apollo/client'
import { debounce, DebouncedFn } from '../../helpers/utils'
import { ProtectedImage } from '../photoGallery/ProtectedMedia'
import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  searchQuery,
  searchQuery_search_albums,
  searchQuery_search_media,
} from './__generated__/searchQuery'

// const SearchField = styled.input`
//   height: 100%;
//   width: 100%;
//   border: 1px solid #eee;
//   border-radius: 4px;
//   padding: 0 8px;
//   font-size: 1rem;
//   font-family: Lato, 'Helvetica Neue', Arial, Helvetica, sans-serif;

//   &:focus {
//     box-shadow: 0 0 4px #eee;
//     border-color: #3d82c6;
//   }
// `

// const Results = styled.div<{ show: boolean }>`
//   display: ${({ show }) => (show ? 'block' : 'none')};
//   position: absolute;
//   width: 100%;
//   min-height: 40px;
//   max-height: calc(100vh - 100px);
//   overflow-y: scroll;
//   padding: 28px 4px 32px;
//   background-color: white;
//   box-shadow: 0 0 4px #eee;
//   border: 1px solid #ccc;
//   border-radius: 4px;
//   top: 50%;
//   z-index: -1;

//   input:not(:focus) ~ & {
//     display: none;
//   }
// `

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
  const { t } = useTranslation()
  const [fetchSearches, fetchResult] = useLazyQuery<searchQuery>(SEARCH_QUERY)
  const [query, setQuery] = useState('')
  const [fetched, setFetched] = useState(false)

  type QueryFn = (query: string) => void

  const debouncedFetch = useRef<null | DebouncedFn<QueryFn>>(null)
  useEffect(() => {
    debouncedFetch.current = debounce<QueryFn>(query => {
      fetchSearches({ variables: { query } })
      setFetched(true)
    }, 250)

    return () => {
      debouncedFetch.current?.cancel()
    }
  }, [])

  const fetchEvent = (e: React.ChangeEvent<HTMLInputElement>) => {
    e.persist()

    setQuery(e.target.value)
    if (e.target.value.trim() != '' && debouncedFetch.current) {
      debouncedFetch.current(e.target.value.trim())
    } else {
      setFetched(false)
    }
  }

  let results = null
  if (query.trim().length > 0 && fetched) {
    results = (
      <SearchResults
        searchData={fetchResult.data}
        loading={fetchResult.loading}
      />
    )
  }

  return (
    <div className="w-full max-w-xs">
      <input
        className="w-full py-2 px-3 rounded-md bg-gray-50 focus:bg-white border border-gray-50 focus:border-blue-400 outline-none focus:ring-2 focus:ring-blue-400 focus:ring-opacity-50"
        type="search"
        placeholder={t('header.search.placeholder', 'Search')}
        onChange={fetchEvent}
      />
      {results}
    </div>
  )
}

const ResultTitle = styled.h1.attrs({
  className: 'uppercase text-gray-700 font-semibold mt-4 mb-2 mx-1',
})``

type SearchResultsProps = {
  searchData?: searchQuery
  loading: boolean
}

const SearchResults = ({ searchData, loading }: SearchResultsProps) => {
  const { t } = useTranslation()
  const query = searchData?.search.query || ''

  const media = searchData?.search.media || []
  const albums = searchData?.search.albums || []

  let message = null
  if (loading) message = t('header.search.loading', 'Loading results...')
  else if (searchData && media.length == 0 && albums.length == 0)
    message = t('header.search.no_results', 'No results found')

  if (message) message = <div className="mt-8 text-center">{message}</div>

  const albumElements = albums
    .slice(0, 5)
    .map(album => <AlbumRow key={album.id} query={query} album={album} />)

  const mediaElements = media
    .slice(0, 5)
    .map(media => <PhotoRow key={media.id} query={query} media={media} />)

  return (
    <div
      className="absolute bg-white left-0 right-0 top-[72px] overflow-y-auto h-[calc(100vh-152px)] border px-4"
      onMouseDown={e => {
        // Prevent input blur event
        e.preventDefault()
      }}
      // show={!!searchData}
    >
      {message}
      {albumElements.length > 0 && (
        <ResultTitle>
          {t('header.search.result_type.albums', 'Albums')}
        </ResultTitle>
      )}
      {albumElements}
      {mediaElements.length > 0 && (
        <ResultTitle>
          {t('header.search.result_type.media', 'Media')}
        </ResultTitle>
      )}
      {mediaElements}
    </div>
  )
}

const RowLink = styled(NavLink).attrs({
  className:
    'focus:bg-gray-100 hover:bg-gray-100 focus:ring-1 outline-none rounded p-1 mt-1 flex items-center',
})``

const RowTitle = styled.span.attrs({ className: 'flex-grow pl-2' })``

type PhotoRowArgs = {
  query: string
  media: searchQuery_search_media
}

const PhotoRow = ({ query, media }: PhotoRowArgs) => (
  <RowLink to={`/album/${media.album.id}`}>
    <ProtectedImage
      src={media?.thumbnail?.url}
      className="w-16 h-16 object-cover"
    />
    <RowTitle>{searchHighlighted(query, media.title)}</RowTitle>
  </RowLink>
)

type AlbumRowArgs = {
  query: string
  album: searchQuery_search_albums
}

const AlbumRow = ({ query, album }: AlbumRowArgs) => (
  <RowLink to={`/album/${album.id}`}>
    <ProtectedImage
      src={album?.thumbnail?.thumbnail?.url}
      className="w-16 h-16 rounded object-cover"
    />
    <RowTitle>{searchHighlighted(query, album.title)}</RowTitle>
  </RowLink>
)

const searchHighlighted = (query: string, text: string) => {
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
      <strong className="font-semibold">{middle}</strong>
      {end}
    </>
  )
}

export default SearchBar
