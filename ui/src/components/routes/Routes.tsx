import React from 'react'
import {
  Route,
  Routes as RouterRoutes,
  Navigate,
  useNavigate,
  NavigateFunction,
  Outlet,
} from 'react-router-dom'

import Layout from '../layout/Layout'
import { authToken, clearTokenCookie } from '../../helpers/authentication'
import { TFunction, useTranslation } from 'react-i18next'
import Loader from '../../primitives/Loader'
import AuthorizedRoute from './AuthorizedRoute'
import sharePageRoute from '../../Pages/SharePage/sharePageRoute'
import peoplePageRoute from '../../Pages/PeoplePage/peoplePageRoute'

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

const SettingsPage = React.lazy(
  () => import('../../Pages/SettingsPage/SettingsPage')
)

const Routes = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()

  return (
    <React.Suspense
      fallback={
        <Layout title={t('general.loading.page', 'Loading page')}>
          <Loader message={t('general.loading.page', 'Loading page')} active />
        </Layout>
      }
    >
      <RouterRoutes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/logout" element={<LogoutPage navigate={navigate} />} />
        <Route path="/initialSetup" element={<InitialSetupPage />} />
        <Route path="/share">{sharePageRoute({ t })}</Route>
        <Route
          path="/albums"
          element={
            <AuthorizedRoute>
              <AlbumsPage />
            </AuthorizedRoute>
          }
        />
        <Route
          path="/album/:id"
          element={
            <AuthorizedRoute>
              <AlbumPage />
            </AuthorizedRoute>
          }
        />
        <Route
          path="/timeline"
          element={
            <AuthorizedRoute>
              <TimelinePage />
            </AuthorizedRoute>
          }
        />
        <Route
          path="/places"
          element={
            <AuthorizedRoute>
              <PlacesPage />
            </AuthorizedRoute>
          }
        />
        <Route
          path="/people"
          element={
            <AuthorizedRoute>
              <Outlet />
            </AuthorizedRoute>
          }
        >
          {peoplePageRoute()}
        </Route>
        <Route
          path="/settings"
          element={
            <AuthorizedRoute>
              <SettingsPage />
            </AuthorizedRoute>
          }
        />
        <Route index element={<IndexPage />} />
        {/* For backwards compatibility */}
        <Route path="/photos" element={<Navigate to="/timeline" />} />
        <Route element={<NotFoundPage t={t} />} />
      </RouterRoutes>
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
