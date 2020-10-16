import { authToken } from '../authentication'
import { Checkbox, Dropdown, Button, Icon } from 'semantic-ui-react'
import React, { useEffect, useState } from 'react'
import styled from 'styled-components'
import PropTypes from 'prop-types'

const sortingOptions = [
  {
    key: 'date_shot',
    value: 'date_shot',
    text: 'Date shot',
  },
  {
    key: 'date_imported',
    value: 'date_imported',
    text: 'Date imported',
  },
  {
    key: 'title',
    value: 'title',
    text: 'Title',
  },
  {
    key: 'kind',
    value: 'kind',
    text: 'Kind',
  },
]

const FavoritesCheckbox = styled(Checkbox)`
  margin-bottom: 16px;
  margin-right: 10px;
`

const OrderDirectionButton = styled(Button)`
  padding: 0.88em;
  margin-left: 10px !important;
`

const AlbumFilter = ({ onlyFavorites, setOnlyFavorites, setOrdering }) => {
  const [orderDirection, setOrderDirection] = useState('ASC')

  const onChangeOrderDirection = (e, data) => {
    const direction = data.children.props.name === 'arrow up' ? 'DESC' : 'ASC'
    setOrderDirection(direction)
  }

  useEffect(() => {
    setOrdering({ orderDirection })
  }, [orderDirection])

  return (
    <>
      {authToken() && (
        <FavoritesCheckbox
          toggle
          label="Show only favorites"
          checked={onlyFavorites}
          onClick={e => e.stopPropagation()}
          onChange={(e, result) => setOnlyFavorites(result.checked)}
        />
      )}
      <strong> Sort by: </strong>
      <Dropdown
        selection
        options={sortingOptions}
        defaultValue={sortingOptions[0].value}
        onChange={(e, data) => {
          setOrdering({ orderBy: data.value })
        }}
      />
      <OrderDirectionButton icon basic onClick={onChangeOrderDirection}>
        <Icon name={'arrow ' + (orderDirection === 'ASC' ? 'up' : 'down')} />
      </OrderDirectionButton>
    </>
  )
}

AlbumFilter.propTypes = {
  onlyFavorites: PropTypes.bool,
  setOnlyFavorites: PropTypes.func,
  setOrdering: PropTypes.func,
}

export default React.memo(AlbumFilter)
