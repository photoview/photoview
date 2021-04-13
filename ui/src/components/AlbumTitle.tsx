import React, { useEffect, useContext } from 'react'
import PropTypes from 'prop-types'
import { Breadcrumb, IconProps } from 'semantic-ui-react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { Icon } from 'semantic-ui-react'
import { SidebarContext } from './sidebar/Sidebar'
import AlbumSidebar from './sidebar/AlbumSidebar'
import { useLazyQuery, gql } from '@apollo/client'
import { authToken } from '../helpers/authentication'
import { albumPathQuery } from './__generated__/albumPathQuery'

const Header = styled.h1`
  margin: 24px 0 8px 0 !important;

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

const SettingsIcon = (props: IconProps) => {
  return <StyledIcon name="settings" size="small" {...props} />
}

const ALBUM_PATH_QUERY = gql`
  query albumPathQuery($id: ID!) {
    album(id: $id) {
      id
      path {
        id
        title
      }
    }
  }
`

type AlbumTitleProps = {
  album?: {
    id: string
    title: string
  }
  disableLink: boolean
}

const AlbumTitle = ({ album, disableLink = false }: AlbumTitleProps) => {
  const [fetchPath, { data: pathData }] = useLazyQuery<albumPathQuery>(
    ALBUM_PATH_QUERY
  )
  const { updateSidebar } = useContext(SidebarContext)

  useEffect(() => {
    if (!album) return

    if (authToken() && disableLink == true) {
      fetchPath({
        variables: {
          id: album.id,
        },
      })
    }
  }, [album])

  if (!album) return <div style={{ height: 36 }}></div>

  let title = <span>{album.title}</span>

  const path = pathData?.album.path || []

  const breadcrumbSections = path
    .slice()
    .reverse()
    .map(x => (
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
    <>
      <Header>
        <Breadcrumb>{breadcrumbSections}</Breadcrumb>
        {title}
        {authToken() && (
          <SettingsIcon
            onClick={() => {
              updateSidebar(<AlbumSidebar albumId={album.id} />)
            }}
          />
        )}
      </Header>
    </>
  )
}

AlbumTitle.propTypes = {
  album: PropTypes.object,
  disableLink: PropTypes.bool,
}

export default AlbumTitle
