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
import { MessageProvider } from './components/messages/MessageState'

updateTheme()
setupLocalization()

const Main = () => (
  <ApolloProvider client={client}>
    <Router basename={import.meta.env.BASE_URL}>
      <MessageProvider>
        <SidebarProvider>
          <App />
        </SidebarProvider>
      </MessageProvider>
    </Router >
  </ApolloProvider >
)

const root = createRoot(document.getElementById('root')!)
root.render(<Main />)

serviceWorkerRegistration.register()
