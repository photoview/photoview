/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ShareTokenValidatePassword
// ====================================================

export interface ShareTokenValidatePassword {
  /**
   * Check if the `ShareToken` credentials are valid
   */
  shareTokenValidatePassword: boolean
}

export interface ShareTokenValidatePasswordVariables {
  token: string
  password?: string | null
}
