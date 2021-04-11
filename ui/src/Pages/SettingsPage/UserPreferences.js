import { useMutation, useQuery } from '@apollo/client'
import gql from 'graphql-tag'
import React from 'react'
import { useTranslation } from 'react-i18next'
import { Dropdown } from 'semantic-ui-react'
import styled from 'styled-components'

import { SectionTitle } from './SettingsPage'

const languagePreferences = [
  { key: 1, text: 'English', value: 'en' },
  { key: 2, text: 'Dansk', value: 'da' },
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

  const { data } = useQuery(MY_USER_PREFERENCES)

  const [changePrefs, { loading: loadingPrefs, error }] = useMutation(
    CHANGE_USER_PREFERENCES
  )

  if (error) {
    return error.message
  }

  return (
    <UserPreferencesWrapper>
      <SectionTitle nospace>
        {t('settings.user_preferences.title', 'User preferences')}
      </SectionTitle>
      <Dropdown
        placeholder={t(
          'settings.user_preferences.language_selector.placeholder',
          'Select language'
        )}
        clearable
        options={languagePreferences}
        onChange={(event, { value: language }) => {
          changePrefs({
            variables: {
              language,
            },
          })
        }}
        selection
        value={data?.myUserPreferences.language}
        loading={loadingPrefs}
        disabled={loadingPrefs}
      />
    </UserPreferencesWrapper>
  )
}

export default UserPreferences
