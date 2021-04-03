import React from 'react'
import PropTypes from 'prop-types'
import { Route, Redirect } from 'react-router-dom'
import { useQuery, gql } from '@apollo/client'
import { authToken } from '../../helpers/authentication'

export const ADMIN_QUERY = gql`
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
    const { error, data } = useQuery(ADMIN_QUERY)

    if (error) alert(error)

    if (data && data.myUser && !data.myUser.admin) {
      adminRedirect = <Redirect to="/" />
    }
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
    />
  )
}

AuthorizedRoute.propTypes = {
  component: PropTypes.elementType.isRequired,
  admin: PropTypes.bool,
}

export default AuthorizedRoute
