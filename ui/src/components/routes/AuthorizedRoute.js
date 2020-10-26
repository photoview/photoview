import React from 'react'
import PropTypes from 'prop-types'
import { Route, Redirect } from 'react-router-dom'
import { gql } from '@apollo/client'
import { Query } from '@apollo/client/react/components'
import { authToken } from '../../authentication'

const adminQuery = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

export const Authorized = ({ children }) => {
  const token = authToken()

  return token ? children : null
}

const AuthorizedRoute = ({ component: Component, admin = false, ...props }) => {
  const token = authToken()

  let unauthorizedRedirect = null
  if (!token) {
    unauthorizedRedirect = <Redirect to="/login" />
  }

  let adminRedirect = null
  if (token && admin) {
    adminRedirect = (
      <Query query={adminQuery}>
        {({ error, data }) => {
          if (error) alert(error)

          if (data && data.myUser && !data.myUser.admin) {
            return <Redirect to="/" />
          }

          return null
        }}
      </Query>
    )
  }

  return (
    <Route
      {...props}
      render={routeProps => (
        <>
          {unauthorizedRedirect}
          {adminRedirect}
          <Component {...routeProps} />
        </>
      )}
    ></Route>
  )
}

AuthorizedRoute.propTypes = {
  component: PropTypes.object.isRequired,
  admin: PropTypes.bool,
}

export default AuthorizedRoute
