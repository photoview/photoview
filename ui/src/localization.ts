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
  i18n
    .use(initReactI18next)
    .init({
      lng: 'en',
      fallbackLng: 'en',
      returnNull: false,
      returnEmptyString: false,

      interpolation: {
        escapeValue: false,
      },

      react: {
        useSuspense: import.meta.env.PROD,
      },
    })
    .catch(err => console.error('Failed to setup localization', err))
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
        import('./extractedTranslations/da/translation.json').then(language => {
          i18n.addResourceBundle('da', 'translation', language)
          i18n.changeLanguage('da')
        })
        return
      case LanguageTranslation.English:
        import('./extractedTranslations/en/translation.json').then(language => {
          i18n.addResourceBundle('en', 'translation', language)
          i18n.changeLanguage('en')
        })
        return
      case LanguageTranslation.French:
        import('./extractedTranslations/fr/translation.json').then(language => {
          i18n.addResourceBundle('fr', 'translation', language)
          i18n.changeLanguage('fr')
        })
        return
      case LanguageTranslation.Swedish:
        import('./extractedTranslations/sv/translation.json').then(language => {
          i18n.addResourceBundle('sv', 'translation', language)
          i18n.changeLanguage('sv')
        })
        return
      case LanguageTranslation.Italian:
        import('./extractedTranslations/it/translation.json').then(language => {
          i18n.addResourceBundle('it', 'translation', language)
          i18n.changeLanguage('it')
        })
        return
      case LanguageTranslation.Spanish:
        import('./extractedTranslations/es/translation.json').then(language => {
          i18n.addResourceBundle('es', 'translation', language)
          i18n.changeLanguage('es')
        })
        return
      case LanguageTranslation.Polish:
        import('./extractedTranslations/pl/translation.json').then(language => {
          i18n.addResourceBundle('pl', 'translation', language)
          i18n.changeLanguage('pl')
        })
        return
      case LanguageTranslation.German:
        import('./extractedTranslations/de/translation.json').then(language => {
          i18n.addResourceBundle('de', 'translation', language)
          i18n.changeLanguage('de')
        })
        return
      case LanguageTranslation.Russian:
        import('./extractedTranslations/ru/translation.json').then(language => {
          i18n.addResourceBundle('ru', 'translation', language)
          i18n.changeLanguage('ru')
        })
        return
      case LanguageTranslation.TraditionalChinese:
        import('./extractedTranslations/zh-HK/translation.json').then(
          language => {
            i18n.addResourceBundle('zh-HK', 'translation', language)
            i18n.changeLanguage('zh-HK')
          }
        )
        return
      case LanguageTranslation.SimplifiedChinese:
        import('./extractedTranslations/zh-CN/translation.json').then(
          language => {
            i18n.addResourceBundle('zh-CN', 'translation', language)
            i18n.changeLanguage('zh-CN')
          }
        )
        return
      case LanguageTranslation.Portuguese:
        import('./extractedTranslations/pt/translation.json').then(language => {
          i18n.addResourceBundle('pt', 'translation', language)
          i18n.changeLanguage('pt')
        })
        return
      case LanguageTranslation.Basque:
        import('./extractedTranslations/eu/translation.json').then(language => {
          i18n.addResourceBundle('eu', 'translation', language)
          i18n.changeLanguage('eu')
        })
        return
    }

    exhaustiveCheck(language)
  }, [data?.myUserPreferences.language])
}
