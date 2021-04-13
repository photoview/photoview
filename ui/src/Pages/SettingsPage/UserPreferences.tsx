import { useMutation, useQuery } from '@apollo/client'
import gql from 'graphql-tag'
import React from 'react'
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
  { key: 2, text: 'Dansk', flag: 'dk', value: LanguageTranslation.Danish },
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
        options={languagePreferences}
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
