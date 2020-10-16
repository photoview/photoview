import React, { useCallback, useRef, useState } from 'react'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import PhotoGallery from '../../components/photoGallery/PhotoGallery'
import AlbumTitle from '../../components/AlbumTitle'
import { authToken } from '../../authentication'
import PropTypes from 'prop-types'
import AlbumFilter from '../../components/AlbumFilter'

const photoQuery = gql`
  query allGalleryGroups(
    $onlyWithFavorites: Boolean
    $mediaOrderBy: String
    $mediaOrderDirection: OrderDirection
  ) {
    myAlbums(
      filter: { order_by: "title", order_direction: ASC, limit: 100 }
      onlyWithFavorites: $onlyWithFavorites
    ) {
      title
      id
      media(
        filter: {
          order_by: $mediaOrderBy
          order_direction: $mediaOrderDirection
          limit: 12
        }
        onlyFavorites: $onlyWithFavorites
      ) {
        id
        title
        type
        thumbnail {
          url
          width
          height
        }
        highRes {
          url
          width
          height
        }
        videoWeb {
          url
        }
        favorite
      }
    }
  }
`

const GalleryGroups = ({ subPage }) => {
  const [activeIndex, setActiveIndex] = useState({ album: -1, media: -1 })
  const [presenting, setPresenting] = useState(false)
  const [onlyWithFavorites, setOnlyWithFavorites] = useState(
    subPage === 'favorites'
  )
  const [ordering, setOrdering] = useState({
    orderBy: 'date_shot',
    orderDirection: 'ASC',
  })

  const setOrderingCallback = useCallback(
    ordering => {
      setOrdering(prevState => {
        return {
          ...prevState,
          ...ordering,
        }
      })
    },
    [setOrdering]
  )

  const refetchNeeded = useRef({ all: false, favorites: false })

  const { loading, error, data, refetch } = useQuery(photoQuery, {
    variables: {
      onlyWithFavorites: onlyWithFavorites,
      mediaOrderBy: ordering.orderBy,
      mediaOrderDirection: ordering.orderDirection,
    },
  })

  const nextImage = useCallback(() => {
    setActiveIndex(index => {
      const albumMediaCount = data.myAlbums[index.album].media.length

      if (index.media + 1 < albumMediaCount) {
        return {
          ...index,
          media: index.media + 1,
        }
      } else {
        return index
      }
    })
  }, [data])

  const previousImage = useCallback(() => {
    setActiveIndex(index =>
      index.media > 0 ? { ...index, media: index.media - 1 } : index
    )
  })

  const setOnlyFavorites = useCallback(
    onlyWithFavorites => {
      history.replaceState(
        {},
        '',
        '/photos' + (onlyWithFavorites ? '/favorites' : '')
      )

      if (
        (refetchNeeded.current.all && !onlyWithFavorites) ||
        (refetchNeeded.current.favorites && onlyWithFavorites)
      ) {
        refetch({ onlyWithFavorites: onlyWithFavorites }).then(() => {
          if (onlyWithFavorites) {
            refetchNeeded.current.favorites = false
          } else {
            refetchNeeded.current.all = false
          }
          setOnlyWithFavorites(onlyWithFavorites)
        })
      } else {
        setOnlyWithFavorites(onlyWithFavorites)
      }
    },
    [setOnlyWithFavorites]
  )

  if (error) return error
  let galleryGroups = []

  if (!loading && data.myAlbums && authToken()) {
    galleryGroups = data.myAlbums.map((album, index) => (
      <div key={album.id}>
        <AlbumTitle album={album} />
        <PhotoGallery
          onSelectImage={mediaIndex => {
            setActiveIndex({ album: index, media: mediaIndex })
          }}
          onFavorite={() => {
            refetchNeeded.current.all = true
            refetchNeeded.current.favorites = true
          }}
          activeIndex={activeIndex.album === index ? activeIndex.media : -1}
          presenting={presenting === index}
          setPresenting={presenting =>
            setPresenting(presenting ? index : false)
          }
          loading={loading}
          media={album.media}
          nextImage={nextImage}
          previousImage={previousImage}
        />
      </div>
    ))
  }

  return (
    <>
      <AlbumFilter
        setOnlyFavorites={setOnlyFavorites}
        setOrdering={setOrderingCallback}
      />
      {galleryGroups}
    </>
  )
}

GalleryGroups.propTypes = {
  subPage: PropTypes.string,
}

export default GalleryGroups
