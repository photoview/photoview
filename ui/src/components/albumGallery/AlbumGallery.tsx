import React, { useState, useEffect } from 'react'
import AlbumTitle from '../AlbumTitle'
import PhotoGallery from '../photoGallery/PhotoGallery'
import AlbumBoxes from './AlbumBoxes'
import AlbumFilter from '../AlbumFilter'
import { albumQuery_album } from '../../Pages/AlbumPage/__generated__/albumQuery'

type AlbumGalleryProps = {
  album: albumQuery_album
  loading?: boolean
  customAlbumLink?(albumID: string): string
  showFilter?: boolean
  setOnlyFavorites?(favorites: boolean): void
  setOrdering?(ordering: { orderBy: string }): void
  ordering?: { orderBy: string }
  onlyFavorites?: boolean
  onFavorite?(): void
}

const AlbumGallery = React.forwardRef(
  (
    {
      album,
      loading = false,
      customAlbumLink,
      showFilter = false,
      setOnlyFavorites,
      setOrdering,
      ordering,
      onlyFavorites = false,
      onFavorite,
    }: AlbumGalleryProps,
    ref: React.ForwardedRef<HTMLDivElement>
  ) => {
    type ImageStateType = {
      activeImage: number
      presenting: boolean
    }

    const [imageState, setImageState] = useState<ImageStateType>({
      activeImage: -1,
      presenting: false,
    })

    const setPresenting = (presenting: boolean) =>
      setImageState(state => ({ ...state, presenting }))

    const setPresentingWithHistory = (presenting: boolean) => {
      setPresenting(presenting)
      if (presenting) {
        history.pushState({ imageState }, '')
      } else {
        history.back()
      }
    }

    const updateHistory = (imageState: ImageStateType) => {
      history.replaceState({ imageState }, '')
      return imageState
    }

    const setActiveImage = (activeImage: number) => {
      setImageState(state => updateHistory({ ...state, activeImage }))
    }

    const nextImage = () => {
      setActiveImage((imageState.activeImage + 1) % album.media.length)
    }

    const previousImage = () => {
      if (imageState.activeImage <= 0) {
        setActiveImage(album.media.length - 1)
      } else {
        setActiveImage(imageState.activeImage - 1)
      }
    }

    useEffect(() => {
      const updateImageState = (event: PopStateEvent) => {
        setImageState(event.state.imageState)
      }

      window.addEventListener('popstate', updateImageState)

      return () => {
        window.removeEventListener('popstate', updateImageState)
      }
    }, [imageState])

    useEffect(() => {
      setActiveImage(-1)
    }, [album])

    let subAlbumElement = null

    if (album) {
      if (album.subAlbums.length > 0) {
        subAlbumElement = (
          <AlbumBoxes
            loading={loading}
            albums={album.subAlbums}
            getCustomLink={customAlbumLink}
          />
        )
      }
    } else {
      subAlbumElement = <AlbumBoxes loading={loading} />
    }

    return (
      <div ref={ref}>
        <AlbumTitle album={album} disableLink />
        {showFilter && (
          <AlbumFilter
            onlyFavorites={onlyFavorites}
            setOnlyFavorites={setOnlyFavorites}
            setOrdering={setOrdering}
            ordering={ordering}
          />
        )}
        {subAlbumElement}
        {
          <h2
            style={{
              opacity: loading ? 0 : 1,
              display: album && album.subAlbums.length > 0 ? 'block' : 'none',
            }}
          >
            Images
          </h2>
        }
        <PhotoGallery
          loading={loading}
          media={album && album.media}
          activeIndex={imageState.activeImage}
          presenting={imageState.presenting}
          onSelectImage={index => {
            setActiveImage(index)
          }}
          onFavorite={onFavorite}
          setPresenting={setPresentingWithHistory}
          nextImage={nextImage}
          previousImage={previousImage}
        />
      </div>
    )
  }
)

export default AlbumGallery
