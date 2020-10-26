import React, { useEffect } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { NavLink } from 'react-router-dom'
import { Icon } from 'semantic-ui-react'
import Sidebar from './components/sidebar/Sidebar'
import { useQuery, useLazyQuery } from 'react-apollo'
import gql from 'graphql-tag'
import { Authorized } from './components/routes/AuthorizedRoute'
import { Helmet } from 'react-helmet'
import Header from './components/header/Header'
import { authToken } from './authentication'

const ADMIN_QUERY = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

const MAPBOX_QUERY = gql`
  query mapboxEnabledQuery {
    mapboxToken
  }
`

const Container = styled.div`
  height: 100%;
  display: flex;
  overflow: hidden;
  position: relative;
`

const SideMenu = styled.div`
  height: 100%;
  width: 80px;
  left: 0;
  padding-top: 70px;

  @media (max-width: 1000px) {
    width: 100%;
    height: 80px;
    position: fixed;
    background: white;
    z-index: 10;
    padding-top: 0;
    display: flex;
    bottom: 0;
    box-shadow: 0 0 2px rgba(0, 0, 0, 0.3);
  }
`

const Content = styled.div`
  margin-top: 70px;
  padding: 10px 12px 0;
  width: 100%;
  overflow-y: scroll;
`

const SideButtonLink = styled(NavLink)`
  text-align: center;
  padding-top: 8px;
  padding-left: 2px;
  display: block;
  width: 60px;
  height: 60px;
  margin: 10px;
  margin-bottom: 24px;

  font-size: 28px;

  color: #888;

  transition: transform 200ms, box-shadow 200ms;

  :hover {
    transform: scale(1.02);
  }
`

const SideButton = props => {
  return (
    <SideButtonLink {...props} activeStyle={{ color: '#4183c4' }}>
      {props.children}
    </SideButtonLink>
  )
}

SideButton.propTypes = {
  children: PropTypes.any,
}

const SideButtonLabel = styled.div`
  font-size: 16px;
`

const Layout = ({ children, title }) => {
  const [loadAdminQuery, adminQuery] = useLazyQuery(ADMIN_QUERY)
  const mapboxQuery = useQuery(MAPBOX_QUERY)

  useEffect(() => {
    if (authToken()) {
      loadAdminQuery()
    }
  }, [])

  const isAdmin =
    adminQuery.data && adminQuery.data.myUser && adminQuery.data.myUser.admin

  const mapboxEnabled = mapboxQuery.data && mapboxQuery.data.mapboxToken != null

  return (
    <Container>
      <Helmet>
        <title>{title ? `${title} - Photoview` : `Photoview`}</title>
      </Helmet>
      <Authorized>
        <SideMenu>
          <SideButton to="/photos" exact>
            <Icon name="image outline" />
            <SideButtonLabel>Photos</SideButtonLabel>
          </SideButton>
          <SideButton to="/albums" exact>
            <Icon name="images outline" />
            <SideButtonLabel>Albums</SideButtonLabel>
          </SideButton>
          {mapboxEnabled ? (
            <SideButton to="/places" exact>
              <Icon name="map outline" />
              <SideButtonLabel>Places</SideButtonLabel>
            </SideButton>
          ) : null}
          {isAdmin ? (
            <SideButton to="/settings" exact>
              <Icon name="settings" />
              <SideButtonLabel>Settings</SideButtonLabel>
            </SideButton>
          ) : null}
          <SideButton to="/logout">
            <Icon name="lock" />
            <SideButtonLabel>Log out</SideButtonLabel>
          </SideButton>
        </SideMenu>
      </Authorized>
      <Sidebar>
        <Content id="layout-content">
          {children}
          <div style={{ height: 24 }}></div>
        </Content>
      </Sidebar>
      <Header />
    </Container>
  )
}

Layout.propTypes = {
  children: PropTypes.any.isRequired,
  title: PropTypes.string,
}

export default Layout
