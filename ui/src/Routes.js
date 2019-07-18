import React, { Component } from 'react'
import { Route, Switch } from 'react-router-dom'

import LoginPage from './Pages/LoginPage'
import HomePage from './Pages/HomePage/HomePage'
import AlbumPage from './Pages/AlbumPage/AlbumPage'

class Routes extends Component {
  render() {
    const token = localStorage.getItem('token')

    let unauthorizedRedirect = null
    if (!token) {
      unauthorizedRedirect = <Redirect to="/login" />
    }

    return (
      <Switch>
        {unauthorizedRedirect}
        <Route path="/login" component={LoginPage} />
        <Route exact path="/" component={HomePage} />
        <Route path="/album/:id" component={AlbumPage} />
      </Switch>
    )
  }
}

export default Routes
