import React, { useState } from 'react'
import PropTypes from 'prop-types'
import { gql, useMutation } from '@apollo/client'
import { Button, Form, Input, Modal } from 'semantic-ui-react'

const changeUserPasswordMutation = gql`
  mutation changeUserPassword($userId: ID!, $password: String!) {
    updateUser(id: $userId, password: $password) {
      id
    }
  }
`

const ChangePasswordModal = ({ onClose, user, ...props }) => {
  const [passwordInput, setPasswordInput] = useState('')

  const [changePassword] = useMutation(changeUserPasswordMutation, {
    onCompleted: () => {
      onClose && onClose()
    },
  })

  return (
    <Modal {...props}>
      <Modal.Header>Change password</Modal.Header>
      <Modal.Content>
        <p>
          Change password for <b>{user.username}</b>
        </p>
        <Form>
          <Form.Field>
            <label>New password</label>
            <Input
              placeholder="password"
              onChange={e => setPasswordInput(e.target.value)}
              type="password"
            />
          </Form.Field>
        </Form>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => onClose && onClose()}>Cancel</Button>
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
          Change password
        </Button>
      </Modal.Actions>
    </Modal>
  )
}

ChangePasswordModal.propTypes = {
  onClose: PropTypes.func,
  user: PropTypes.object.isRequired,
}

export default ChangePasswordModal
