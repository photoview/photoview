/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: updatePassword
// ====================================================

export interface updatePassword_updatePassword {
  __typename: 'PasswordChangeResult'
  success: boolean
  message: string | null
}

export interface updatePassword {
  updatePassword: updatePassword_updatePassword
}

export interface updatePasswordVariables {
  currentPassword: string
  newPassword: string
}
