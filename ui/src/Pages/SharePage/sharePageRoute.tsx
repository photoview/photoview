/**
 * Light lazy-loadable wrapper for share page module
 */

import React from 'react'
import type { TFunction } from 'react-i18next'
import { Route } from 'react-router-dom'
import { NotFoundPage } from '../../components/routes/Routes'

const TokenRoute = React.lazy(() =>
  import('./SharePage').then(x => ({
    default: x.TokenRoute,
  }))
)

const sharePageRoute = ({ t }: { t: TFunction }) => {
  return (
    <>
      <Route path={':token'} element={<TokenRoute />} />
      <Route index element={<NotFoundPage t={t} />} />
    </>
  )
}

export default sharePageRoute
