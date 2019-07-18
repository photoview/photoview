import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { Redirect } from 'react-router-dom'
import { Button, Form, Message, Container, Header } from 'semantic-ui-react'

const authorizeMutation = gql`
  mutation Authorize($username: String!, $password: String!) {
    authorizeUser(username: $username, password: $password) {
      success
      status
      token
    }
  }
`

function setCookie(cname, cvalue, exdays) {
  var d = new Date()
  d.setTime(d.getTime() + exdays * 24 * 60 * 60 * 1000)
  var expires = 'expires=' + d.toUTCString()
  document.cookie = cname + '=' + cvalue + ';' + expires + ';path=/'
}

class LoginPage extends Component {
  constructor(props) {
    super(props)

    this.state = {
      username: '',
      password: '',
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
      },
    })
  }

  render() {
    if (localStorage.getItem('token')) {
      return <Redirect to="/" />
    }

    return (
      <div>
        <Container>
          <Header as="h1" textAlign="center">
            Welcome
          </Header>
          <Mutation
            mutation={authorizeMutation}
            onCompleted={data => {
              const { success, token } = data.authorizeUser

              if (success) {
                localStorage.setItem('token', token)
                setCookie('token', token, 360)
                window.location = '/'
              }
            }}
          >
            {(authorize, { loading, error, data }) => {
              let errorMessage = null
              if (data) {
                if (!data.authorizeUser.success)
                  errorMessage = data.authorizeUser.status
              }

              return (
                <Form
                  style={{ width: 500, margin: 'auto' }}
                  error={!!errorMessage}
                  onSubmit={e => this.signIn(e, authorize)}
                  loading={loading || (data && data.authorizeUser.success)}
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
                  <Message error content={errorMessage} />
                  <Button type="submit">Sign in</Button>
                </Form>
              )
            }}
          </Mutation>
        </Container>
      </div>
    )
  }
}

export default LoginPage
