import React, { createContext, useContext, useEffect, useState } from 'react'

export type UpdateSidebarFn = (content: React.ReactNode) => void

interface SidebarContextType {
  updateSidebar: UpdateSidebarFn
  content: React.ReactNode
}

export const SidebarContext = createContext<SidebarContextType>({
  updateSidebar: content => {
    console.warn(
      'SidebarContext: updateSidebar was called before initialized',
      content
    )
  },
  content: null,
})
SidebarContext.displayName = 'SidebarContext'

type SidebarProviderProps = {
  children: React.ReactChild | React.ReactChild[]
}

export const SidebarProvider = ({ children }: SidebarProviderProps) => {
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
    </SidebarContext.Provider>
  )
}

export const Sidebar = () => {
  const { content } = useContext(SidebarContext)

  useEffect(() => {
    const body = document.body

    if (content == null) {
      body.classList.remove('overflow-y-hidden')
      body.classList.remove('lg:overflow-y-auto')
    } else {
      body.classList.add('overflow-y-hidden')
      body.classList.add('lg:overflow-y-auto')
    }

    return () => {
      body.classList.remove('overflow-y-hidden')
      body.classList.remove('lg:overflow-y-auto')
    }
  })

  return (
    <div
      className={`fixed top-[72px] bg-white bottom-0 w-full overflow-y-scroll transform transition-transform motion-reduce:transition-none ${
        content == null ? 'translate-x-full' : 'translate-x-0'
      } lg:w-[420px] lg:right-0 lg:shadow-separator lg:top-0 lg:z-40`}
    >
      {content}
      <div className="h-24"></div>
    </div>
  )
}
