import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { InMemoryCache } from '@apollo/client'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import UserPreferences, {
  CHANGE_USER_PREFERENCES,
  MY_USER_PREFERENCES,
  MY_USERNAME_QUERY,
} from './UserPreferences'

const savedPreferences = {
  request: { query: MY_USER_PREFERENCES },
  result: {
    data: {
      myUserPreferences: {
        id: '1',
        language: null,
        searchResultLimit: 25,
      },
    },
  },
}

const username = {
  request: { query: MY_USERNAME_QUERY },
  result: { data: { myUser: { id: '1', username: 'someone' } } },
}

test('emptying the limit field clears the preference', async () => {
  let cleared = false

  const clearMutation = {
    request: {
      query: CHANGE_USER_PREFERENCES,
      variables: { searchResultLimit: -1 },
    },
    result: () => {
      cleared = true

      return {
        data: {
          changeUserPreferences: {
            id: '1',
            language: null,
            searchResultLimit: null,
          },
        },
      }
    },
  }

  render(
    <MockedProvider
      mocks={[savedPreferences, username, clearMutation]}
      cache={new InMemoryCache({ addTypename: false })}
    >
      <UserPreferences />
    </MockedProvider>
  )

  const field = await screen.findByRole('spinbutton')
  await waitFor(() => expect(field).toHaveValue(25))

  await userEvent.clear(field)
  // The field commits on blur, the same as typing a number does.
  await userEvent.tab()

  // 0 already means unlimited, so clearing has to say so some other way -
  // without it there is no route back to the server default.
  await waitFor(() => expect(cleared).toBe(true))
})
