import React, { useRef, useEffect, useReducer, useContext } from 'react'
import PropTypes from 'prop-types'
import { useQuery, gql } from '@apollo/client'
import TimelineGroupDate from './TimelineGroupDate'
import PresentView from '../photoGallery/presentView/PresentView'
import useURLParameters from '../../hooks/useURLParameters'
import { FavoritesCheckbox } from '../album/AlbumFilter'
import useScrollPagination from '../../hooks/useScrollPagination'
import PaginateLoader from '../PaginateLoader'
import { useTranslation } from 'react-i18next'
import {
  myTimeline,
  myTimelineVariables,
  myTimeline_myTimeline,
} from './__generated__/myTimeline'
import {
  getActiveTimelineImage as getActiveTimelineMedia,
  timelineGalleryReducer,
} from './timelineGalleryReducer'
import MediaSidebar from '../sidebar/MediaSidebar'
import { SidebarContext } from '../sidebar/Sidebar'
import { urlPresentModeSetupHook } from '../photoGallery/photoGalleryReducer'

const MY_TIMELINE_QUERY = gql`
  query myTimeline($onlyFavorites: Boolean, $limit: Int, $offset: Int) {
    myTimeline(
      onlyFavorites: $onlyFavorites
      paginate: { limit: $limit, offset: $offset }
    ) {
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

export type TimelineActiveIndex = {
  albumGroup: number
  media: number
}

export type TimelineGroup = {
  date: string
  groups: myTimeline_myTimeline[]
}

const TimelineGallery = () => {
  const { t } = useTranslation()
  const { updateSidebar } = useContext(SidebarContext)

  const { getParam, setParam } = useURLParameters()

  const onlyFavorites = getParam('favorites') == '1' ? true : false
  const setOnlyFavorites = (favorites: boolean) =>
    setParam('favorites', favorites ? '1' : '0')

  const favoritesNeedsRefresh = useRef(false)

  const [mediaState, dispatchMedia] = useReducer(timelineGalleryReducer, {
    presenting: false,
    timelineGroups: [],
    activeIndex: {
      media: -1,
      album: -1,
      date: -1,
    },
  })

  const { data, error, loading, refetch, fetchMore } = useQuery<
    myTimeline,
    myTimelineVariables
  >(MY_TIMELINE_QUERY, {
    variables: {
      onlyFavorites,
      offset: 0,
      limit: 50,
    },
  })

  const { containerElem, finished: finishedLoadingMore } =
    useScrollPagination<myTimeline>({
      loading,
      fetchMore,
      data,
      getItems: data => data.myTimeline,
    })

  useEffect(() => {
    dispatchMedia({
      type: 'replaceTimelineGroups',
      timeline: data?.myTimeline || [],
    })
  }, [data])

  useEffect(() => {
    const activeMedia = getActiveTimelineMedia({ mediaState })
    if (activeMedia) {
      updateSidebar(<MediaSidebar media={activeMedia} />)
    } else {
      updateSidebar(null)
    }
  }, [mediaState.activeIndex])

  urlPresentModeSetupHook({
    dispatchMedia,
    openPresentMode: event => {
      dispatchMedia({
        type: 'openPresentMode',
        activeIndex: event.state.activeIndex,
      })
    },
  })

  useEffect(() => {
    favoritesNeedsRefresh.current = false
    refetch({
      onlyFavorites: onlyFavorites,
    })
  }, [onlyFavorites])

  if (error) {
    return <div>{error.message}</div>
  }

  const timelineGroups = mediaState.timelineGroups.map((_, i) => (
    <TimelineGroupDate
      key={i}
      groupIndex={i}
      mediaState={mediaState}
      dispatchMedia={dispatchMedia}
    />
  ))

  return (
    <div className="overflow-x-hidden">
      <div className="mb-2">
        <FavoritesCheckbox
          onlyFavorites={onlyFavorites}
          setOnlyFavorites={setOnlyFavorites}
        />
      </div>
      <div className="-mx-3 flex flex-wrap" ref={containerElem}>
        {timelineGroups}
      </div>
      <PaginateLoader
        active={!finishedLoadingMore && !loading}
        text={t('general.loading.paginate.media', 'Loading more media')}
      />
      {mediaState.presenting && (
        <PresentView
          activeMedia={getActiveTimelineMedia({ mediaState })!}
          dispatchMedia={dispatchMedia}
        />
      )}
    </div>
  )
}

TimelineGallery.propTypes = {
  favorites: PropTypes.bool,
  setFavorites: PropTypes.func,
}

export default TimelineGallery
