import React, { createContext } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'

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
      content: null,
    }

    this.update = content => {
      this.setState({ content })
    }
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

Sidebar.propTypes = {
  children: PropTypes.element,
}

export default Sidebar
