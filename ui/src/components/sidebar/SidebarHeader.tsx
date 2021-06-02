import React, { useContext } from 'react'

import { ReactComponent as CloseIcon } from './icons/closeSidebarIcon.svg'
import { ReactComponent as PinIconOutline } from './icons/pinSidebarIconOutline.svg'
import { ReactComponent as PinIconFilled } from './icons/pinSidebarIconFilled.svg'
import { SidebarContext } from './Sidebar'

type SidebarHeaderProps = {
  title: string
}

const SidebarHeader = ({ title }: SidebarHeaderProps) => {
  const { updateSidebar, setPinned, pinned } = useContext(SidebarContext)

  const PinIcon = pinned ? PinIconFilled : PinIconOutline

  return (
    <div className="m-2 flex items-center gap-2">
      {!pinned && (
        <button title="Close sidebar" onClick={() => updateSidebar(null)}>
          <CloseIcon className="m-2" />
        </button>
      )}
      <span className="flex-grow -mt-1">{title}</span>
      <button title="Pin sidebar" onClick={() => setPinned(!pinned)}>
        <PinIcon className="m-2" />
      </button>
    </div>
  )
}

export default SidebarHeader
