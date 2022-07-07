import React, { useContext } from 'react'
import SearchBar from './Searchbar'

import { authToken } from '../../helpers/authentication'
import { SidebarContext } from '../sidebar/Sidebar'
import classNames from 'classnames'

const Header = () => {
  const { pinned } = useContext(SidebarContext)

  return (
    <div
      className={classNames(
        'sticky top-0 z-10 bg-white dark:bg-dark-bg flex items-center justify-between py-3 px-4 lg:px-8 lg:pt-4 shadow-separator lg:shadow-none',
        { 'mr-[404px]': pinned }
      )}
    >
      <h1 className="mr-4 lg:mr-8 flex-shrink-0 flex items-center">
        <img
          className="h-12 lg:h-10"
          src={import.meta.env.BASE_URL + 'photoview-logo.svg'}
          alt="logo"
        />
        <span className="hidden lg:block ml-2 text-2xl font-light">
          Photoview
        </span>
      </h1>
      {authToken() ? <SearchBar /> : null}
    </div>
  )
}

export default Header
