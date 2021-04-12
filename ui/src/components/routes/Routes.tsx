import React from 'react'
import { Route, Switch, Redirect } from 'react-router-dom'

import { Loader } from 'semantic-ui-react'
import Layout from '../../Layout'
import { clearTokenCookie } from '../../helpers/authentication'
import { useTranslation } from 'react-i18next'

const AuthorizedRoute = React.lazy(() => import('./AuthorizedRoute'))

const AlbumsPage = React.lazy(
  () => import('../../Pages/AllAlbumsPage/AlbumsPage')
)
const AlbumPage = React.lazy(() => import('../../Pages/AlbumPage/AlbumPage'))
const PhotosPage = React.lazy(() => import('../../Pages/PhotosPage/PhotosPage'))
const PlacesPage = React.lazy(() => import('../../Pages/PlacesPage/PlacesPage'))
const SharePage = React.lazy(() => import('../../Pages/SharePage/SharePage'))
const PeoplePage = React.lazy(() => import('../../Pages/PeoplePage/PeoplePage'))

const LoginPage = React.lazy(() => import('../../Pages/LoginPage/LoginPage'))
const InitialSetupPage = React.lazy(
  () => import('../../Pages/LoginPage/InitialSetupPage')
)

const SettingsPage = React.lazy(
  () => import('../../Pages/SettingsPage/SettingsPage')
)

const Routes = () => {
  const { t } = useTranslation()

  return (
    <React.Suspense
      fallback={
        <Layout title={t('general.loading.page', 'Loading page')}>
          <Loader active>{t('general.loading.page', 'Loading page')}</Loader>
        </Layout>
      }
    >
      <Switch>
        <Route path="/login" component={LoginPage} />
        <Route path="/logout">
          {() => {
            clearTokenCookie()
            location.href = '/'
          }}
        </Route>
        <Route path="/initialSetup" component={InitialSetupPage} />
        <Route path="/share" component={SharePage} />
        <AuthorizedRoute exact path="/albums" component={AlbumsPage} />
        <AuthorizedRoute path="/album/:id" component={AlbumPage} />
        <AuthorizedRoute path="/photos" component={PhotosPage} />
        <AuthorizedRoute path="/places" component={PlacesPage} />
        <AuthorizedRoute path="/people/:person?" component={PeoplePage} />
        <AuthorizedRoute path="/settings" component={SettingsPage} />
        <Route path="/" exact render={() => <Redirect to="/photos" />} />
        <Route
          render={() => (
            <div>{t('routes.page_not_found', 'Page not found')}</div>
          )}
        />
      </Switch>
    </React.Suspense>
  )
}

export default Routes
