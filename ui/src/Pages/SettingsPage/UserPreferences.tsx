import { useMutation, useQuery } from '@apollo/client'
import gql from 'graphql-tag'
import React, { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Dropdown } from 'semantic-ui-react'
import styled from 'styled-components'
import { LanguageTranslation } from '../../../__generated__/globalTypes'
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

const languagePreferences = [
  { key: 1, text: 'English', flag: 'uk', value: LanguageTranslation.English },
  { key: 2, text: 'Français', flag: 'fr', value: LanguageTranslation.French },
  { key: 3, text: 'Svenska', flag: 'se', value: LanguageTranslation.Swedish },
  { key: 4, text: 'Dansk', flag: 'dk', value: LanguageTranslation.Danish },
  { key: 5, text: 'Español', flag: 'es', value: LanguageTranslation.Spanish },
  { key: 6, text: 'polski', flag: 'pl', value: LanguageTranslation.Polish },
  { key: 7, text: 'Italiano', flag: 'it', value: LanguageTranslation.Italian },
  { key: 8, text: 'Deutsch', flag: 'de', value: LanguageTranslation.German },
]

const CHANGE_USER_PREFERENCES = gql`
  mutation changeUserPreferences($language: String) {
    changeUserPreferences(language: $language) {
      id
      language
    }
  }
`

const MY_USER_PREFERENCES = gql`
  query myUserPreferences {
    myUserPreferences {
      id
      language
    }
  }
`

const UserPreferencesWrapper = styled.div`
  margin-bottom: 24px;
`

const UserPreferences = () => {
  const { t } = useTranslation()

  const { data } = useQuery<myUserPreferences>(MY_USER_PREFERENCES)

  const [changePrefs, { loading: loadingPrefs, error }] = useMutation<
    changeUserPreferences,
    changeUserPreferencesVariables
  >(CHANGE_USER_PREFERENCES)

  const sortedLanguagePrefs = useMemo(
    () => languagePreferences.sort((a, b) => a.text.localeCompare(b.text)),
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
      <label id="user_pref_change_language_field">
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
        clearable
        options={sortedLanguagePrefs}
        onChange={(event, { value: language }) => {
          changePrefs({
            variables: {
              language: language as LanguageTranslation,
            },
          })
        }}
        selection
        search
        value={data?.myUserPreferences.language || undefined}
        loading={loadingPrefs}
        disabled={loadingPrefs}
      />
    </UserPreferencesWrapper>
  )
}

export default UserPreferences
