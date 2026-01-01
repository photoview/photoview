import { useMutation, useQuery } from '@apollo/client'
import gql from 'graphql-tag'
import React, { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import { LanguageTranslation } from '../../__generated__/globalTypes'
import Dropdown from '../../primitives/form/Dropdown'
import { Button } from '../../primitives/form/Input'
import {
  InputLabelDescription,
  InputLabelTitle,
  SectionTitle,
} from './SettingsPage'
import {
  changeUserPreferences,
  changeUserPreferencesVariables,
} from './__generated__/changeUserPreferences'
import { myUserPreferences } from './__generated__/myUserPreferences'
import { TranslationFn } from '../../localization'
import { changeTheme, getTheme } from '../../theme'

const languagePreferences = [
  { key: 1, label: 'English', value: LanguageTranslation.English },
  { key: 2, label: 'Français', value: LanguageTranslation.French },
  { key: 3, label: 'Svenska', value: LanguageTranslation.Swedish },
  { key: 4, label: 'Dansk', value: LanguageTranslation.Danish },
  { key: 5, label: 'Español', value: LanguageTranslation.Spanish },
  { key: 6, label: 'Polski', value: LanguageTranslation.Polish },
  { key: 7, label: 'Italiano', value: LanguageTranslation.Italian },
  { key: 8, label: 'Deutsch', value: LanguageTranslation.German },
  { key: 9, label: 'Русский', value: LanguageTranslation.Russian },
  { key: 10, label: '繁體中文 (香港)', value: LanguageTranslation.TraditionalChineseHK },
  { key: 16, label: '繁體中文 (台灣)', value: LanguageTranslation.TraditionalChineseTW },
  { key: 11, label: '简体中文', value: LanguageTranslation.SimplifiedChinese },
  { key: 12, label: 'Português', value: LanguageTranslation.Portuguese },
  { key: 13, label: 'Euskara', value: LanguageTranslation.Basque },
  { key: 14, label: 'Türkçe', value: LanguageTranslation.Turkish },
  { key: 15, label: 'Українська', value: LanguageTranslation.Ukrainian },
  { key: 17, label: '日本語', value: LanguageTranslation.Japanese },
]

const themePreferences = (t: TranslationFn) => [
  {
    key: 1,
    label: t('settings.user_preferences.theme.auto.label', 'Same as system'),
    value: 'auto',
  },
  {
    key: 2,
    label: t('settings.user_preferences.theme.light.label', 'Light'),
    value: 'light',
  },
  {
    key: 2,
    label: t('settings.user_preferences.theme.dark.label', 'Dark'),
    value: 'dark',
  },
]

const landingPagePreferences = (
  t: TranslationFn,
  mapboxEnabled: boolean,
  faceDetectionEnabled: boolean
) => {
  const pages = [
    {
      key: 1,
      label: t('settings.user_preferences.landing_page.timeline', 'Timeline'),
      value: '/timeline',
    },
    {
      key: 2,
      label: t('settings.user_preferences.landing_page.albums', 'Albums'),
      value: '/albums',
    },
  ]

  if (mapboxEnabled) {
    pages.push({
      key: 3,
      label: t('settings.user_preferences.landing_page.places', 'Places'),
      value: '/places',
    })
  }

  if (faceDetectionEnabled) {
    pages.push({
      key: 4,
      label: t('settings.user_preferences.landing_page.people', 'People'),
      value: '/people',
    })
  }

  return pages
}

const CHANGE_USER_PREFERENCES = gql`
  mutation changeUserPreferences($language: String, $defaultLandingPage: String) {
    changeUserPreferences(language: $language, defaultLandingPage: $defaultLandingPage) {
      id
      language
      defaultLandingPage
    }
  }
`

const MY_USER_PREFERENCES = gql`
  query myUserPreferences {
    myUserPreferences {
      id
      language
      defaultLandingPage
    }
  }
`

const AVAILABLE_FEATURES = gql`
  query availableFeatures {
    mapboxToken
    siteInfo {
      faceDetectionEnabled
    }
  }
`

interface availableFeatures {
  mapboxToken: string | null
  siteInfo: {
    faceDetectionEnabled: boolean
  }
}

const LogoutButton = () => {
  const { t } = useTranslation()

  return (
    <Button
      className="mb-4"
      onClick={() => {
        location.href = '/logout'
      }}
    >
      {t('settings.logout', 'Log out')}
    </Button>
  )
}

const UserPreferencesWrapper = styled.div`
  margin-bottom: 24px;
`

const UserPreferences = () => {
  const { t } = useTranslation()
  const [theme, setTheme] = useState(getTheme())

  const changeStateTheme = (value: string) => {
    changeTheme(value)
    setTheme(value)
  }

  const { data: featuresData } = useQuery<availableFeatures>(AVAILABLE_FEATURES)

  const mapboxEnabled = !!featuresData?.mapboxToken
  const faceDetectionEnabled = !!featuresData?.siteInfo?.faceDetectionEnabled

  const { data } = useQuery<myUserPreferences>(MY_USER_PREFERENCES)

  const [changePrefs, { loading: loadingPrefs, error }] = useMutation<
    changeUserPreferences,
    changeUserPreferencesVariables
  >(CHANGE_USER_PREFERENCES)

  const sortedLanguagePrefs = useMemo(
    () => [...languagePreferences].sort((a, b) => a.label.localeCompare(b.label)),
    []
  )

  if (error) {
    return <div>{error.message}</div>
  }

  return (
    <UserPreferencesWrapper>
      <SectionTitle nospace>
        {t('settings.user_preferences.title', 'User preferences')}
      </SectionTitle>
      <LogoutButton />
      <label htmlFor="user_pref_change_language_field">
        <InputLabelTitle>
          {t(
            'settings.user_preferences.change_language.label',
            'Website language'
          )}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.user_preferences.change_language.description',
            'Change website language specific for this user'
          )}
        </InputLabelDescription>
      </label>
      <Dropdown
        id="user_pref_change_language_field"
        placeholder={t(
          'settings.user_preferences.language_selector.placeholder',
          'Select language'
        )}
        items={sortedLanguagePrefs}
        setSelected={language => {
          changePrefs({
            variables: {
              language: language as LanguageTranslation,
            },
          })
        }}
        selected={data?.myUserPreferences.language || undefined}
        disabled={loadingPrefs}
      />
      <label htmlFor="user_pref_change_theme_field">
        <InputLabelTitle>
          {t('settings.user_preferences.theme.title', 'Theme preferences')}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.user_preferences.theme.description',
            'Change the appearance of the website'
          )}
        </InputLabelDescription>
      </label>
      <Dropdown
        id="user_pref_change_theme_field"
        items={themePreferences(t)}
        setSelected={changeStateTheme}
        selected={theme}
      />
      <label htmlFor="user_pref_change_landing_page_field">
        <InputLabelTitle>
          {t(
            'settings.user_preferences.landing_page.label',
            'Default landing page'
          )}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.user_preferences.landing_page.description',
            'Choose the page to be shown after login'
          )}
        </InputLabelDescription>
      </label>
      <Dropdown
        id="user_pref_change_landing_page_field"
        placeholder={t(
          'settings.user_preferences.landing_page.placeholder',
          'Select default page'
        )}
        items={landingPagePreferences(t, mapboxEnabled, faceDetectionEnabled)}
        setSelected={landingPage => {
          changePrefs({
            variables: {
              defaultLandingPage: landingPage as string,
            },
          })
        }}
        selected={data?.myUserPreferences.defaultLandingPage || undefined}
        disabled={loadingPrefs}
      />
    </UserPreferencesWrapper>
  )
}

export default UserPreferences
