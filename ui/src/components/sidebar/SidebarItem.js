import React from 'react'
import styled from 'styled-components'

const ItemName = styled.div`
  display: inline-block;
  width: 100px;
  font-weight: 600;
  font-size: 12px;
  color: #888;
  text-align: right;
  margin-right: 8px;
`

const ItemValue = styled.div`
  display: inline-block;
`

export const SidebarItem = ({ name, value }) => (
  <div>
    <ItemName>{name}</ItemName>
    <ItemValue>{value}</ItemValue>
  </div>
)
