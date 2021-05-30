import React, { useContext } from 'react'

import { ReactComponent as CloseIcon } from './icons/closeSidebarIcon.svg'
import { ReactComponent as PinIcon } from './icons/pinSidebarIcon.svg'
import { SidebarContext } from './Sidebar'

type SidebarHeaderProps = {
  title: string
}

const SidebarHeader = ({ title }: SidebarHeaderProps) => {
  const { updateSidebar } = useContext(SidebarContext)

  return (
    <div className="m-2 flex items-center gap-2">
      <button title="Close sidebar" onClick={() => updateSidebar(null)}>
        <CloseIcon className="m-2" />
      </button>
      <span className="flex-grow -mt-1">{title}</span>
      <button title="Pin sidebar">
        <PinIcon className="m-2" />
      </button>
    </div>
  )
}

export default SidebarHeader
