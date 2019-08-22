import React, { Component } from 'react'
import { createGlobalStyle } from 'styled-components'
import Routes from './Routes'
import Messages from './components/messages/Messages'

const GlobalStyle = createGlobalStyle`
  #root, body {
    height: 100%;
    margin: 0;
  }
`

class App extends Component {
  render() {
    return (
      <>
        <GlobalStyle />
        <Routes />
        <Messages />
      </>
    )
  }
}

export default App
