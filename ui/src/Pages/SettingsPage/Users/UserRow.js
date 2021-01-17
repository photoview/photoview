import PropTypes from 'prop-types'
import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import EditUserRow from './EditUserRow'
import ViewUserRow from './ViewUserRow'

const updateUserMutation = gql`
  mutation updateUser($id: ID!, $username: String, $admin: Boolean) {
    updateUser(id: $id, username: $username, admin: $admin) {
      id
      username
      admin
    }
  }
`

const deleteUserMutation = gql`
  mutation deleteUser($id: ID!) {
    deleteUser(id: $id) {
      id
      username
    }
  }
`

const scanUserMutation = gql`
  mutation scanUser($userId: ID!) {
    scanUser(userId: $userId) {
      success
    }
  }
`

const UserRow = ({ user, refetchUsers }) => {
  const [state, setState] = useState({
    ...user,
    editing: false,
    newRootPath: '',
  })

  const [showConfirmDelete, setConfirmDelete] = useState(false)
  const [showChangePassword, setChangePassword] = useState(false)

  const [updateUser, { loading: updateUserLoading }] = useMutation(
    updateUserMutation,
    {
      onCompleted: data => {
        setState({
          ...data.updateUser,
          editing: false,
        })
        refetchUsers()
      },
    }
  )

  const [deleteUser] = useMutation(deleteUserMutation, {
    onCompleted: () => {
      refetchUsers()
    },
  })

  const [scanUser, { called: scanUserCalled }] = useMutation(scanUserMutation, {
    onCompleted: () => {
      refetchUsers()
    },
  })

  const props = {
    user,
    state,
    setState,
    scanUser,
    updateUser,
    updateUserLoading,
    deleteUser,
    setChangePassword,
    setConfirmDelete,
    scanUserCalled,
    showChangePassword,
    showConfirmDelete,
  }

  if (state.editing) {
    return <EditUserRow {...props} />
  }

  return <ViewUserRow {...props} />
}

UserRow.propTypes = {
  user: PropTypes.object.isRequired,
  refetchUsers: PropTypes.func.isRequired,
}

export const UserRowProps = {
  user: PropTypes.object.isRequired,
  state: PropTypes.object.isRequired,
  setState: PropTypes.func.isRequired,
  scanUser: PropTypes.func.isRequired,
  updateUser: PropTypes.func.isRequired,
  updateUserLoading: PropTypes.bool.isRequired,
  deleteUser: PropTypes.func.isRequired,
  setChangePassword: PropTypes.func.isRequired,
  setConfirmDelete: PropTypes.func.isRequired,
  scanUserCalled: PropTypes.func.isRequired,
  showChangePassword: PropTypes.func.isRequired,
  showConfirmDelete: PropTypes.func.isRequired,
}

export default UserRow
