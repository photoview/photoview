import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { availableRoles } from './__generated__/availableRoles'
import Dropdown, { DropdownItem } from '../../../primitives/form/Dropdown'
import { useTranslation } from 'react-i18next'

export const ROLE_QUERY = gql`
  query availableRoles {
    roles {
      id
      name
    }
  }
`

interface RoleSelectorProps {
  onRoleSelected: (roleId: string) => void | undefined
  selected: string
}

export const RoleSelector = (props: RoleSelectorProps) => {
  const { loading, data, error } = useQuery<availableRoles>(ROLE_QUERY)
  const { t } = useTranslation()

  const items: DropdownItem[] = []

  if (error) {
    return <div> Error</div>
  }
  const placeholder = loading
    ? t('general.loading.default', 'Loading...')
    : t('general.please_select', 'Please Select')
  items.push({ value: '', label: placeholder })

  if (!loading) {
    items.push(
      ...data!.roles.map(
        role => ({ value: role.id, label: role.name } as DropdownItem)
      )
    )
  }

  const onSelected = (selected: string) => {
    props.onRoleSelected(selected)
  }

  return (
    <Dropdown
      selected={props.selected}
      items={items}
      setSelected={onSelected}
    />
  )
}
