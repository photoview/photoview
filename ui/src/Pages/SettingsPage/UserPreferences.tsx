import { useMutation, useQuery } from '@apollo/client'
import gql from 'graphql-tag'
import React, { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import { LanguageTranslation } from '../../__generated__/globalTypes'
import Dropdown from '../../primitives/form/Dropdown'
import { Button, TextField } from '../../primitives/form/Input'
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
import { myUsername } from './__generated__/myUsername'
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
  {
    key: 10,
    label: '繁體中文 (香港)',
    value: LanguageTranslation.TraditionalChineseHK,
  },
  {
    key: 16,
    label: '繁體中文 (台灣)',
    value: LanguageTranslation.TraditionalChineseTW,
  },
  { key: 11, label: '简体中文', value: LanguageTranslation.SimplifiedChinese },
  { key: 12, label: 'Português', value: LanguageTranslation.Portuguese },
  { key: 13, label: 'Euskara', value: LanguageTranslation.Basque },
  { key: 14, label: 'Türkçe', value: LanguageTranslation.Turkish },
  { key: 15, label: 'Українська', value: LanguageTranslation.Ukrainian },
  { key: 17, label: '日本語', value: LanguageTranslation.Japanese },
  { key: 18, label: 'Nederlands', value: LanguageTranslation.Dutch },
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

export const CHANGE_USER_PREFERENCES = gql`
  mutation changeUserPreferences($language: String, $searchResultLimit: Int) {
    changeUserPreferences(
      language: $language
      searchResultLimit: $searchResultLimit
    ) {
      id
      language
      searchResultLimit
    }
  }
`

export const MY_USER_PREFERENCES = gql`
  query myUserPreferences {
    myUserPreferences {
      id
      language
      searchResultLimit
    }
  }
`

export const MY_USERNAME_QUERY = gql`
  query myUsername {
    myUser {
      id
      username
    }
  }
`

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

// The largest value GraphQL's Int can carry.
const MAX_SEARCH_RESULT_LIMIT = 2147483647

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

  const { data, loading: loadingSavedPrefs } =
    useQuery<myUserPreferences>(MY_USER_PREFERENCES)
  const { data: usernameData } = useQuery<myUsername>(MY_USERNAME_QUERY)

  const [changePrefs, { loading: loadingPrefs, error }] = useMutation<
    changeUserPreferences,
    changeUserPreferencesVariables
  >(CHANGE_USER_PREFERENCES)

  const sortedLanguagePrefs = useMemo(
    () =>
      [...languagePreferences].sort((a, b) => a.label.localeCompare(b.label)),
    []
  )

  const savedSearchLimit = data?.myUserPreferences.searchResultLimit
  const [searchLimitInput, setSearchLimitInput] = useState('')

  const showSavedSearchLimit = () =>
    setSearchLimitInput(
      savedSearchLimit == null ? '' : String(savedSearchLimit)
    )

  useEffect(showSavedSearchLimit, [savedSearchLimit])

  const commitSearchLimit = () => {
    const trimmed = searchLimitInput.trim()
    const parsed = Number(trimmed)

    // Emptying the field is how a user gets back to "whatever the server does
    // by default". 0 already means unlimited, so it cannot also mean that -
    // the mutation takes a negative value for it instead.
    if (trimmed === '') {
      if (savedSearchLimit == null) return

      changePrefs({ variables: { searchResultLimit: -1 } })
      return
    }

    // A nonsensical field puts the saved value back, rather than leaving it
    // showing something that was never stored.
    //
    // The upper bound is GraphQL's Int: a larger number is rejected before
    // the mutation runs, and that error would replace this whole page.
    if (
      !Number.isInteger(parsed) ||
      parsed < 0 ||
      parsed > MAX_SEARCH_RESULT_LIMIT
    ) {
      showSavedSearchLimit()
      return
    }

    if (parsed === savedSearchLimit) return

    changePrefs({ variables: { searchResultLimit: parsed } })
  }

  if (error) {
    return <div>{error.message}</div>
  }

  return (
    <UserPreferencesWrapper>
      <SectionTitle nospace>
        {usernameData?.myUser
          ? t(
              'settings.user_preferences.title_with_username',
              'User preferences ({{username}})',
              { username: usernameData.myUser.username }
            )
          : t('settings.user_preferences.title', 'User preferences')}
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
      <label htmlFor="user_pref_search_result_limit_field">
        <InputLabelTitle>
          {t(
            'settings.user_preferences.search_result_limit.label',
            'Search results'
          )}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.user_preferences.search_result_limit.description',
            'How many albums and how many media files a search returns. Enter 0 to show every match, or leave it empty to use the default.'
          )}
        </InputLabelDescription>
      </label>
      <TextField
        id="user_pref_search_result_limit_field"
        type="number"
        min={0}
        max={MAX_SEARCH_RESULT_LIMIT}
        step={1}
        placeholder="10"
        value={searchLimitInput}
        // Also while the saved value is still loading: it would otherwise
        // land in the field and overwrite whatever was typed meanwhile.
        disabled={loadingPrefs || loadingSavedPrefs}
        onChange={e => setSearchLimitInput(e.target.value)}
        onBlur={commitSearchLimit}
        action={commitSearchLimit}
      />
    </UserPreferencesWrapper>
  )
}

export default UserPreferences
