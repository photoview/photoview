import { gql } from '@apollo/client'
import { saveTokenCookie } from '../../helpers/authentication'
import styled from 'styled-components'

export const INITIAL_SETUP_QUERY = gql`
  query CheckInitialSetup {
    siteInfo {
      initialSetup
    }
  }
`

export function login(token: string) {
  saveTokenCookie(token)
  window.location.href = '/'
}

export const Container = styled.div.attrs({ className: 'mt-20' })``
