import React from 'react'
import {
  Navigate,
  useNavigate,
  NavigateFunction,
  Outlet,
  useRoutes,
} from 'react-router-dom'

import Layout from '../layout/Layout'
import { authToken, clearTokenCookie } from '../../helpers/authentication'
import { TFunction, useTranslation } from 'react-i18next'
import Loader from '../../primitives/Loader'
import AuthorizedRoute from './AuthorizedRoute'

const AlbumsPage = React.lazy(
  () => import('../../Pages/AllAlbumsPage/AlbumsPage')
)
const AlbumPage = React.lazy(() => import('../../Pages/AlbumPage/AlbumPage'))
const TimelinePage = React.lazy(
  () => import('../../Pages/TimelinePage/TimelinePage')
)
const PlacesPage = React.lazy(() => import('../../Pages/PlacesPage/PlacesPage'))

const LoginPage = React.lazy(() => import('../../Pages/LoginPage/LoginPage'))
const InitialSetupPage = React.lazy(
  () => import('../../Pages/LoginPage/InitialSetupPage')
)

const SharePageTokenRoute = React.lazy(() =>
  import('../../Pages/SharePage/SharePage').then(x => ({
    default: x.TokenRoute,
  }))
)

const SettingsPage = React.lazy(
  () => import('../../Pages/SettingsPage/SettingsPage')
)

const PeoplePage = React.lazy(() =>
  import('../../Pages/PeoplePage/PeoplePage').then(x => ({
    default: x.PeoplePage,
  }))
)

const PersonPage = React.lazy(() =>
  import('../../Pages/PeoplePage/PeoplePage').then(x => ({
    default: x.PersonPage,
  }))
)

const Routes = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const authorized = (element: React.ReactNode) => (
    <AuthorizedRoute>{element}</AuthorizedRoute>
  )

  const routes = useRoutes([
    {
      index: true,
      element: <IndexPage />,
    },
    {
      path: '/login',
      element: <LoginPage />,
    },
    {
      path: '/logout',
      element: <LogoutPage navigate={navigate} />,
    },
    {
      path: '/initialSetup',
      element: <InitialSetupPage />,
    },
    {
      path: '/share/:token/*',
      element: <SharePageTokenRoute />,
    },
    {
      path: '/albums',
      element: authorized(<AlbumsPage />),
    },
    {
      path: '/album/:id',
      element: authorized(<AlbumPage />),
    },
    {
      path: '/timeline',
      element: authorized(<TimelinePage />),
    },
    {
      path: '/places',
      element: authorized(<PlacesPage />),
    },
    {
      path: '/settings',
      element: authorized(<SettingsPage />),
    },
    {
      path: '/people',
      element: authorized(<Outlet />),
      children: [
        {
          path: ':person',
          element: <PersonPage />,
        },
        {
          index: true,
          element: <PeoplePage />,
        },
      ],
    },
    {
      // for backwards-compatibility
      path: '/photos',
      element: <Navigate to="/timeline" />,
    },
    {
      path: '*',
      element: <NotFoundPage t={t} />,
    },
  ])

  return (
    <React.Suspense
      fallback={
        <Layout title={t('general.loading.page', 'Loading page')}>
          <Loader message={t('general.loading.page', 'Loading page')} active />
        </Layout>
      }
    >
      {routes}
    </React.Suspense>
  )
}

const IndexPage = () => {
  const token = authToken()

  const dest = token ? '/timeline' : '/login'

  return <Navigate to={dest} />
}

export const NotFoundPage = ({ t }: { t: TFunction }) => {
  return <div>{t('routes.page_not_found', 'Page not found')}</div>
}

const LogoutPage = ({ navigate }: { navigate: NavigateFunction }) => {
  clearTokenCookie()
  navigate('/')
  return null
}

export default Routes
