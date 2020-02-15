import React, { createContext } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'

const SidebarContainer = styled.div`
  width: 28vw;
  max-width: 500px;
  min-width: 300px;
  flex-shrink: 0;
  overflow-y: scroll;
  right: 0;
  margin-top: 60px;
  background-color: white;
  padding: 12px;
  border-left: 1px solid #eee;

  @media (max-width: 700px) {
    position: absolute;
    width: 100%;
    max-width: calc(100vw - 85px);
    transform: translateX(100vw);
  }
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
          {value => (
            <SidebarContainer>
              {value.content}
              <div style={{ height: 100 }}></div>
            </SidebarContainer>
          )}
        </Consumer>
      </Provider>
    )
  }
}

Sidebar.propTypes = {
  children: PropTypes.element,
}

export default Sidebar
