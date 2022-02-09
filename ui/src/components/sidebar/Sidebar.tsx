import React, { createContext, useContext, useEffect, useState } from 'react'

export type UpdateSidebarFn = (content: React.ReactNode) => void
export type SidebarPinnedFn = (pin: boolean) => void

interface SidebarContextType {
  updateSidebar: UpdateSidebarFn
  setPinned: SidebarPinnedFn
  content: React.ReactNode
  pinned: boolean
}

export const SidebarContext = createContext<SidebarContextType>({
  updateSidebar: content => {
    console.warn(
      'SidebarContext: updateSidebar was called before initialized',
      content
    )
  },
  setPinned: content => {
    console.warn(
      'SidebarContext: setPinned was called before initialized',
      content
    )
  },
  content: null,
  pinned: false,
})
SidebarContext.displayName = 'SidebarContext'

type SidebarProviderProps = {
  children: React.ReactChild | React.ReactChild[]
}

export const SidebarProvider = ({ children }: SidebarProviderProps) => {
  const [state, setState] = useState<{
    content: React.ReactNode | null
    pinned: boolean
  }>({
    content: null,
    pinned: false,
  })

  const updateSidebar = (content: React.ReactNode | null) => {
    if (content) {
      setState(state => ({ ...state, content }))
    } else {
      setState(state => ({ ...state, content: null, pinned: false }))
    }
  }

  const setPinned = (pinned: boolean) => {
    setState(state => ({ ...state, pinned }))
  }

  return (
    <SidebarContext.Provider
      value={{
        updateSidebar,
        setPinned,
        content: state.content,
        pinned: state.pinned,
      }}
    >
      {children}
    </SidebarContext.Provider>
  )
}

export const Sidebar = () => {
  const { content, pinned } = useContext(SidebarContext)

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
      className={`fixed top-[72px] bg-white dark:bg-dark-bg2 dark:border-dark-border2 bottom-0 w-full overflow-y-auto transform transition-transform motion-reduce:transition-none ${
        content == null && !pinned ? 'translate-x-full' : 'translate-x-0'
      } ${
        pinned ? 'lg:border-l' : 'lg:shadow-separator'
      } lg:w-[420px] lg:right-0 lg:top-0 lg:z-40`}
    >
      {content}
      <div className="h-24"></div>
    </div>
  )
}
