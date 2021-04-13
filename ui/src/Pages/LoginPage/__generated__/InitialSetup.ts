/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: InitialSetup
// ====================================================

export interface InitialSetup_initialSetupWizard {
  __typename: "AuthorizeResult";
  success: boolean;
  status: string;
  token: string | null;
}

export interface InitialSetup {
  /**
   * Registers the initial user, can only be called if initialSetup from SiteInfo is true
   */
  initialSetupWizard: InitialSetup_initialSetupWizard | null;
}

export interface InitialSetupVariables {
  username: string;
  password: string;
  rootPath: string;
}
