import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { NavLink } from 'react-router-dom'
import { Icon } from 'semantic-ui-react'
import Sidebar from './components/sidebar/Sidebar'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import { Authorized } from './AuthorizedRoute'
import { Helmet } from 'react-helmet'
import Header from './components/header/Header'

const adminQuery = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

const Container = styled.div`
  height: 100%;
  display: flex;
  /* margin-right: 500px; */
  /* display: grid;
  grid-template-columns: 80px 1fr 500px; */
`

const SideMenu = styled.div`
  height: 100%;
  width: 80px;
  left: 0;
  padding-top: 70px;
`

const Content = styled.div`
  margin-top: 60px;
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

const Layout = ({ children, title }) => (
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
        <Query query={adminQuery}>
          {({ loading, error, data }) => {
            if (data && data.myUser && data.myUser.admin) {
              return (
                <SideButton to="/settings" exact>
                  <Icon name="settings" />
                  <SideButtonLabel>Settings</SideButtonLabel>
                </SideButton>
              )
            }

            return null
          }}
        </Query>
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

Layout.propTypes = {
  children: PropTypes.any.isRequired,
  title: PropTypes.string,
}

export default Layout
