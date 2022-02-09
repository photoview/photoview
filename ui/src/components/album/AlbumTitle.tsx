import React, { useEffect, useContext } from 'react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { SidebarContext } from '../sidebar/Sidebar'
import AlbumSidebar from '../sidebar/AlbumSidebar'
import { useLazyQuery, gql } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { albumPathQuery } from './__generated__/albumPathQuery'
import useDelay from '../../hooks/useDelay'

import { ReactComponent as GearIcon } from './icons/gear.svg'
import { tailwindClassNames } from '../../helpers/utils'
import { buttonStyles } from '../../primitives/form/Input'

export const BreadcrumbList = styled.ol<{ hideLastArrow?: boolean }>`
  &
    ${({ hideLastArrow }) =>
      hideLastArrow ? 'li:not(:last-child)::after' : 'li::after'} {
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

  const delay = useDelay(200, [album])

  if (!album) {
    return (
      <div
        className={`flex mb-6 flex-col h-14 transition-opacity animate-pulse ${
          delay ? 'opacity-100' : 'opacity-0'
        }`}
      >
        <div className="w-32 h-4 bg-gray-100 mb-2 mt-1"></div>
        <div className="w-72 h-6 bg-gray-100"></div>
      </div>
    )
  }

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
    <div className="flex mb-6 items-end h-14">
      <div className="min-w-0">
        <nav aria-label="Album breadcrumb">
          <BreadcrumbList>{breadcrumbSections}</BreadcrumbList>
        </nav>
        <h1 className="text-2xl truncate min-w-0">{title}</h1>
      </div>
      {authToken() && (
        <button
          title="Album options"
          aria-label="Album options"
          className={tailwindClassNames(buttonStyles({}), 'px-2 py-2 ml-2')}
          onClick={() => {
            updateSidebar(<AlbumSidebar albumId={album.id} />)
          }}
        >
          <GearIcon />
        </button>
      )}
    </div>
  )
}

export default AlbumTitle
