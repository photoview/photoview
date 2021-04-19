import React from 'react'
import styled from 'styled-components'

const ItemName = styled.div`
  display: inline-block;
  width: 100px;
  font-weight: 600;
  font-size: 0.85rem;
  color: #888;
  text-align: right;
  margin-right: 0.5rem;
`

const ItemValue = styled.div`
  display: inline-block;
  font-size: 1rem;
`

type SidebarItemProps = {
  name: string
  value: string
}

const SidebarItem = ({ name, value }: SidebarItemProps) => (
  <div>
    <ItemName>{name}</ItemName>
    <ItemValue>{value}</ItemValue>
  </div>
)

export default SidebarItem
