import React, { useEffect } from 'react'
import { createGlobalStyle } from 'styled-components'
import { Helmet } from 'react-helmet'
import Routes from './components/routes/Routes'
import Messages from './components/messages/Messages'
import i18n from 'i18next'
import { gql, useLazyQuery } from '@apollo/client'
import { authToken } from './helpers/authentication'

const GlobalStyle = createGlobalStyle`
  * {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  html {
    font-size: 0.85rem;
  }

  #root, body {
    height: 100%;
    margin: 0;
    font-size: inherit;
  }

  /* Make dimmer lighter */
  .ui.dimmer {
    background-color: rgba(0, 0, 0, 0.5);
  }
`

import 'semantic-ui-css/semantic.min.css'

const SITE_TRANSLATION = gql`
  query {
    myUserPreferences {
      id
      language
    }
  }
`

const loadTranslations = () => {
  const [loadLang, { data }] = useLazyQuery(SITE_TRANSLATION)

  useEffect(() => {
    if (authToken()) {
      loadLang()
    }
  }, [authToken()])

  useEffect(() => {
    switch (data?.myUserPreferences.language) {
      case 'da':
        import('../extractedTranslations/da/translation.json').then(danish => {
          i18n.addResourceBundle('da', 'translation', danish)
          i18n.changeLanguage('da')
        })
        break
      default:
        i18n.changeLanguage('en')
    }
  }, [data?.myUserPreferences.language])
}

const App = () => {
  loadTranslations()

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

export default App
