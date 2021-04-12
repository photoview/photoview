import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

export default function setupLocalization(): void {
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
      useSuspense: process.env.NODE_ENV === 'production',
    },
  })
}
