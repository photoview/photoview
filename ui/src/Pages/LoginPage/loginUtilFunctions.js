import gql from 'graphql-tag'

export const checkInitialSetupQuery = gql`
  query CheckInitialSetup {
    siteInfo {
      initialSetup
    }
  }
`

export function login(token) {
  localStorage.setItem('token', token)
  window.location = '/'
}
