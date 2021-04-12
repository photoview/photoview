import React, { ReactChild, useEffect } from 'react'
import PropTypes, { ReactComponentLike } from 'prop-types'
import { Route, Redirect } from 'react-router-dom'
import { useLazyQuery } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { ADMIN_QUERY } from '../../Layout'

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

export const Authorized = ({ children }: { children: JSX.Element }) => {
  const token = authToken()

  return token ? children : null
}

type AuthorizedRouteProps = {
  component: ReactComponentLike
  admin: boolean
}

const AuthorizedRoute = ({
  component: Component,
  admin = false,
  ...props
}: AuthorizedRouteProps) => {
  const token = authToken()
  const isAdmin = useIsAdmin(admin)

  let unauthorizedRedirect: null | ReactChild = null
  if (!token) {
    unauthorizedRedirect = <Redirect to="/login" />
  }

  let adminRedirect: null | ReactChild = null
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
