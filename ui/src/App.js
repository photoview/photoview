import React, { Component } from 'react'
import { createGlobalStyle } from 'styled-components'
import { Helmet } from 'react-helmet'
import Routes from './Routes'
import Messages from './components/messages/Messages'

const GlobalStyle = createGlobalStyle`
  #root, body {
    height: 100%;
    margin: 0;
  }
`

import 'semantic-ui-css/components/reset.css'
import 'semantic-ui-css/components/site.css'
import 'semantic-ui-css/components/transition.css'
import 'semantic-ui-css/components/menu.css'

class App extends Component {
  render() {
    return (
      <>
        <Helmet>
          <meta
            name="description"
            content="Simple and User-friendly Photo Gallery for Personal Servers"
          />
        </Helmet>
        <GlobalStyle />
        <Routes />
        <Messages />
      </>
    )
  }
}

export default App
