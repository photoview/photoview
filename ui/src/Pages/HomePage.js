import React, { Component } from 'react'
import Albums from '../Albums'

class HomePage extends Component {
  render() {
    return (
      <div>
        <h1>Home</h1>
        <Albums />
      </div>
    )
  }
}

export default HomePage
