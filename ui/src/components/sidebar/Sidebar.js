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
    /* full height - header - tabbar */
    height: calc(100% - 60px - 80px);
    max-width: calc(100vw - 85px);
    transform: translateX(100vw);
  }
`

export const SidebarContext = createContext()
SidebarContext.displayName = 'SidebarContext'

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
      <SidebarContext.Provider
        value={{ updateSidebar: this.update, content: this.state.content }}
      >
        {this.props.children}
        <SidebarContext.Consumer>
          {value => (
            <SidebarContainer>
              {value.content}
              <div style={{ height: 100 }}></div>
            </SidebarContainer>
          )}
        </SidebarContext.Consumer>
      </SidebarContext.Provider>
    )
  }
}

Sidebar.propTypes = {
  children: PropTypes.element,
}

export default Sidebar
