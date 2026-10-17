import { gql, useQuery } from '@apollo/client'
import React, { useContext } from 'react'
import { Helmet } from 'react-helmet'
import AlbumTree from '../albumTree/AlbumTree'
import { AlbumTreeSearchProvider } from '../albumTree/AlbumTreeSearchContext'
import Header from '../header/Header'
import { Authorized } from '../routes/AuthorizedRoute'
import { Sidebar, SidebarContext } from '../sidebar/Sidebar'
import MainMenu from './MainMenu'
import { authToken } from '../../helpers/authentication'
import { layoutAlbumTreePreferenceQuery } from './__generated__/layoutAlbumTreePreferenceQuery'

export const ADMIN_QUERY = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

export const ALBUM_TREE_PREFERENCE_QUERY = gql`
  query layoutAlbumTreePreferenceQuery {
    myUserPreferences {
      id
      showAlbumTree
    }
  }
`

type LayoutProps = {
  children: React.ReactNode
  title: string
}

const Layout = ({ children, title, ...otherProps }: LayoutProps) => {
  const { pinned, content: sidebarContent } = useContext(SidebarContext)

  const token = authToken()

  const albumTreePreferenceQuery = useQuery<layoutAlbumTreePreferenceQuery>(
    ALBUM_TREE_PREFERENCE_QUERY,
    { skip: !token }
  )

  // The tree is opt-in: an account that never touched the setting gets the
  // gallery it had before. Waiting for the preference to actually arrive
  // matters for the other direction - defaulting to on while it loads would
  // mount the tree, fire its root query, and unmount it again as soon as a
  // stored "off" arrived, costing a query and a flash.
  const showAlbumTree =
    !!token &&
    albumTreePreferenceQuery.data != null &&
    (albumTreePreferenceQuery.data.myUserPreferences.showAlbumTree ?? false)

  return (
    <>
      <Helmet>
        <title>{title ? `${title} - Photoview` : `Photoview`}</title>
      </Helmet>
      <AlbumTreeSearchProvider>
        <div className="relative" {...otherProps} data-testid="Layout">
          <Header />
          <div className="">
            <Authorized>
              <MainMenu />
            </Authorized>
            {showAlbumTree && (
              <div className="hidden lg:block fixed lg:top-[84px] bottom-0 left-[292px] w-[260px] border-r border-gray-200 dark:border-dark-border bg-white dark:bg-dark-bg z-20">
                <AlbumTree />
              </div>
            )}
            <div
              className={`mx-3 my-3 lg:mt-5 lg:mr-8 ${
                showAlbumTree ? 'lg:ml-[576px]' : 'lg:ml-[292px]'
              } ${pinned && sidebarContent ? 'lg:pr-[420px]' : ''}`}
              id="layout-content"
            >
              {children}
            </div>
          </div>
          <Sidebar />
        </div>
      </AlbumTreeSearchProvider>
    </>
  )
}

export default Layout
