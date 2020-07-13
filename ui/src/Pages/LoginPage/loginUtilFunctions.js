import gql from 'graphql-tag'
import { saveTokenCookie } from '../../authentication'

export const checkInitialSetupQuery = gql`
  query CheckInitialSetup {
    siteInfo {
      initialSetup
    }
  }
`

export function login(token) {
  saveTokenCookie(token)
  window.location = '/'
}
