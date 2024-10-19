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
import { ReDetectModalProvider } from './components/sidebar/ReDetection/ReDetectFacesContext'

updateTheme()
setupLocalization()

const Main = () => (
  <ApolloProvider client={client}>
    <Router basename={import.meta.env.BASE_URL}>
      <SidebarProvider>
        <ReDetectModalProvider>
          <App />
        </ReDetectModalProvider>
      </SidebarProvider>
    </Router>
  </ApolloProvider>
)

const root = createRoot(document.getElementById('root')!)
root.render(<Main />)

serviceWorkerRegistration.register()
