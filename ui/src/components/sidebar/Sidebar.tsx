import React, { createContext, useState } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Icon } from 'semantic-ui-react'

const SidebarContainer = styled.div<{ highlighted: boolean }>`
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
    max-width: min(calc(100vw - 85px), 400px);
    ${({ highlighted }) => `right: ${highlighted ? 0 : -100}%;`}
    padding-top: 45px;
  }

  transition: right 200ms ease-in-out;
`

const SidebarDismissButton = styled(Icon)`
  position: absolute;
  top: 10px;
  right: 10px;

  @media (min-width: 700px) {
    display: none;
  }
`

export type UpdateSidebarFn = (content: React.ReactNode) => void

interface SidebarContextType {
  updateSidebar: UpdateSidebarFn
  content: React.ReactNode
}

export const SidebarContext = createContext<SidebarContextType>({
  updateSidebar: content => {
    console.warn(
      'SidebarContext: updateSidebar was called before initialezed',
      content
    )
  },
  content: null,
})
SidebarContext.displayName = 'SidebarContext'

type SidebarProps = {
  children: React.ReactElement
}

const Sidebar = ({ children }: SidebarProps) => {
  const [state, setState] = useState<{ content: React.ReactNode | null }>({
    content: null,
  })

  const update = (content: React.ReactNode | null) => {
    setState({ content })
  }

  return (
    <SidebarContext.Provider
      value={{ updateSidebar: update, content: state.content }}
    >
      {children}
      <SidebarContext.Consumer>
        {value => (
          <SidebarContainer highlighted={value.content != null}>
            {value.content}
            <SidebarDismissButton
              name="angle double right"
              size="big"
              link
              onClick={() => setState({ content: null })}
            />
            <div style={{ height: 100 }}></div>
          </SidebarContainer>
        )}
      </SidebarContext.Consumer>
    </SidebarContext.Provider>
  )
}

Sidebar.propTypes = {
  children: PropTypes.element,
}

export default Sidebar
