import React, { Component } from 'react'
import Albums from './Albums'
import Layout from '../../Layout'

class HomePage extends Component {
  render() {
    return (
      <Layout>
        <h1>Albums</h1>
        <Albums />
      </Layout>
    )
  }
}

export default HomePage
