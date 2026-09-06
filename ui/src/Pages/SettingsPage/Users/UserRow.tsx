import React, { useState } from 'react'
import {
  FetchResult,
  gql,
  MutationFunctionOptions,
  useMutation,
} from '@apollo/client'
import EditUserRow from './EditUserRow'
import ViewUserRow from './ViewUserRow'
import { settingsUsersQuery_user } from './__generated__/settingsUsersQuery'
import { scanUser, scanUserVariables } from './__generated__/scanUser'
import { updateUser, updateUserVariables } from './__generated__/updateUser'
import { deleteUser, deleteUserVariables } from './__generated__/deleteUser'

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

interface UserRowState extends settingsUsersQuery_user {
  editing: boolean
  newRootPath: string
  oldState?: Omit<UserRowState, 'oldState'>
}

type ApolloMutationFn<MutationType, VariablesType> = (
  options?: MutationFunctionOptions<MutationType, VariablesType>
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
) => Promise<FetchResult<MutationType, any, any>>

export type UserRowChildProps = {
  user: settingsUsersQuery_user
  state: UserRowState
  setState: React.Dispatch<React.SetStateAction<UserRowState>>
  scanUser: ApolloMutationFn<scanUser, scanUserVariables>
  updateUser: ApolloMutationFn<updateUser, updateUserVariables>
  updateUserLoading: boolean
  deleteUser: ApolloMutationFn<deleteUser, deleteUserVariables>
  setChangePassword: React.Dispatch<React.SetStateAction<boolean>>
  setConfirmDelete: React.Dispatch<React.SetStateAction<boolean>>
  scanUserCalled: boolean
  showChangePassword: boolean
  showConfirmDelete: boolean
}

export type UserRowProps = {
  user: settingsUsersQuery_user
  refetchUsers: () => void
}

const UserRow = ({ user, refetchUsers }: UserRowProps) => {
  const [state, setState] = useState<UserRowState>({
    ...user,
    editing: false,
    newRootPath: '',
  })

  const [showConfirmDelete, setConfirmDelete] = useState(false)
  const [showChangePassword, setChangePassword] = useState(false)

  const [updateUser, { loading: updateUserLoading }] = useMutation<
    updateUser,
    updateUserVariables
  >(updateUserMutation, {
    onCompleted: data => {
      setState(state => ({
        ...state,
        ...data.updateUser,
        editing: false,
      }))
      refetchUsers()
    },
  })

  const [deleteUser] = useMutation<deleteUser, deleteUserVariables>(
    deleteUserMutation,
    {
      onCompleted: () => {
        refetchUsers()
      },
    }
  )

  const [scanUser, { called: scanUserCalled }] = useMutation<
    scanUser,
    scanUserVariables
  >(scanUserMutation, {
    onCompleted: () => {
      refetchUsers()
    },
  })

  const props: UserRowChildProps = {
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

export default UserRow
