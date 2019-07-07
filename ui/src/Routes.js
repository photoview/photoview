import React, { Component } from 'react'
import { Route, Switch } from 'react-router-dom'

import LoginPage from './Pages/LoginPage'
import HomePage from './Pages/HomePage'

class Routes extends Component {
  render() {
    return (
      <Switch>
        <Route path="/login" component={LoginPage} />
        <Route path="/" component={HomePage} />
      </Switch>
    )
  }
}

export default Routes
