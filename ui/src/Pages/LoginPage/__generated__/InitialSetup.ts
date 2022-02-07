/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: InitialSetup
// ====================================================

export interface InitialSetup_initialSetupWizard {
  __typename: 'AuthorizeResult'
  success: boolean
  /**
   * A textual status message describing the result, can be used to show an error message when `success` is false
   */
  status: string
  /**
   * An access token used to authenticate new API requests as the newly authorized user. Is present when success is true
   */
  token: string | null
}

export interface InitialSetup {
  /**
   * Registers the initial user, can only be called if initialSetup from SiteInfo is true
   */
  initialSetupWizard: InitialSetup_initialSetupWizard | null
}

export interface InitialSetupVariables {
  username: string
  password: string
  rootPath: string
}
