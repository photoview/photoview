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

const TimelineGroupDate = ({ date, groups }) => {
  const groupElms = groups.map(group => (
    <TimelineGroupAlbum key={`${group.date}_${group.album.id}`} group={group} />
  ))

  const formattedDate = dateFormatter.format(new Date(date))

  return (
    <GroupDateWrapper>
      <DateTitle>{formattedDate}</DateTitle>
      <div>{groupElms}</div>
    </GroupDateWrapper>
  )
}

TimelineGroupDate.propTypes = {
  date: PropTypes.string.isRequired,
  groups: PropTypes.array.isRequired,
}

export default TimelineGroupDate
