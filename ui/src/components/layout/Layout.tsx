import { gql } from '@apollo/client'
import React, { useContext } from 'react'
import { Helmet } from 'react-helmet'
import Header from '../header/Header'
import { Authorized } from '../routes/AuthorizedRoute'
import { Sidebar, SidebarContext } from '../sidebar/Sidebar'
import MainMenu from './MainMenu'
import { useModal } from '../sidebar/ReDetection/ReDetectFacesContext'
import ReDetectFacesModal from '../sidebar/ReDetection/ReDetectFacesModal'

export const ADMIN_QUERY = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

type LayoutProps = {
  children: React.ReactNode
  title: string
}

const Layout = ({ children, title, ...otherProps }: LayoutProps) => {
  const { pinned, content: sidebarContent } = useContext(SidebarContext)
  const { isModalOpen, closeModal } = useModal()

  return (
    <>
      <Helmet>
        <title>{title ? `${title} - Photoview` : `Photoview`}</title>
      </Helmet>
      <div className="relative" {...otherProps} data-testid="Layout">
        <Header />
        <div className="">
          <Authorized>
            <MainMenu />
          </Authorized>
          <div
            className={`mx-3 my-3 lg:mt-5 lg:mr-8 lg:ml-[292px] ${
              pinned && sidebarContent ? 'lg:pr-[420px]' : ''
            }`}
            id="layout-content"
          >
            {children}
          </div>
        </div>
        <Sidebar />
        <ReDetectFacesModal
          isOpen={isModalOpen}
          onClose={closeModal}
          onConfirm={closeModal}
        />
      </div>
    </>
  )
}

export default Layout
