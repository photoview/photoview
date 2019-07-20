import React, { Component } from 'react'
import styled from 'styled-components'
import { NavLink } from 'react-router-dom'
import { Icon } from 'semantic-ui-react'

const Container = styled.div`
  height: 100%;
  margin-right: 500px;
  /* display: grid;
  grid-template-columns: 80px 1fr 500px; */
`

const LeftSidebar = styled.div`
  height: 100%;
  width: 80px;
  position: fixed;
  top: 60px;
  left: 0;
  padding-top: 10px;
`

const Content = styled.div`
  margin-top: 60px;
  margin-left: 80px;
  padding: 12px 8px 0;
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
        <LeftSidebar>
          <SideButton to="/photos" exact>
            <Icon name="image outline" />
            <SideButtonLabel>Photos</SideButtonLabel>
          </SideButton>
          <SideButton to="/albums" exact>
            <Icon name="images outline" />
            <SideButtonLabel>Albums</SideButtonLabel>
          </SideButton>
        </LeftSidebar>
        <Content>{this.props.children}</Content>
        <Header>
          <Title>Photoview</Title>
        </Header>
      </Container>
    )
  }
}

export default Layout
