import React from 'react'
import { Route, Switch, Redirect } from 'react-router-dom'

import { Loader } from 'semantic-ui-react'
import Layout from './Layout'
import { clearTokenCookie } from './authentication'

const AlbumsPage = React.lazy(() => import('./Pages/AllAlbumsPage/AlbumsPage'))
const AlbumPage = React.lazy(() => import('./Pages/AlbumPage/AlbumPage'))
const AuthorizedRoute = React.lazy(() => import('./AuthorizedRoute'))
const PhotosPage = React.lazy(() => import('./Pages/PhotosPage/PhotosPage'))
const SharePage = React.lazy(() => import('./Pages/SharePage/SharePage'))

const LoginPage = React.lazy(() => import('./Pages/LoginPage/LoginPage'))
const InitialSetupPage = React.lazy(() =>
  import('./Pages/LoginPage/InitialSetupPage')
)

const SettingsPage = React.lazy(() =>
  import('./Pages/SettingsPage/SettingsPage')
)

class Routes extends React.Component {
  render() {
    return (
      <React.Suspense
        fallback={
          <Layout>
            <Loader active>Loading page</Loader>
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
          <AuthorizedRoute path="/album/:id/:subPage?" component={AlbumPage} />
          <AuthorizedRoute path="/photos/:subPage?" component={PhotosPage} />
          <AuthorizedRoute admin path="/settings" component={SettingsPage} />
          <Route path="/" exact render={() => <Redirect to="/photos" />} />
          <Route render={() => <div>Page not found</div>} />
        </Switch>
      </React.Suspense>
    )
  }
}

export default Routes
