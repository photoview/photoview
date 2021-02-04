import React from 'react'
import { useQuery, gql } from '@apollo/client'
import TimelineGroupDate from './TimelineGroupDate'
import styled from 'styled-components'

const MY_TIMELINE_QUERY = gql`
  query myTimeline {
    myTimeline {
      album {
        id
        title
      }
      media {
        id
        thumbnail {
          url
          width
          height
        }
      }
      mediaTotal
      date
    }
  }
`

const GalleryWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
`

const TimelineGallery = () => {
  const { data, error } = useQuery(MY_TIMELINE_QUERY)

  if (error) {
    return error
  }

  let timelineGroups = null
  if (data?.myTimeline) {
    const dateGroupedAlbums = data.myTimeline.reduce((acc, val) => {
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

    timelineGroups = dateGroupedAlbums.map(({ date, groups }) => (
      <TimelineGroupDate key={date} date={date} groups={groups} />
    ))
  }

  return <GalleryWrapper>{timelineGroups}</GalleryWrapper>
}

export default TimelineGallery
