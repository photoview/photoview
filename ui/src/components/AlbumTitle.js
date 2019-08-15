import React from 'react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { Icon } from 'semantic-ui-react'
import { SidebarConsumer } from './sidebar/Sidebar'

const Header = styled.h1`
  color: black;
  margin: 24px 0 12px 0;
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

  return <StyledIcon name="edit outline" size="small" {...props} />
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
          <SettingsIcon
            onClick={() => {
              updateSidebar(<div>Title stuff {album.title}</div>)
            }}
          />
        </Header>
      )}
    </SidebarConsumer>
  )
}

export default AlbumTitle
