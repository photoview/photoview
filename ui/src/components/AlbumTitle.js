import React, { useEffect } from 'react'
import PropTypes from 'prop-types'
import { Breadcrumb } from 'semantic-ui-react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { Icon } from 'semantic-ui-react'
import { SidebarConsumer } from './sidebar/Sidebar'
import AlbumSidebar from './sidebar/AlbumSidebar'
import gql from 'graphql-tag'
import { useLazyQuery } from '@apollo/react-hooks'

const Header = styled.h1`
  margin: 0 0 12px 0 !important;

  & a {
    color: black;

    &:hover {
      text-decoration: underline;
    }
  }
`

const StyledIcon = styled(Icon)`
  margin-left: 8px !important;
  display: inline-block;
  color: #888;
  cursor: pointer;

  &:hover {
    color: #1e70bf;
  }
`

const SettingsIcon = (props) => {
  return <StyledIcon name="settings" size="small" {...props} />
}

const ALBUM_PATH_QUERY = gql`
  query albumPathQuery($id: Int!) {
    album(id: $id) {
      id
      path {
        id
        title
      }
    }
  }
`

const AlbumTitle = ({ album, disableLink = false }) => {
  if (!album) return <div style={{ height: 36 }}></div>

  let title = <span>{album.title}</span>

  const [fetchPath, { data: pathData }] = useLazyQuery(ALBUM_PATH_QUERY)

  useEffect(() => {
    if (localStorage.getItem('token') && disableLink == true) {
      fetchPath({
        variables: {
          id: album.id,
        },
      })
    }
  }, [album])

  let path = []
  if (pathData) {
    path = pathData.album.path
  }

  const breadcrumbSections = path
    .slice()
    .reverse()
    .map((x) => (
      <span key={x.id}>
        <Breadcrumb.Section as={Link} to={`/album/${x.id}`}>
          {x.title}
        </Breadcrumb.Section>
        <Breadcrumb.Divider icon="right angle" />
      </span>
    ))

  if (!disableLink) {
    title = <Link to={`/album/${album.id}`}>{title}</Link>
  }

  return (
    <SidebarConsumer>
      {({ updateSidebar }) => (
        <Header>
          <Breadcrumb>{breadcrumbSections}</Breadcrumb>
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
