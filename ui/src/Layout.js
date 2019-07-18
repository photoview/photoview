import React, { Component } from 'react'
import styled from 'styled-components'

const Container = styled.div`
  height: 100%;
  /* display: grid;
  grid-template-columns: 80px 1fr 500px; */
`

const LeftSidebar = styled.div`
  height: 100%;
  width: 80px;
  position: fixed;
  top: 0;
  left: 0;
  background-color: #eee;
`

const RightSidebar = styled.div`
  height: 100%;
  width: 500px;
  position: fixed;
  right: 0;
  top: 0;
  background-color: #eee;
`

const Content = styled.div`
  margin-left: 80px;
  margin-right: 500px;
  padding: 0 8px;
`

class Layout extends Component {
  render() {
    return (
      <Container>
        <LeftSidebar>Left</LeftSidebar>
        <Content>{this.props.children}</Content>
        <RightSidebar>Right</RightSidebar>
      </Container>
    )
  }
}

export default Layout
