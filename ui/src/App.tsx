import React from 'react'
import { Helmet } from 'react-helmet'
import Routes from './components/routes/Routes'
import Messages from './components/messages/Messages'
import { useTranslation } from 'react-i18next'
import { loadTranslations } from './localization'

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
      <Routes />
      <Messages />
    </>
  )
}

export default App
