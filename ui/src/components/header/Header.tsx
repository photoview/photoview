import React from 'react'
import SearchBar from './Searchbar'

import logoPath from '../../assets/photoview-logo.svg'
import { authToken } from '../../helpers/authentication'

const Header = () => (
  <div className="bg-white flex items-center py-3 px-4 shadow-separator">
    <h1 className="mr-4 flex-shrink-0">
      <img className="h-12" src={logoPath} alt="logo" />
      <span className="hidden">Photoview</span>
    </h1>
    {authToken() ? <SearchBar /> : null}
  </div>
)

export default Header
