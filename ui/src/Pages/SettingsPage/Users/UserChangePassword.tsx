import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { Button, Form, Input, Modal, ModalProps } from 'semantic-ui-react'
import { Trans, useTranslation } from 'react-i18next'
import { settingsUsersQuery_user } from './__generated__/settingsUsersQuery'

const changeUserPasswordMutation = gql`
  mutation changeUserPassword($userId: ID!, $password: String!) {
    updateUser(id: $userId, password: $password) {
      id
    }
  }
`

interface ChangePasswordModalProps extends ModalProps {
  onClose(): void
  open: boolean
  user: settingsUsersQuery_user
}

const ChangePasswordModal = ({
  onClose,
  user,
  open,
  ...props
}: ChangePasswordModalProps) => {
  const { t } = useTranslation()
  const [passwordInput, setPasswordInput] = useState('')

  const [changePassword] = useMutation(changeUserPasswordMutation, {
    onCompleted: () => {
      onClose && onClose()
    },
  })

  return (
    <Modal open={open} {...props}>
      <Modal.Header>
        {t('settings.users.password_reset.title', 'Change password')}
      </Modal.Header>
      <Modal.Content>
        <p>
          <Trans t={t} i18nKey="settings.users.password_reset.description">
            Change password for <b>{user.username}</b>
          </Trans>
        </p>
        <Form>
          <Form.Field>
            <label>
              {t('settings.users.password_reset.form.label', 'New password')}
            </label>
            <Input
              placeholder={t(
                'settings.users.password_reset.form.placeholder',
                'password'
              )}
              onChange={e => setPasswordInput(e.target.value)}
              type="password"
            />
          </Form.Field>
        </Form>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => onClose && onClose()}>
          {t('general.action.cancel', 'Cancel')}
        </Button>
        <Button
          positive
          onClick={() => {
            changePassword({
              variables: {
                userId: user.id,
                password: passwordInput,
              },
            })
          }}
        >
          {t('settings.users.password_reset.form.submit', 'Change password')}
        </Button>
      </Modal.Actions>
    </Modal>
  )
}

export default ChangePasswordModal
