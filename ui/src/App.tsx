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
import { siteTranslation } from './__generated__/siteTranslation'
import { LanguageTranslation } from '../__generated__/globalTypes'
import { useTranslation } from 'react-i18next'

const SITE_TRANSLATION = gql`
  query siteTranslation {
    myUserPreferences {
      id
      language
    }
  }
`

const loadTranslations = () => {
  console.log('load translation')
  const [loadLang, { data }] = useLazyQuery<siteTranslation>(SITE_TRANSLATION)

  useEffect(() => {
    if (authToken()) {
      loadLang()
    }
  }, [authToken()])

  useEffect(() => {
    console.log('loading translations', data)
    switch (data?.myUserPreferences.language) {
      case LanguageTranslation.Danish:
        import('../extractedTranslations/da/translation.json').then(danish => {
          console.log('loading danish')
          i18n.addResourceBundle('da', 'translation', danish)
          i18n.changeLanguage('da')
        })
        break
      case LanguageTranslation.English:
        import('../extractedTranslations/en/translation.json').then(english => {
          console.log('loading english')
          i18n.addResourceBundle('en', 'translation', english)
          i18n.changeLanguage('en')
        })
        break
      default:
        i18n.changeLanguage('en')
    }
  }, [data?.myUserPreferences.language])
}

const App = () => {
  const { t } = useTranslation()
  loadTranslations()

  return (
    <>
      <Helmet>
        <meta
          name="description"
          content={t(
            'meta.description',
            'Simple and User-friendly Photo Gallery for Personal Servers'
          )}
        />
      </Helmet>
      <GlobalStyle />
      <Routes />
      <Messages />
    </>
  )
}

export default App
