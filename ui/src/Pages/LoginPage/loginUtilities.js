import { gql } from '@apollo/client'
import { saveTokenCookie } from '../../helpers/authentication'
import styled from 'styled-components'
import { Container as SemanticContainer } from 'semantic-ui-react'

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

export const Container = styled(SemanticContainer)`
  margin-top: 80px;
`
