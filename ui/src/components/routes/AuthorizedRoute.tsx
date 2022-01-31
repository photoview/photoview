import React, { useEffect } from 'react'
import { RouteProps, Navigate } from 'react-router-dom'
import { useLazyQuery } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { ADMIN_QUERY } from '../layout/Layout'

export const useIsAdmin = () => {
  const [fetchAdminQuery, { data, called }] = useLazyQuery(ADMIN_QUERY)

  useEffect(() => {
    if (authToken() && !called) {
      fetchAdminQuery()
    }
  }, [authToken()])

  if (!authToken()) {
    return false
  }

  return data?.myUser?.admin
}

export const Authorized = ({ children }: { children: JSX.Element }) => {
  const token = authToken()

  return token ? children : null
}

interface AuthorizedRouteProps extends Omit<RouteProps, 'component'> {
  children: React.ReactNode
  admin?: boolean
}

const AuthorizedRoute = ({ admin = false, children }: AuthorizedRouteProps) => {
  const token = authToken()
  const isAdmin = useIsAdmin()

  if (!token) {
    return <Navigate to="/" />
  }

  if (admin && !isAdmin) {
    return <Navigate to="/" />
  }

  return <>{children}</>
}

export default AuthorizedRoute
