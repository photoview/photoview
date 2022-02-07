import 'regenerator-runtime/runtime'

import React from 'react'
import ReactDOM from 'react-dom'
import App from './App'
import client from './apolloClient'
import { ApolloProvider } from '@apollo/client'
import { BrowserRouter as Router } from 'react-router-dom'
import { setupLocalization } from './localization'
import { updateTheme } from './theme'
import * as serviceWorkerRegistration from './serviceWorkerRegistration'

import './index.css'
import { SidebarProvider } from './components/sidebar/Sidebar'

updateTheme()
setupLocalization()

const Main = () => (
  <ApolloProvider client={client}>
    <Router>
      <SidebarProvider>
        <App />
      </SidebarProvider>
    </Router>
  </ApolloProvider>
)

ReactDOM.render(<Main />, document.getElementById('root'))

serviceWorkerRegistration.register()
