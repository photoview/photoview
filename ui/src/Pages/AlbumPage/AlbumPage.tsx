import { useCallback } from 'react'
import { useQuery, gql } from '@apollo/client'
import AlbumGallery, {
  ALBUM_GALLERY_FRAGMENT,
} from '../../components/albumGallery/AlbumGallery'
import Layout from '../../components/layout/Layout'
import useURLParameters from '../../hooks/useURLParameters'
import useScrollPagination from '../../hooks/useScrollPagination'
import PaginateLoader from '../../components/PaginateLoader'
import { useTranslation } from 'react-i18next'
import { albumQuery, albumQueryVariables } from './__generated__/albumQuery'
import useOrderingParams from '../../hooks/useOrderingParams'
import { useParams } from 'react-router-dom'
import { isNil } from '../../helpers/utils'

const ALBUM_QUERY = gql`
  ${ALBUM_GALLERY_FRAGMENT}

  query albumQuery(
    $id: ID!
    $onlyFavorites: Boolean
    $mediaOrderBy: String
    $orderDirection: OrderDirection
    $limit: Int
    $offset: Int
  ) {
    album(id: $id) {
      ...AlbumGalleryFields
    }
  }
`

let refetchNeededAll = false
let refetchNeededFavorites = false

/**
 * Displays an album page with a media gallery, supporting filtering, ordering, and infinite scroll pagination.
 *
 * Fetches album data by ID from the URL, allows toggling between all media and favorites, and manages pagination and ordering of media items. Handles loading and error states, and updates the page title dynamically based on album data.
 *
 * @throws {Error} If the album ID parameter is missing from the URL.
 */
function AlbumPage() {
  const { id: albumId } = useParams()
  if (isNil(albumId))
    throw new Error('Expected parameter `id` to be defined for AlbumPage')

  const { t } = useTranslation()

  const urlParams = useURLParameters()
  const orderParams = useOrderingParams(urlParams)

  const onlyFavorites = urlParams.getParam('favorites') == '1' ? true : false
  const setOnlyFavorites = useCallback((favorites: boolean) =>
    urlParams.setParam('favorites', favorites ? '1' : '0'),
    [urlParams]
  )

  const { loading, error, data, refetch, fetchMore } = useQuery<
    albumQuery,
    albumQueryVariables
  >(ALBUM_QUERY, {
    variables: {
      id: albumId,
      onlyFavorites,
      mediaOrderBy: orderParams.orderBy,
      orderDirection: orderParams.orderDirection,
      offset: 0,
      limit: 200,
    },
  })

  const { containerElem, finished: finishedLoadingMore } =
    useScrollPagination<albumQuery>({
      loading,
      fetchMore,
      data,
      getItems: data => data.album.media,
    })

  const toggleFavorites = useCallback(
    (onlyFavorites: boolean) => {
      if (
        (refetchNeededAll && !onlyFavorites) ||
        (refetchNeededFavorites && onlyFavorites)
      ) {
        refetch({ id: albumId, onlyFavorites: onlyFavorites }).then(() => {
          if (onlyFavorites) {
            refetchNeededFavorites = false
          } else {
            refetchNeededAll = false
          }
          setOnlyFavorites(onlyFavorites)
        })
      } else {
        setOnlyFavorites(onlyFavorites)
      }
    }, [setOnlyFavorites, refetch, albumId]
  )

  if (error) return <div>Error</div>

  return (
    <Layout
      title={
        !data ? t('title.loading_album', 'Loading album') :
          !data.album ? t('title.not_found', 'Not found') :
            data.album.title
      }
    >
      <AlbumGallery
        ref={containerElem}
        album={data && data.album}
        loading={loading}
        setOnlyFavorites={toggleFavorites}
        onlyFavorites={onlyFavorites}
        onFavorite={() => (refetchNeededAll = refetchNeededFavorites = true)}
        showFilter
        setOrdering={orderParams.setOrdering}
        ordering={orderParams}
      />
      <PaginateLoader
        active={!finishedLoadingMore && !loading}
        text={t('general.loading.paginate.media', 'Loading more media')}
      />
    </Layout>
  )
}

export default AlbumPage
