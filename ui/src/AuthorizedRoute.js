import React from 'react'
import PropTypes from 'prop-types'
import { Route, Redirect } from 'react-router-dom'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'

const adminQuery = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

const AuthorizedRoute = ({ component: Component, admin = false, ...props }) => {
  const token = localStorage.getItem('token')

  let unauthorizedRedirect = null
  if (!token) {
    unauthorizedRedirect = <Redirect to="/login" />
  }

  let adminRedirect = null
  if (token && admin) {
    adminRedirect = (
      <Query query={adminQuery}>
        {({ loading, error, data }) => {
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
