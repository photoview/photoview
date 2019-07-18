import React, { Component } from 'react'
import { Route, Switch } from 'react-router-dom'

import LoginPage from './Pages/LoginPage'
import HomePage from './Pages/HomePage'
import AlbumPage from './Pages/AlbumPage'

class Routes extends Component {
  render() {
    return (
      <Switch>
        <Route path="/login" component={LoginPage} />
        <Route exact path="/" component={HomePage} />
        <Route path="/album/:id" component={AlbumPage} />
      </Switch>
    )
  }
}

export default Routes
