import React from 'react'
import { createRoot } from 'react-dom/client'
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

const root = createRoot(document.getElementById('root')!)
root.render(<Main />)

serviceWorkerRegistration.register()
