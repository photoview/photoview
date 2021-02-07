import React from 'react'
import PropTypes from 'prop-types'
import TimelineGroupAlbum from './TimelineGroupAlbum'
import styled from 'styled-components'

const dateFormatter = new Intl.DateTimeFormat(navigator.language, {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
})

const GroupDateWrapper = styled.div`
  margin: 12px 12px;
`

const DateTitle = styled.h1`
  font-size: 1.5rem;
  margin: 0 0 -12px;
`

const GroupAlbumWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
  margin: 0 -8px;
`

const TimelineGroupDate = ({
  date,
  groups,
  onSelectDateGroup,
  activeIndex,
  setPresenting,
  onFavorite,
}) => {
  const albumGroupElms = groups.map((group, i) => (
    <TimelineGroupAlbum
      key={`${group.date}_${group.album.id}`}
      group={group}
      onSelectMedia={mediaIndex => {
        onSelectDateGroup({
          media: mediaIndex,
          albumGroup: i,
        })
      }}
      activeIndex={activeIndex.albumGroup == i ? activeIndex.media : -1}
      setPresenting={setPresenting}
      onFavorite={onFavorite}
    />
  ))

  const formattedDate = dateFormatter.format(new Date(date))

  return (
    <GroupDateWrapper>
      <DateTitle>{formattedDate}</DateTitle>
      <GroupAlbumWrapper>{albumGroupElms}</GroupAlbumWrapper>
    </GroupDateWrapper>
  )
}

TimelineGroupDate.propTypes = {
  date: PropTypes.string.isRequired,
  groups: PropTypes.array.isRequired,
  onSelectDateGroup: PropTypes.func.isRequired,
  activeIndex: PropTypes.object.isRequired,
  setPresenting: PropTypes.func.isRequired,
  onFavorite: PropTypes.func,
}

export default TimelineGroupDate
