import React from 'react'
import { Route, Redirect } from 'react-router-dom'

const AuthorizedRoute = ({ component: Component, ...props }) => {
  const token = localStorage.getItem('token')

  let unauthorizedRedirect = null
  if (!token) {
    unauthorizedRedirect = <Redirect to="/login" />
  }

  return (
    <Route
      {...props}
      render={routeProps => (
        <>
          {unauthorizedRedirect}
          <Component {...routeProps} />
        </>
      )}
    ></Route>
  )
}

export default AuthorizedRoute
