import { useLazyQuery } from '@apollo/client'
import React, { useEffect } from 'react'
import { Navigate } from 'react-router-dom'
import { authToken } from '../../helpers/authentication'
import { adminQuery } from '../../__generated__/adminQuery'
import { ADMIN_QUERY } from '../layout/Layout'

export const useIsAdmin = () => {
  const [fetchAdminQuery, { data, called }] =
    useLazyQuery<adminQuery>(ADMIN_QUERY)

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

interface AuthorizedRouteProps {
  children: React.ReactNode
}

const AuthorizedRoute = ({ children }: AuthorizedRouteProps) => {
  const token = authToken()

  if (!token) {
    return <Navigate to="/" />
  }

  return <>{children}</>
}

export default AuthorizedRoute
