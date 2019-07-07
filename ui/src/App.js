import React, { Component } from 'react'
import { Redirect } from 'react-router-dom'

import Routes from './Routes'

class App extends Component {
  render() {
    const token = localStorage.getItem('token')

    if (!token) {
      return (
        <>
          <Redirect to="/login" />
          <Routes />
        </>
      )
    }

    return <Routes />
  }
}

export default App
