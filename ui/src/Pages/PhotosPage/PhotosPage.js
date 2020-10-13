import React, { useCallback, useRef, useState } from 'react'
import Layout from '../../Layout'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import PhotoGallery from '../../components/photoGallery/PhotoGallery'
import AlbumTitle from '../../components/AlbumTitle'
import { authToken } from '../../authentication'
import PropTypes from 'prop-types'
import AlbumFilter from '../../components/AlbumFilter'

const photoQuery = gql`
  query allPhotosPage(
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

const PhotosPage = ({ match }) => {
  const [activeIndex, setActiveIndex] = useState({ album: -1, media: -1 })
  const [presenting, setPresenting] = useState(false)
  const [onlyWithFavorites, setOnlyWithFavorites] = useState(
    match.params.subPage === 'favorites'
  )

  const refetchNeeded = useRef({ all: false, favorites: false })

  const [orderBy, setOrderBy] = useState('date_shot')
  const [orderDirection, setOrderDirection] = useState('ASC')

  const { loading, error, data, refetch } = useQuery(photoQuery, {
    variables: {
      onlyWithFavorites: onlyWithFavorites,
      mediaOrderBy: orderBy,
      mediaOrderDirection: orderDirection,
    },
  })

  const setSorting = useCallback(
    (e, d) => {
      const [orderBy, orderDirection] = d.value.split('.')
      setOrderBy(orderBy)
      setOrderDirection(orderDirection)
      refetch({
        onlyWithFavorites: onlyWithFavorites,
        mediaOrderBy: orderBy,
        mediaOrderDirection: orderDirection,
      })
    },
    [refetch, setOrderBy, setOrderDirection]
  )

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

  const setOnlyFavorites = useCallback(() => {
    const updatedWithFavorites = !onlyWithFavorites

    history.replaceState(
      {},
      '',
      '/photos' + (updatedWithFavorites ? '/favorites' : '')
    )

    if (
      (refetchNeeded.current.all && !updatedWithFavorites) ||
      (refetchNeeded.current.favorites && updatedWithFavorites)
    ) {
      refetch({ onlyWithFavorites: updatedWithFavorites }).then(() => {
        if (updatedWithFavorites) {
          refetchNeeded.current.favorites = false
        } else {
          refetchNeeded.current.all = false
        }
        setOnlyWithFavorites(updatedWithFavorites)
      })
    } else {
      setOnlyWithFavorites(updatedWithFavorites)
    }
  }, [onlyWithFavorites])

  if (error) return error
  if (loading) return null

  let galleryGroups = []

  if (data.myAlbums && authToken()) {
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
      <Layout title="Photos">
        <AlbumFilter
          onlyFavorites={onlyWithFavorites}
          setOnlyFavorites={setOnlyFavorites}
          setSorting={setSorting}
        />
        {galleryGroups}
      </Layout>
    </>
  )
}

PhotosPage.propTypes = {
  match: PropTypes.shape({
    params: PropTypes.shape({
      subPage: PropTypes.string,
    }),
  }),
}

export default PhotosPage
