import React, { Component } from 'react'
import { Route, Switch } from 'react-router-dom'

import LoginPage from './Pages/LoginPage'
import HomePage from './Pages/HomePage/HomePage'
import AlbumPage from './Pages/AlbumPage/AlbumPage'
import AuthorizedRoute from './AuthorizedRoute'
import PhotosPage from './Pages/PhotosPage/PhotosPage'

class Routes extends Component {
  render() {
    return (
      <Switch>
        <Route path="/login" component={LoginPage} />
        <AuthorizedRoute exact path="/" component={HomePage} />
        <AuthorizedRoute path="/album/:id" component={AlbumPage} />
        <AuthorizedRoute path="/photos" component={PhotosPage} />
      </Switch>
    )
  }
}

export default Routes
