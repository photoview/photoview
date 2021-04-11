import React, { useEffect } from 'react'
import PropTypes from 'prop-types'
import { Route, Redirect } from 'react-router-dom'
import { gql, useLazyQuery } from '@apollo/client'
import { authToken } from '../../helpers/authentication'

export const ADMIN_QUERY = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

export const useIsAdmin = (enabled = true) => {
  const [fetchAdminQuery, { data }] = useLazyQuery(ADMIN_QUERY)

  useEffect(() => {
    if (authToken() && !data && enabled) {
      fetchAdminQuery()
    }
  }, [authToken(), enabled])

  if (!authToken()) {
    return false
  }

  return data?.myUser?.admin
}

export const Authorized = ({ children }) => {
  const token = authToken()

  return token ? children : null
}

const AuthorizedRoute = ({ component: Component, admin = false, ...props }) => {
  const token = authToken()
  const isAdmin = useIsAdmin(admin)

  let unauthorizedRedirect = null
  if (!token) {
    unauthorizedRedirect = <Redirect to="/login" />
  }

  let adminRedirect = null
  if (token && admin) {
    if (isAdmin === false) {
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
