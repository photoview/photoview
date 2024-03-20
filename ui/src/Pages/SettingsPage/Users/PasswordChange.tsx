import React, { useState } from 'react'
import { Button, TextField } from '../../../primitives/form/Input'
import { gql, useMutation } from '@apollo/client'
import {
  updatePassword,
  updatePasswordVariables,
} from './__generated__/updatePassword'
import { SectionTitle } from '../SettingsPage'
import { useTranslation } from 'react-i18next'
import MessageBox from '../../../primitives/form/MessageBox'

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

const PasswordChange = () => {
  const [state, setState] = useState(initialState)
  const { t } = useTranslation()

  const [updatePassword, { data, reset }] = useMutation<
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
    reset()
    setState({
      ...state,
      [key]: event.target.value,
    })
  }

  const newPasswordsMatch = state.password1 === state.password2

  function statusMessages() {
    let type: 'negative' | 'neutral' | 'positive' = 'negative'
    let message = t(
      'settings.user.password_reset.form.mismatch',
      'Passwords do not match'
    )
    const dataExists = !!data?.updatePassword

    if (dataExists && data.updatePassword.success) {
      type = 'positive'
      message = t(
        'settings.user.password_reset.form.success',
        'Password changed successfully'
      )
    }

    if (
      dataExists &&
      !data.updatePassword.success &&
      data.updatePassword.message
    ) {
      message = data.updatePassword.message
    }

    return (
      <MessageBox
        type={type}
        message={message}
        show={dataExists || !newPasswordsMatch}
      />
    )
  }

  return (
    <div>
      <SectionTitle nospace>
        {t('settings.user.password_reset.title', 'Change password')}
      </SectionTitle>
      <TextField
        type="password"
        label={t(
          'settings.user.password_reset.form.current_password',
          'Current password'
        )}
        value={state.currentPassword}
        onChange={e => updateInput(e, 'currentPassword')}
      />
      <TextField
        type="password"
        label={t(
          'settings.user.password_reset.form.new_password',
          'New password'
        )}
        value={state.password1}
        onChange={e => updateInput(e, 'password1')}
      />
      <TextField
        type="password"
        label={t(
          'settings.user.password_reset.form.confirm_password',
          'Confirm password'
        )}
        value={state.password2}
        onChange={e => updateInput(e, 'password2')}
      />
      {statusMessages()}
      {state.password1 !== '' && newPasswordsMatch && (
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
          {t('settings.user.password_reset.form.submit', 'Change Password')}
        </Button>
      )}
    </div>
  )
}

export default PasswordChange
