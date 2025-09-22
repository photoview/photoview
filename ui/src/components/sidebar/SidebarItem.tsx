import React from 'react'

type SidebarItemProps = {
  name: string
  value: string
}

const SidebarItem = ({ name, value }: SidebarItemProps) => (
  <div className="grid grid-cols-[100px_1fr] gap-2 items-start">
    <div className="font-semibold text-sm text-[#888] text-right">
      {name}
    </div>
    <div className="text-base break-words">
      {value}
    </div>
  </div>
)

export default SidebarItem
