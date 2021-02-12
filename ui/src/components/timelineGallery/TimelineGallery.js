import React, { useState, useCallback, useRef, useEffect } from 'react'
import PropTypes from 'prop-types'
import { useQuery, gql } from '@apollo/client'
import TimelineGroupDate from './TimelineGroupDate'
import styled from 'styled-components'
import PresentView from '../photoGallery/presentView/PresentView'
import { Loader } from 'semantic-ui-react'
import useURLParameters from '../../hooks/useURLParameters'
import { FavoritesCheckbox } from '../AlbumFilter'
import useScrollPagination from '../../hooks/useScrollPagination'

const MY_TIMELINE_QUERY = gql`
  query myTimeline($onlyFavorites: Boolean, $limit: Int, $offset: Int) {
    myTimeline(onlyFavorites: $onlyFavorites, limit: $limit, offset: $offset) {
      album {
        id
        title
      }
      media {
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
      mediaTotal
      date
    }
  }
`

const GalleryWrapper = styled.div`
  margin: -12px;
  display: flex;
  flex-wrap: wrap;
`

const TimelineGallery = () => {
  const [activeIndex, setActiveIndex] = useState({
    dateGroup: -1,
    albumGroup: -1,
    media: -1,
  })
  const [presenting, setPresenting] = useState(false)

  const { getParam, setParam } = useURLParameters()

  const onlyFavorites = getParam('favorites') == '1' ? true : false
  const setOnlyFavorites = favorites => setParam('favorites', favorites ? 1 : 0)
  const favoritesNeedsRefresh = useRef(false)

  const nextMedia = useCallback(() => {
    setActiveIndex(activeIndex => {
      const albumGroups = dateGroupedAlbums[activeIndex.dateGroup].groups
      const albumMedia = albumGroups[activeIndex.albumGroup].media

      if (activeIndex.media < albumMedia.length - 1) {
        return {
          ...activeIndex,
          media: activeIndex.media + 1,
        }
      }

      if (activeIndex.albumGroup < albumGroups.length - 1) {
        return {
          ...activeIndex,
          albumGroup: activeIndex.albumGroup + 1,
          media: 0,
        }
      }

      if (activeIndex.dateGroup < dateGroupedAlbums.length - 1) {
        return {
          dateGroup: activeIndex.dateGroup + 1,
          albumGroup: 0,
          media: 0,
        }
      }

      // reached the end
      return activeIndex
    })
  }, [activeIndex])

  const previousMedia = useCallback(() => {
    setActiveIndex(activeIndex => {
      if (activeIndex.media > 0) {
        return {
          ...activeIndex,
          media: activeIndex.media - 1,
        }
      }

      if (activeIndex.albumGroup > 0) {
        return {
          ...activeIndex,
          albumGroup: activeIndex.albumGroup - 1,
          media: 0,
        }
      }

      if (activeIndex.dateGroup > 0) {
        return {
          dateGroup: activeIndex.dateGroup - 1,
          albumGroup: 0,
          media: 0,
        }
      }

      // reached the start
      return activeIndex
    })
  }, [activeIndex])

  const { data, error, loading, refetch, fetchMore } = useQuery(
    MY_TIMELINE_QUERY,
    {
      variables: {
        onlyFavorites,
        offset: 0,
        limit: 50,
      },
    }
  )

  const { containerElem, finished: finishedLoadingMore } = useScrollPagination({
    loading,
    fetchMore,
    data,
    getItems: data => data.myTimeline,
  })

  useEffect(() => {
    if (favoritesNeedsRefresh) {
      favoritesNeedsRefresh.current = false
      refetch({
        onlyFavorites: onlyFavorites,
      })
    }
  }, [onlyFavorites])

  if (error) {
    return error
  }

  let timelineGroups = null
  let dateGroupedAlbums = []
  if (data?.myTimeline) {
    dateGroupedAlbums = data.myTimeline.reduce((acc, val) => {
      if (acc.length == 0 || acc[acc.length - 1].date != val.date) {
        acc.push({
          date: val.date,
          groups: [val],
        })
      } else {
        acc[acc.length - 1].groups.push(val)
      }

      return acc
    }, [])

    timelineGroups = dateGroupedAlbums.map(({ date, groups }, i) => (
      <TimelineGroupDate
        key={date}
        date={date}
        groups={groups}
        activeIndex={
          activeIndex.dateGroup == i
            ? activeIndex
            : { albumGroup: -1, media: -1 }
        }
        setPresenting={setPresenting}
        onSelectDateGroup={({ media, albumGroup }) => {
          setActiveIndex({
            media,
            albumGroup,
            dateGroup: i,
          })
        }}
        onFavorite={() => {
          favoritesNeedsRefresh.current = true
        }}
      />
    ))
  }

  return (
    <>
      <Loader active={loading}>Loading timeline</Loader>
      <FavoritesCheckbox
        onlyFavorites={onlyFavorites}
        setOnlyFavorites={setOnlyFavorites}
      />
      <GalleryWrapper ref={containerElem}>{timelineGroups}</GalleryWrapper>
      <Loader
        style={{ margin: '42px 0 24px 0' }}
        active={!finishedLoadingMore && !loading}
        inline="centered"
      >
        Loading more media
      </Loader>
      {presenting && (
        <PresentView
          media={
            dateGroupedAlbums &&
            dateGroupedAlbums[activeIndex.dateGroup].groups[
              activeIndex.albumGroup
            ].media[activeIndex.media]
          }
          nextImage={nextMedia}
          previousImage={previousMedia}
          setPresenting={setPresenting}
        />
      )}
    </>
  )
}

TimelineGallery.propTypes = {
  favorites: PropTypes.bool,
  setFavorites: PropTypes.func,
}

export default TimelineGallery
