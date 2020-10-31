import React, { useState, useCallback } from 'react'
import { useQuery, gql, useMutation } from '@apollo/client'
import { Redirect } from 'react-router-dom'
import styled from 'styled-components'
import { Button, Form, Message, Header } from 'semantic-ui-react'
import { checkInitialSetupQuery, login, Container } from './loginUtilities'
import { authToken } from '../../authentication'

import logoPath from '../../assets/photoview-logo.svg'

const authorizeMutation = gql`
  mutation Authorize($username: String!, $password: String!) {
    authorizeUser(username: $username, password: $password) {
      success
      status
      token
    }
  }
`

const StyledLogo = styled.img`
  max-height: 128px;
`

const LogoHeader = props => (
  <Header {...props} as="h1" textAlign="center">
    <StyledLogo src={logoPath} alt="photoview logo" />
    <p style={{ fontWeight: 400 }}>Welcome to Photoview</p>
  </Header>
)

const LogoHeaderStyled = styled(LogoHeader)`
  margin-bottom: 72px !important;
`

const LoginPage = () => {
  const [credentials, setCredentials] = useState({
    username: '',
    password: '',
  })

  const handleChange = useCallback(
    (event, key) => {
      const value = event.target.value
      setCredentials(credentials => {
        return {
          ...credentials,
          [key]: value,
        }
      })
    },
    [setCredentials]
  )

  const signIn = useCallback(
    (event, authorize) => {
      event.preventDefault()

      authorize({
        variables: {
          username: credentials.username,
          password: credentials.password,
        },
      })
    },
    [credentials]
  )

  const { initialSetupData } = useQuery(checkInitialSetupQuery)

  const [authorize, { loading, data }] = useMutation(authorizeMutation, {
    onCompleted: data => {
      const { success, token } = data.authorizeUser

      if (success) {
        login(token)
      }
    },
  })

  const errorMessage =
    data && !data.authorizeUser.success ? data.authorizeUser.status : null

  if (authToken()) {
    return <Redirect to="/" />
  }

  return (
    <div>
      <Container>
        <LogoHeaderStyled />
        {initialSetupData?.siteInfo?.initialSetup && (
          <Redirect to="/initialSetup" />
        )}
        <Form
          style={{ width: 500, margin: 'auto' }}
          error={!!errorMessage}
          onSubmit={e => signIn(e, authorize)}
          loading={loading || (data && data.authorizeUser.success)}
        >
          <Form.Field>
            <label htmlFor="username_field">Username</label>
            <input
              id="username_field"
              onChange={e => handleChange(e, 'username')}
            />
          </Form.Field>
          <Form.Field>
            <label htmlFor="password_field">Password</label>
            <input
              type="password"
              id="password_field"
              onChange={e => handleChange(e, 'password')}
            />
          </Form.Field>
          <Message error content={errorMessage} />
          <Button type="submit">Sign in</Button>
        </Form>
      </Container>
    </div>
  )
}

export default LoginPage
