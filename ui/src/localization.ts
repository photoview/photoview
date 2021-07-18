import { useEffect } from 'react'
import { siteTranslation } from './__generated__/siteTranslation'
import { gql, useLazyQuery } from '@apollo/client'
import i18n from 'i18next'
import { initReactI18next, TFunction } from 'react-i18next'
import { LanguageTranslation } from './__generated__/globalTypes'
import { authToken } from './helpers/authentication'
import { exhaustiveCheck, isNil } from './helpers/utils'

export type TranslationFn = TFunction<'translation'>

export function setupLocalization(): void {
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

    react: {
      useSuspense: process.env.NODE_ENV == 'production',
    },
  })
}

const SITE_TRANSLATION = gql`
  query siteTranslation {
    myUserPreferences {
      id
      language
    }
  }
`

export const loadTranslations = () => {
  const [loadLang, { data }] = useLazyQuery<siteTranslation>(SITE_TRANSLATION)

  useEffect(() => {
    if (authToken()) {
      loadLang()
    }
  }, [authToken()])

  useEffect(() => {
    const language = data?.myUserPreferences.language
    if (isNil(language)) {
      i18n.changeLanguage('en')
      return
    }

    switch (language) {
      case LanguageTranslation.Danish:
        import('./extractedTranslations/da/translation.json').then(danish => {
          i18n.addResourceBundle('da', 'translation', danish)
          i18n.changeLanguage('da')
        })
        return
      case LanguageTranslation.English:
        import('./extractedTranslations/en/translation.json').then(english => {
          i18n.addResourceBundle('en', 'translation', english)
          i18n.changeLanguage('en')
        })
        return
      case LanguageTranslation.French:
        import('./extractedTranslations/fr/translation.json').then(english => {
          i18n.addResourceBundle('fr', 'translation', english)
          i18n.changeLanguage('fr')
        })
        return
      case LanguageTranslation.Swedish:
        import('./extractedTranslations/sv/translation.json').then(swedish => {
          i18n.addResourceBundle('sv', 'translation', swedish)
          i18n.changeLanguage('sv')
        })
        return
      case LanguageTranslation.Italian:
        import('./extractedTranslations/it/translation.json').then(italian => {
          i18n.addResourceBundle('it', 'translation', italian)
          i18n.changeLanguage('it')
        })
        return
      case LanguageTranslation.Spanish:
        import('./extractedTranslations/es/translation.json').then(spanish => {
          i18n.addResourceBundle('es', 'translation', spanish)
          i18n.changeLanguage('es')
        })
        return
      case LanguageTranslation.Polish:
        import('./extractedTranslations/pl/translation.json').then(polish => {
          i18n.addResourceBundle('pl', 'translation', polish)
          i18n.changeLanguage('pl')
        })
        return
      case LanguageTranslation.German:
        import('./extractedTranslations/de/translation.json').then(german => {
          i18n.addResourceBundle('de', 'translation', german)
          i18n.changeLanguage('de')
        })
        return
    }

    exhaustiveCheck(language)
  }, [data?.myUserPreferences.language])
}
