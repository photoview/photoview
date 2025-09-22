import React from 'react'

type SidebarItemProps = {
  name: string
  value: string
}

const SidebarItem = ({ name, value }: SidebarItemProps) => (
  <div className="grid grid-cols-[8rem_1fr] md:grid-cols-[12rem_1fr] gap-2 items-start">
    <div className="font-semibold text-sm text-gray-500 dark:text-gray-400 text-right">
      {name}
    </div>
    <div className="text-base break-words select-text">
      {value}
    </div>
  </div>
)

export default SidebarItem
