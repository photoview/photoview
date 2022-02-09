import React from 'react'

type SidebarItemProps = {
  name: string
  value: string
}

const SidebarItem = ({ name, value }: SidebarItemProps) => (
  <div>
    <div className="inline-block w-[100px] font-semibold text-sm text-[#888] text-right mr-2">
      {name}
    </div>
    <div className="inline-block text-base">{value}</div>
  </div>
)

export default SidebarItem
