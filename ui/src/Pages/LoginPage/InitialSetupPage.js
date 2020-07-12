import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Mutation, Query } from 'react-apollo'
import { Redirect } from 'react-router-dom'
import { Button, Form, Message, Header } from 'semantic-ui-react'
import { Container } from './LoginPage'

import { checkInitialSetupQuery, login } from './loginUtilFunctions'
import { authToken } from '../../authentication'

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

class InitialSetupPage extends Component {
  constructor(props) {
    super(props)

    this.state = {
      username: '',
      password: '',
      rootPath: '',
    }
  }

  handleChange(event, key) {
    this.setState({ [key]: event.target.value })
  }

  signIn(event, authorize) {
    event.preventDefault()

    authorize({
      variables: {
        username: this.state.username,
        password: this.state.password,
        rootPath: this.state.rootPath,
      },
    })
  }

  render() {
    if (authToken()) {
      return <Redirect to="/" />
    }

    return (
      <div>
        <Container>
          <Header as="h1" textAlign="center">
            Initial Setup
          </Header>
          <Query query={checkInitialSetupQuery}>
            {({ loading, error, data }) => {
              if (data && data.siteInfo && data.siteInfo.initialSetup) {
                return null
              }

              return <Redirect to="/" />
            }}
          </Query>
          <Mutation
            mutation={initialSetupMutation}
            onCompleted={data => {
              const { success, token } = data.initialSetupWizard

              if (success) {
                login(token)
              }
            }}
          >
            {(authorize, { loading, error, data }) => {
              let errorMessage = null
              if (data) {
                if (!data.initialSetupWizard.success)
                  errorMessage = data.initialSetupWizard.status
              }

              return (
                <Form
                  style={{ width: 500, margin: 'auto' }}
                  error={!!errorMessage}
                  onSubmit={e => this.signIn(e, authorize)}
                  loading={loading || (data && data.initialSetupWizard.success)}
                >
                  <Form.Field>
                    <label>Username</label>
                    <input onChange={e => this.handleChange(e, 'username')} />
                  </Form.Field>
                  <Form.Field>
                    <label>Password</label>
                    <input
                      type="password"
                      onChange={e => this.handleChange(e, 'password')}
                    />
                  </Form.Field>
                  <Form.Field>
                    <label>Photo Path</label>
                    <input
                      placeholder="/path/to/photos"
                      type="text"
                      onChange={e => this.handleChange(e, 'rootPath')}
                    />
                  </Form.Field>
                  <Message error content={errorMessage} />
                  <Button type="submit">Setup Photoview</Button>
                </Form>
              )
            }}
          </Mutation>
        </Container>
      </div>
    )
  }
}

export default InitialSetupPage
