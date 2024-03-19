import React, { useState } from 'react'
import { Button, TextField } from '../../../primitives/form/Input'
import { ApolloError, gql, useMutation } from '@apollo/client'
import {
  updatePassword,
  updatePasswordVariables,
} from './__generated__/updatePassword'
import { SectionTitle } from '../SettingsPage'

export const USER_CHANGE_PASSWORD_MUTATION = gql`
  mutation updatePassword($currentPassword: String!, $newPassword: String!) {
    updatePassword(
      currentPassword: $currentPassword
      newPassword: $newPassword
    ) {
      success
      message
    }
  }
`

const initialState = {
  password1: '',
  password2: '',
  currentPassword: '',
}

const errorMessage = (
  data: updatePassword | null | undefined,
  error: ApolloError | undefined
) => {
  return (
    <div>
      {error && <span>Something went wrong</span>}
      {data &&
        data.updatePassword &&
        (data.updatePassword.success ? (
          <span>Successfully updated password</span>
        ) : (
          <span style={{ color: 'red' }}>{data.updatePassword.message}</span>
        ))}
    </div>
  )
}

const PasswordChange = () => {
  const [state, setState] = useState(initialState)

  const [updatePassword, { data, error }] = useMutation<
    updatePassword,
    updatePasswordVariables
  >(USER_CHANGE_PASSWORD_MUTATION, {
    onCompleted: data => {
      if (data?.updatePassword.success) {
        setState(initialState)
      }
    },
  })

  function updateInput(
    event: React.ChangeEvent<HTMLInputElement>,
    key: string
  ) {
    setState({
      ...state,
      [key]: event.target.value,
    })
  }

  return (
    <div>
      {/* TODO Add password confirmation field. */}
      <SectionTitle nospace>Change Password</SectionTitle>
      <TextField
        type="password"
        label="current password"
        value={state.currentPassword}
        onChange={e => updateInput(e, 'currentPassword')}
      />
      <TextField
        type="password"
        label="new password"
        value={state.password1}
        onChange={e => updateInput(e, 'password1')}
      />
      <TextField
        type="password"
        label="confirm password"
        value={state.password2}
        onChange={e => updateInput(e, 'password2')}
      />
      {(state.password1 !== state.password2 && (
        <span style={{ color: 'red' }}>Passwords do not match</span>
      )) ||
        (state.password1 !== '' && (
          <Button
            style={{ marginTop: '5px' }}
            onClick={() =>
              updatePassword({
                variables: {
                  currentPassword: state.currentPassword,
                  newPassword: state.password1,
                },
              })
            }
          >
            Update Password
          </Button>
        ))}
      {errorMessage(data, error)}
    </div>
  )
}

export default PasswordChange
