import React, { Component } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { NavLink } from 'react-router-dom'
import { Icon } from 'semantic-ui-react'
import Sidebar from './components/sidebar/Sidebar'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import { Authorized } from './AuthorizedRoute'

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

const SideButton = props => {
  const StyledLink = styled(NavLink)`
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

  return (
    <StyledLink {...props} activeStyle={{ color: '#4183c4' }}>
      {props.children}
    </StyledLink>
  )
}

SideButton.propTypes = {
  children: PropTypes.any,
}

const SideButtonLabel = styled.div`
  font-size: 16px;
`

const Header = styled.div`
  height: 60px;
  width: 100%;
  position: fixed;
  background: white;
  top: 0;
  /* border-bottom: 1px solid rgba(0, 0, 0, 0.1); */
  box-shadow: 0 0 2px rgba(0, 0, 0, 0.3);
`

const Title = styled.h1`
  font-size: 36px;
  padding: 5px 12px;
`

class Layout extends Component {
  render() {
    return (
      <Container>
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
          </SideMenu>
        </Authorized>
        <Sidebar>
          <Content id="layout-content">
            {this.props.children}
            <div style={{ height: 24 }}></div>
          </Content>
        </Sidebar>
        <Header>
          <Title>Photoview</Title>
        </Header>
      </Container>
    )
  }
}

Layout.propTypes = {
  children: PropTypes.any.isRequired,
}

export default Layout
