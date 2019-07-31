import gql from 'graphql-tag'

export const checkInitialSetupQuery = gql`
  query CheckInitialSetup {
    siteInfo {
      initialSetup
    }
  }
`

export function setCookie(cname, cvalue, exdays) {
  var d = new Date()
  d.setTime(d.getTime() + exdays * 24 * 60 * 60 * 1000)
  var expires = 'expires=' + d.toUTCString()
  document.cookie = cname + '=' + cvalue + ';' + expires + ';path=/'
}

export function login(token) {
  localStorage.setItem('token', token)
  setCookie('token', token, 360)
  window.location = '/'
}
