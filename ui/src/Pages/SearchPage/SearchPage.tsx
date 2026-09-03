import { gql, useQuery } from '@apollo/client'
import React, { useEffect, useReducer } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../components/layout/Layout'
import MediaGallery, {
  MEDIA_GALLERY_FRAGMENT,
} from '../../components/photoGallery/MediaGallery'
import {
  mediaGalleryReducer,
  urlPresentModeSetupHook,
} from '../../components/photoGallery/mediaGalleryReducer'
import useURLParameters from '../../hooks/useURLParameters'
import {
  searchPageQuery,
  searchPageQueryVariables,
  searchPageQuery_search_media,
} from './__generated__/searchPageQuery'

const SEARCH_PAGE_QUERY = gql`
  ${MEDIA_GALLERY_FRAGMENT}

  query searchPageQuery($query: String!) {
    search(query: $query, limitAlbums: 0, limitMedia: 0) {
      albums {
        id
        title
        thumbnail {
          id
          thumbnail {
            url
          }
        }
      }
      media {
        ...MediaGalleryFields
        album {
          id
          title
        }
      }
    }
  }
`

type MediaAlbumGroup = {
  id: string
  title: string
  media: searchPageQuery_search_media[]
}

const groupMediaByAlbum = (
  media: searchPageQuery_search_media[]
): MediaAlbumGroup[] => {
  const groups = new Map<string, MediaAlbumGroup>()

  for (const item of media) {
    const existing = groups.get(item.album.id)
    if (existing) {
      existing.media.push(item)
    } else {
      groups.set(item.album.id, {
        id: item.album.id,
        title: item.album.title,
        media: [item],
      })
    }
  }

  return [...groups.values()]
}

const SearchAlbumMediaGroup = ({ id, title, media }: MediaAlbumGroup) => {
  const [mediaState, dispatchMedia] = useReducer(mediaGalleryReducer, {
    presenting: false,
    activeIndex: -1,
    media,
  })

  useEffect(() => {
    dispatchMedia({ type: 'replaceMedia', media })
  }, [media])

  urlPresentModeSetupHook({
    dispatchMedia,
    openPresentMode: event => {
      dispatchMedia({
        type: 'openPresentMode',
        activeIndex: event.state.activeIndex,
      })
    },
    groupId: id,
  })

  return (
    <div className="mb-8">
      <Link to={`/album/${id}`} className="hover:underline">
        <h2 className="text-xl mb-2">{title}</h2>
      </Link>
      <MediaGallery
        loading={false}
        mediaState={mediaState}
        dispatchMedia={dispatchMedia}
        groupId={id}
      />
    </div>
  )
}

const SearchPage = () => {
  const { t } = useTranslation()
  const urlParams = useURLParameters()
  const query = urlParams.getParam('q') ?? ''

  const { data, loading, error } = useQuery<
    searchPageQuery,
    searchPageQueryVariables
  >(SEARCH_PAGE_QUERY, {
    variables: { query },
    skip: query.trim() === '',
  })

  const albums = data?.search.albums ?? []
  const mediaGroups = groupMediaByAlbum(data?.search.media ?? [])

  const noResults =
    !loading && data != null && albums.length === 0 && mediaGroups.length === 0

  return (
    <Layout
      title={t('search_page.title', 'Search results for "{{query}}"', {
        query,
      })}
    >
      <h1 className="text-2xl mb-6 truncate">
        {t('search_page.heading', 'Search results for "{{query}}"', {
          query,
        })}
      </h1>

      {error && <div>{t('search_page.error', 'Error loading results')}</div>}

      {loading && (
        <div className="text-gray-400">
          {t('general.loading.default', 'Loading...')}
        </div>
      )}

      {noResults && (
        <div className="text-gray-400">
          {t('search_page.no_results', 'No results found')}
        </div>
      )}

      {albums.length > 0 && (
        <>
          <h2 className="text-xl mb-2">
            {t('search_page.albums_heading', 'Albums')}
          </h2>
          <AlbumBoxes albums={albums} />
        </>
      )}

      {mediaGroups.map(group => (
        <SearchAlbumMediaGroup key={group.id} {...group} />
      ))}
    </Layout>
  )
}

export default SearchPage
