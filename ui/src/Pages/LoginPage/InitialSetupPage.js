import React, { useState } from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'
import { Redirect } from 'react-router-dom'
import { Button, Form, Message, Header } from 'semantic-ui-react'
import { Container } from './loginUtilities'

import { checkInitialSetupQuery, login } from './loginUtilities'
import { authToken } from '../../helpers/authentication'
import { useTranslation } from 'react-i18next'

const initialSetupMutation = gql`
  mutation InitialSetup(
    $username: String!
    $password: String!
    $rootPath: String!
  ) {
    initialSetupWizard(
      username: $username
      password: $password
      rootPath: $rootPath
    ) {
      success
      status
      token
    }
  }
`

const InitialSetupPage = () => {
  const { t } = useTranslation()

  const [state, setState] = useState({
    username: '',
    password: '',
    rootPath: '',
  })

  const handleChange = (event, key) => {
    const value = event.target.value
    setState(prevState => ({
      ...prevState,
      [key]: value,
    }))
  }

  const signIn = (event, authorize) => {
    event.preventDefault()

    authorize({
      variables: {
        username: state.username,
        password: state.password,
        rootPath: state.rootPath,
      },
    })
  }

  if (authToken()) {
    return <Redirect to="/" />
  }

  const { data: initialSetupData } = useQuery(checkInitialSetupQuery)
  const initialSetupRedirect = initialSetupData?.siteInfo
    ?.initialSetup ? null : (
    <Redirect to="/" />
  )

  const [
    authorize,
    { loading: authorizeLoading, data: authorizationData },
  ] = useMutation(initialSetupMutation, {
    onCompleted: data => {
      const { success, token } = data.initialSetupWizard

      if (success) {
        login(token)
      }
    },
  })

  let errorMessage = null
  if (authorizationData && !authorizationData.initialSetupWizard.success) {
    errorMessage = authorizationData.initialSetupWizard.status
  }

  return (
    <div>
      {initialSetupRedirect}
      <Container>
        <Header as="h1" textAlign="center">
          {t('login_page.initial_setup.title', 'Initial Setup')}
        </Header>
        <Form
          style={{ width: 500, margin: 'auto' }}
          error={!!errorMessage}
          onSubmit={e => signIn(e, authorize)}
          loading={
            authorizeLoading || authorizationData?.initialSetupWizard?.success
          }
        >
          <Form.Field>
            <label>{t('login_page.field.username', 'Username')}</label>
            <input onChange={e => handleChange(e, 'username')} />
          </Form.Field>
          <Form.Field>
            <label>{t('login_page.field.password', 'Password')}</label>
            <input
              type="password"
              onChange={e => handleChange(e, 'password')}
            />
          </Form.Field>
          <Form.Field>
            <label>
              {t(
                'login_page.initial_setup.field.photo_path.label',
                'Photo path'
              )}
            </label>
            <input
              placeholder={t(
                'login_page.initial_setup.field.photo_path.placeholder',
                '/path/to/photos'
              )}
              type="text"
              onChange={e => handleChange(e, 'rootPath')}
            />
          </Form.Field>
          <Message error content={errorMessage} />
          <Button type="submit">
            {t('login_page.initial_setup.field.submit', 'Setup Photoview')}
          </Button>
        </Form>
      </Container>
    </div>
  )
}

export default InitialSetupPage
