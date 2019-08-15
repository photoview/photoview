import React, { createContext } from 'react'
import styled from 'styled-components'
import { Route } from 'react-router-dom'

const SidebarContainer = styled.div`
  height: 100%;
  width: 500px;
  position: fixed;
  overflow-y: scroll;
  right: 0;
  top: 60px;
  background-color: white;
  padding: 12px 12px 100px;
  border-left: 1px solid #eee;
`

export const SidebarContext = createContext()

export const SidebarConsumer = SidebarContext.Consumer

const { Consumer, Provider } = SidebarContext

class Sidebar extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      content: 'Start value',
    }

    this.update = content => this.setState({ content })
  }

  render() {
    return (
      <Provider
        value={{ updateSidebar: this.update, content: this.state.content }}
      >
        {this.props.children}
        <Consumer>
          {value => <SidebarContainer>{value.content}</SidebarContainer>}
        </Consumer>
      </Provider>
    )
  }
}

export default Sidebar
