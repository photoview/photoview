import React, { Component } from 'react'
import { Route, Switch, Redirect } from 'react-router-dom'

import LoginPage from './Pages/LoginPage'
import AlbumsPage from './Pages/AllAlbumsPage/AlbumsPage'
import AlbumPage from './Pages/AlbumPage/AlbumPage'
import AuthorizedRoute from './AuthorizedRoute'
import PhotosPage from './Pages/PhotosPage/PhotosPage'

class Routes extends Component {
  render() {
    return (
      <Switch>
        <Route path="/login" component={LoginPage} />
        <AuthorizedRoute exact path="/albums" component={AlbumsPage} />
        <AuthorizedRoute path="/album/:id" component={AlbumPage} />
        <AuthorizedRoute path="/photos" component={PhotosPage} />
        <Route path="/" exact render={() => <Redirect to="photos" />} />
        <Route render={() => <div>Page not found</div>} />
      </Switch>
    )
  }
}

export default Routes
