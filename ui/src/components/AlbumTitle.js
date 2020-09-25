import React, { useEffect, useContext } from 'react'
import PropTypes from 'prop-types'
import { Breadcrumb, Checkbox } from 'semantic-ui-react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { Icon } from 'semantic-ui-react'
import { SidebarContext } from './sidebar/Sidebar'
import AlbumSidebar from './sidebar/AlbumSidebar'
import gql from 'graphql-tag'
import { useLazyQuery } from '@apollo/react-hooks'
import { authToken } from '../authentication'

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

const FavoritesCheckbox = styled(Checkbox)`
  float: right;
  padding-left: 10px;
  margin-top: 0.5rem;
`

const SettingsIcon = props => {
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

const AlbumTitle = ({
  album,
  disableLink = false,
  showFavoritesToggle,
  setOnlyFavorites,
  onlyFavorites = false,
}) => {
  const [fetchPath, { data: pathData }] = useLazyQuery(ALBUM_PATH_QUERY)
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

  let path = []
  if (pathData) {
    path = pathData.album.path
  }

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
      {authToken() && showFavoritesToggle && (
        <FavoritesCheckbox
          toggle
          label="Show only the favorites"
          checked={onlyFavorites}
          onClick={e => e.stopPropagation()}
          onChange={setOnlyFavorites}
        />
      )}
    </Header>
  )
}

AlbumTitle.propTypes = {
  album: PropTypes.object,
  disableLink: PropTypes.bool,
  showFavoritesToggle: PropTypes.bool,
  setOnlyFavorites: PropTypes.func,
  onlyFavorites: PropTypes.bool,
}

export default AlbumTitle
