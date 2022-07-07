import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { Trans, useTranslation } from 'react-i18next'
import { settingsUsersQuery_user } from './__generated__/settingsUsersQuery'
import Modal from '../../../primitives/Modal'
import { TextField } from '../../../primitives/form/Input'

const changeUserPasswordMutation = gql`
  mutation changeUserPassword($userId: ID!, $password: String!) {
    updateUser(id: $userId, password: $password) {
      id
    }
  }
`

interface ChangePasswordModalProps {
  onClose(): void
  open: boolean
  user: settingsUsersQuery_user
}

const ChangePasswordModal = ({
  onClose,
  user,
  open,
}: ChangePasswordModalProps) => {
  const { t } = useTranslation()
  const [passwordInput, setPasswordInput] = useState('')

  const [changePassword] = useMutation(changeUserPasswordMutation, {
    onCompleted: () => {
      onClose && onClose()
    },
  })

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('settings.users.password_reset.title', 'Change password')}
      description={
        <Trans t={t} i18nKey="settings.users.password_reset.description">
          Change password for <b>{user.username}</b>
        </Trans>
      }
      actions={[
        {
          key: 'cancel',
          label: t('general.action.cancel', 'Cancel'),
          onClick: () => onClose && onClose(),
        },
        {
          key: 'change_password',
          label: t(
            'settings.users.password_reset.form.submit',
            'Change password'
          ),
          variant: 'positive',
          onClick: () => {
            changePassword({
              variables: {
                userId: user.id,
                password: passwordInput,
              },
            })
          },
        },
      ]}
    >
      <div className="w-[360px]">
        <TextField
          label={t('settings.users.password_reset.form.label', 'New password')}
          placeholder={t(
            'settings.users.password_reset.form.placeholder',
            'password'
          )}
          onChange={e => setPasswordInput(e.target.value)}
          type="password"
        />
      </div>
    </Modal>
  )
}

export default ChangePasswordModal
