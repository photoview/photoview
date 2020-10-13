import { authToken } from '../authentication'
import { Checkbox, Dropdown } from 'semantic-ui-react'
import React from 'react'
import styled from 'styled-components'
import PropTypes from 'prop-types'

const sortingOptions = [
  {
    key: 'date_shot.ASC',
    value: 'date_shot.ASC',
    text: 'Date shot ↑',
  },
  {
    key: 'date_shot.DESC',
    value: 'date_shot.DESC',
    text: 'Date shot ↓',
  },
  {
    key: 'date_imported.ASC',
    value: 'date_imported.ASC',
    text: 'Date imported ↑',
  },
  {
    key: 'date_imported.DESC',
    value: 'date_imported.DESC',
    text: 'Date imported ↓',
  },
  {
    key: 'title.ASC',
    value: 'title.ASC',
    text: 'Title ↑',
  },
  {
    key: 'title.DESC',
    value: 'title.DESC',
    text: 'Title ↓',
  },
  {
    key: 'kind.ASC',
    value: 'kind.ASC',
    text: 'Kind ↑',
  },
  {
    key: 'kind.DESC',
    value: 'kind.DESC',
    text: 'Kind ↓',
  },
]

const FavoritesCheckbox = styled(Checkbox)`
  margin-bottom: 16px;
  margin-right: 10px;
`

const AlbumFilter = ({ onlyFavorites, setOnlyFavorites, setSorting }) => {
  return (
    <>
      {authToken() && (
        <FavoritesCheckbox
          toggle
          label="Show only favorites"
          checked={onlyFavorites}
          onClick={e => e.stopPropagation()}
          onChange={setOnlyFavorites}
        />
      )}
      <strong> Sort by: </strong>
      <Dropdown
        options={sortingOptions}
        defaultValue={sortingOptions[0].value}
        onChange={setSorting}
      />
    </>
  )
}

AlbumFilter.propTypes = {
  onlyFavorites: PropTypes.bool,
  setOnlyFavorites: PropTypes.func,
  setSorting: PropTypes.func,
}

export default React.memo(AlbumFilter)
