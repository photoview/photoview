import 'regenerator-runtime/runtime'

import React from 'react'
import ReactDOM from 'react-dom'
import App from './App'
import registerServiceWorker from './registerServiceWorker'
import client from './apolloClient'
import { ApolloProvider } from '@apollo/client'
import { BrowserRouter as Router } from 'react-router-dom'
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

i18n.use(initReactI18next).init({
  resources: {
    en: {
      translation: {
        'Welcome to React': 'Welcome to React and react-i18next',
      },
    },
  },
  lng: 'en',
  fallbackLng: 'en',
  returnNull: false,

  interpolation: {
    escapeValue: false,
  },
})

import('../extractedTranslations/da/translation.json').then(danish => {
  i18n.addResourceBundle('da', 'translation', danish)
  console.log('loaded danish')
  i18n.changeLanguage('da')
})

const Main = () => (
  <ApolloProvider client={client}>
    <Router>
      <App />
    </Router>
  </ApolloProvider>
)

ReactDOM.render(<Main />, document.getElementById('root'))

registerServiceWorker()
