import React from 'react'
import PropTypes from 'prop-types'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { Icon } from 'semantic-ui-react'
import { SidebarConsumer } from './sidebar/Sidebar'
import AlbumSidebar from './sidebar/AlbumSidebar'

const Header = styled.h1`
  margin: 0 0 12px 0 !important;

  & a {
    color: black;

    &:hover {
      text-decoration: underline;
    }
  }
`

const SettingsIcon = props => {
  const StyledIcon = styled(Icon)`
    margin-left: 8px !important;
    display: inline-block;
    color: #888;
    cursor: pointer;

    &:hover {
      color: #1e70bf;
    }
  `

  return <StyledIcon name="settings" size="small" {...props} />
}

const AlbumTitle = ({ album, disableLink = false }) => {
  if (!album) return null

  let title = <span>{album.title}</span>

  if (!disableLink) {
    title = <Link to={`/album/${album.id}`}>{title}</Link>
  }

  return (
    <SidebarConsumer>
      {({ updateSidebar }) => (
        <Header>
          {title}
          {localStorage.getItem('token') && (
            <SettingsIcon
              onClick={() => {
                updateSidebar(<AlbumSidebar albumId={album.id} />)
              }}
            />
          )}
        </Header>
      )}
    </SidebarConsumer>
  )
}

AlbumTitle.propTypes = {
  album: PropTypes.object,
  disableLink: PropTypes.bool,
}

export default AlbumTitle
