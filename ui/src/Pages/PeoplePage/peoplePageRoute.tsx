/**
 * Light lazy-loadable wrapper for the people module
 */

import React from 'react'
import { Route } from 'react-router-dom'

const PeoplePage = React.lazy(() =>
  import('./PeoplePage').then(x => ({
    default: x.PeoplePage,
  }))
)

const PersonPage = React.lazy(() =>
  import('./PeoplePage').then(x => ({
    default: x.PersonPage,
  }))
)

const peoplePageRoute = () => (
  <>
    <Route path=":person" element={<PersonPage />} />
    <Route index element={<PeoplePage />} />
  </>
)

export default peoplePageRoute
