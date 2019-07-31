import React from 'react'
import { Route, Switch, Redirect } from 'react-router-dom'

import { Loader } from 'semantic-ui-react'

const AlbumsPage = React.lazy(() => import('./Pages/AllAlbumsPage/AlbumsPage'))
const AlbumPage = React.lazy(() => import('./Pages/AlbumPage/AlbumPage'))
const AuthorizedRoute = React.lazy(() => import('./AuthorizedRoute'))
const PhotosPage = React.lazy(() => import('./Pages/PhotosPage/PhotosPage'))

const LoginPage = React.lazy(() => import('./Pages/LoginPage/LoginPage'))
const InitialSetupPage = React.lazy(() =>
  import('./Pages/LoginPage/InitialSetupPage')
)

class Routes extends React.Component {
  render() {
    return (
      <React.Suspense fallback={<Loader active>Loading page</Loader>}>
        <Switch>
          <Route path="/login" component={LoginPage} />
          <Route path="/initialSetup" component={InitialSetupPage} />
          <AuthorizedRoute exact path="/albums" component={AlbumsPage} />
          <AuthorizedRoute path="/album/:id" component={AlbumPage} />
          <AuthorizedRoute path="/photos" component={PhotosPage} />
          <Route path="/" exact render={() => <Redirect to="photos" />} />
          <Route render={() => <div>Page not found</div>} />
        </Switch>
      </React.Suspense>
    )
  }
}

export default Routes
