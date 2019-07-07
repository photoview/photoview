import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { Redirect } from 'react-router-dom'

const authorizeMutation = gql`
  mutation Authorize($username: String!, $password: String!) {
    authorizeUser(username: $username, password: $password) {
      success
      status
      token
    }
  }
`

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
        <h1>Welcome</h1>
        <Mutation
          mutation={authorizeMutation}
          onCompleted={data => {
            const { success, token } = data.authorizeUser

            if (success) {
              localStorage.setItem('token', token)
              window.location = '/'
            }
          }}
        >
          {(authorize, { loading, error, data }) => {
            let signInBtn = <input type="submit" value="Sign in" />

            if (loading) signInBtn = <span>Signing in...</span>

            let status = ''
            if (data) {
              if (!data.authorizeUser.success)
                status = data.authorizeUser.status
            }

            return (
              <form onSubmit={e => this.signIn(e, authorize)}>
                <label htmlFor="username-field">Username:</label>
                <input
                  id="username-field"
                  onChange={e => this.handleChange(e, 'username')}
                />
                <br />
                <label htmlFor="password-field">Password:</label>
                <input
                  id="password-field"
                  type="password"
                  onChange={e => this.handleChange(e, 'password')}
                />
                <br />
                {signInBtn}
                <br />
                <span>{status}</span>
              </form>
            )
          }}
        </Mutation>
      </div>
    )
  }
}

export default LoginPage
