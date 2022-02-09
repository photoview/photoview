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
    <div className="m-2 flex items-center">
      <button
        className={`${pinned ? 'lg:hidden' : ''}`}
        title="Close sidebar"
        onClick={() => updateSidebar(null)}
      >
        <CloseIcon className="m-2 text-[#1F2021] dark:text-[#abadaf]" />
      </button>
      <span className="flex-grow -mt-1 ml-2">{title}</span>
      <button
        className="hidden lg:block"
        title="Pin sidebar"
        onClick={() => setPinned(!pinned)}
      >
        <PinIcon className="m-2 text-[#1F2021] dark:text-[#abadaf]" />
      </button>
    </div>
  )
}

export default SidebarHeader
