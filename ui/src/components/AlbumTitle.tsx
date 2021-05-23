import React, { useEffect, useContext } from 'react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { SidebarContext } from './sidebar/Sidebar'
import AlbumSidebar from './sidebar/AlbumSidebar'
import { useLazyQuery, gql } from '@apollo/client'
import { authToken } from '../helpers/authentication'
import { albumPathQuery } from './__generated__/albumPathQuery'

const BreadcrumbList = styled.ol`
  & li::after {
    content: '';
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='5px' height='6px' viewBox='0 0 5 6'%3E%3Cpolyline fill='none' stroke='%23979797' points='0.74 0.167710644 3.57228936 3 0.74 5.83228936' /%3E%3C/svg%3E");
    width: 5px;
    height: 6px;
    display: inline-block;
    margin: 6px;
    vertical-align: middle;
  }
`

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
  const [fetchPath, { data: pathData }] =
    useLazyQuery<albumPathQuery>(ALBUM_PATH_QUERY)
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
      <li key={x.id} className="inline-block hover:underline">
        <Link to={`/album/${x.id}`}>{x.title}</Link>
      </li>
    ))

  if (!disableLink) {
    title = <Link to={`/album/${album.id}`}>{title}</Link>
  }

  return (
    <div className="flex">
      <div>
        <nav aria-label="Album breadcrumb">
          <BreadcrumbList className="">{breadcrumbSections}</BreadcrumbList>
        </nav>
        <h1 className="text-2xl">{title}</h1>
      </div>
      {authToken() && (
        <button
          onClick={() => {
            updateSidebar(<AlbumSidebar albumId={album.id} />)
          }}
        >
          More
        </button>
      )}
    </div>
  )
}

export default AlbumTitle
