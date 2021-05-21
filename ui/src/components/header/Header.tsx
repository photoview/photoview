import React from 'react'
import SearchBar from './Searchbar'

import logoPath from '../../assets/photoview-logo.svg'
import { authToken } from '../../helpers/authentication'

const Header = () => (
  <div className="bg-white flex items-center justify-between py-3 px-4 lg:px-8 lg:pt-4 shadow-separator lg:shadow-none">
    <h1 className="mr-4 lg:mr-8 flex-shrink-0 flex items-center">
      <img className="h-12 lg:h-10" src={logoPath} alt="logo" />
      <span className="hidden lg:block ml-2 text-2xl font-light">
        Photoview
      </span>
    </h1>
    {authToken() ? <SearchBar /> : null}
  </div>
)

export default Header
