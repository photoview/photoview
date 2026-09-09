import { useMutation, useQuery } from '@apollo/client'
import gql from 'graphql-tag'
import React, { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import { LanguageTranslation } from '../../__generated__/globalTypes'
import Checkbox from '../../primitives/form/Checkbox'
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
import { unhideAllAlbums } from './__generated__/unhideAllAlbums'
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

const CHANGE_USER_PREFERENCES = gql`
  mutation changeUserPreferences(
    $language: String
    $searchResultLimit: Int
    $showAlbumTree: Boolean
    $showHiddenAlbums: Boolean
  ) {
    changeUserPreferences(
      language: $language
      searchResultLimit: $searchResultLimit
      showAlbumTree: $showAlbumTree
      showHiddenAlbums: $showHiddenAlbums
    ) {
      id
      language
      searchResultLimit
      showAlbumTree
      showHiddenAlbums
    }
  }
`

const MY_USER_PREFERENCES = gql`
  query myUserPreferences {
    myUserPreferences {
      id
      language
      searchResultLimit
      showAlbumTree
      showHiddenAlbums
    }
  }
`

const MY_USERNAME_QUERY = gql`
  query myUsername {
    myUser {
      id
      username
    }
  }
`

const UNHIDE_ALL_ALBUMS = gql`
  mutation unhideAllAlbums {
    unhideAllAlbums
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

  const { data } = useQuery<myUserPreferences>(MY_USER_PREFERENCES)
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

  // Preferences haven't loaded yet: don't let interactions fire mutations
  // with these placeholder values, since changeUserPreferences persists all
  // three fields on every call and would otherwise clobber the real ones.
  const preferencesLoaded = data?.myUserPreferences != null
  const currentLanguage = data?.myUserPreferences.language ?? null
  const currentSearchResultLimit =
    data?.myUserPreferences.searchResultLimit ?? null
  const currentShowAlbumTree = data?.myUserPreferences.showAlbumTree ?? true
  const currentShowHiddenAlbums =
    data?.myUserPreferences.showHiddenAlbums ?? false

  const [unhideAllAlbums, { loading: unhideAllLoading }] =
    useMutation<unhideAllAlbums>(UNHIDE_ALL_ALBUMS, {
      // The album tree only shows unhidden albums by default - without a
      // refetch, restored albums stay filtered out of an already-loaded
      // tree even though the mutation succeeded.
      refetchQueries: ['albumTreeRootQuery', 'albumTreeSubAlbumsQuery'],
      // The global Apollo error link already shows a toast; this only
      // consumes the rejected promise so it isn't left unhandled.
      onError: () => undefined,
    })

  const [searchResultLimitInput, setSearchResultLimitInput] = useState('')

  useEffect(() => {
    setSearchResultLimitInput(
      currentSearchResultLimit != null ? String(currentSearchResultLimit) : ''
    )
  }, [currentSearchResultLimit])

  const commitSearchResultLimit = () => {
    if (!preferencesLoaded) return

    const trimmed = searchResultLimitInput.trim()
    const parsed = trimmed === '' ? null : Number(trimmed)

    // $searchResultLimit is a GraphQL Int (signed 32-bit); reject values
    // outside that range instead of letting the mutation get rejected.
    const isValid =
      parsed === null ||
      (Number.isInteger(parsed) && parsed >= 0 && parsed <= 2147483647)

    if (!isValid) {
      setSearchResultLimitInput(
        currentSearchResultLimit != null ? String(currentSearchResultLimit) : ''
      )
      return
    }

    if (parsed === currentSearchResultLimit) return

    changePrefs({
      variables: {
        language: currentLanguage,
        searchResultLimit: parsed,
        showAlbumTree: currentShowAlbumTree,
        showHiddenAlbums: currentShowHiddenAlbums,
      },
    })
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
          if (!preferencesLoaded) return

          changePrefs({
            variables: {
              language: language as LanguageTranslation,
              searchResultLimit: currentSearchResultLimit,
              showAlbumTree: currentShowAlbumTree,
              showHiddenAlbums: currentShowHiddenAlbums,
            },
          })
        }}
        selected={data?.myUserPreferences.language || undefined}
        disabled={loadingPrefs || !preferencesLoaded}
      />
      <label htmlFor="user_pref_search_result_limit_field">
        <InputLabelTitle>
          {t(
            'settings.user_preferences.search_result_limit.title',
            'Number of search results'
          )}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.user_preferences.search_result_limit.description',
            'Maximum number of albums and media shown per category in search results. Use 0 to show all results.'
          )}
        </InputLabelDescription>
      </label>
      <TextField
        id="user_pref_search_result_limit_field"
        type="number"
        min={0}
        step={1}
        value={searchResultLimitInput}
        onChange={e => setSearchResultLimitInput(e.target.value)}
        onBlur={commitSearchResultLimit}
        onKeyDown={e => {
          if (e.key === 'Enter') {
            e.preventDefault()
            e.currentTarget.blur()
          }
        }}
        disabled={loadingPrefs || !preferencesLoaded}
        wrapperClassName="mb-4"
      />
      <label htmlFor="user_pref_show_album_tree_field">
        <InputLabelTitle>
          {t(
            'settings.user_preferences.show_album_tree.title',
            'Album tree sidebar'
          )}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.user_preferences.show_album_tree.description',
            'Show a collapsible tree of your albums in a sidebar for quick navigation'
          )}
        </InputLabelDescription>
      </label>
      <Checkbox
        id="user_pref_show_album_tree_field"
        label={t(
          'settings.user_preferences.show_album_tree.checkbox_label',
          'Show album tree sidebar'
        )}
        disabled={loadingPrefs || !preferencesLoaded}
        checked={currentShowAlbumTree}
        onChange={event => {
          if (!preferencesLoaded) return

          changePrefs({
            variables: {
              language: currentLanguage,
              searchResultLimit: currentSearchResultLimit,
              showAlbumTree: event.target.checked,
              showHiddenAlbums: currentShowHiddenAlbums,
            },
          })
        }}
        className="mb-4"
      />
      <label htmlFor="user_pref_show_hidden_albums_field">
        <InputLabelTitle>
          {t(
            'settings.user_preferences.show_hidden_albums.title',
            'Hidden albums'
          )}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.user_preferences.show_hidden_albums.description',
            "Show albums you've hidden, dimmed, so you can bring them back"
          )}
        </InputLabelDescription>
      </label>
      <Checkbox
        id="user_pref_show_hidden_albums_field"
        label={t(
          'settings.user_preferences.show_hidden_albums.checkbox_label',
          'Show hidden albums'
        )}
        disabled={loadingPrefs || !preferencesLoaded}
        checked={currentShowHiddenAlbums}
        onChange={event => {
          if (!preferencesLoaded) return

          changePrefs({
            variables: {
              language: currentLanguage,
              searchResultLimit: currentSearchResultLimit,
              showAlbumTree: currentShowAlbumTree,
              showHiddenAlbums: event.target.checked,
            },
          })
        }}
      />
      <Button
        className="mt-2 mb-4"
        disabled={unhideAllLoading}
        onClick={() => unhideAllAlbums()}
      >
        {t(
          'settings.user_preferences.show_hidden_albums.unhide_all',
          'Unhide all albums'
        )}
      </Button>
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
    </UserPreferencesWrapper>
  )
}

export default UserPreferences
