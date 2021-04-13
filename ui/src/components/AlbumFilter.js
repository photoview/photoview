import React from 'react'
import { authToken } from '../helpers/authentication'
import { Checkbox, Dropdown, Button, Icon } from 'semantic-ui-react'
import styled from 'styled-components'
import PropTypes from 'prop-types'
import { useTranslation } from 'react-i18next'

const FavoritesCheckboxStyle = styled(Checkbox)`
  margin-bottom: 16px;
  margin-right: 10px;

  &.ui.toggle.checkbox label {
    padding-left: 0;
    padding-right: 4em;
    font-weight: bold;
  }

  &.ui.checkbox input,
  &.ui.toggle.checkbox label:before {
    left: auto;
    right: 0;
  }

  &.ui.toggle.checkbox label:after {
    left: auto;
    right: 1.75em;
    transition: background 0.3s ease 0s, right 0.3s ease 0s;
  }

  &.ui.toggle.checkbox input:checked + label:after {
    left: auto;
    right: 0.08em;
    transition: background 0.3s ease 0s, right 0.3s ease 0s;
  }
`

export const FavoritesCheckbox = ({ onlyFavorites, setOnlyFavorites }) => {
  const { t } = useTranslation()

  return (
    <FavoritesCheckboxStyle
      toggle
      label={t('album_filter.only_favorites', 'Show only favorites')}
      checked={onlyFavorites}
      onChange={(e, result) => setOnlyFavorites(result.checked)}
    />
  )
}

FavoritesCheckbox.propTypes = {
  onlyFavorites: PropTypes.bool.isRequired,
  setOnlyFavorites: PropTypes.func.isRequired,
}

const OrderDirectionButton = styled(Button)`
  padding: 0.88em;
  margin-left: 10px !important;
`

const SortByLabel = styled.strong`
  margin-left: 4px;
  margin-right: 6px;
`

const AlbumFilter = ({
  onlyFavorites,
  setOnlyFavorites,
  setOrdering,
  ordering,
}) => {
  const { t } = useTranslation()
  const onChangeOrderDirection = (e, data) => {
    const direction = data.children.props.name === 'arrow up' ? 'DESC' : 'ASC'
    setOrdering({ orderDirection: direction })
  }

  const sortingOptions = [
    {
      key: 'date_shot',
      value: 'date_shot',
      text: t('album_filter.sorting_options.date_shot', 'Date shot'),
    },
    {
      key: 'updated_at',
      value: 'updated_at',
      text: t('album_filter.sorting_options.date_imported', 'Date imported'),
    },
    {
      key: 'title',
      value: 'title',
      text: t('album_filter.sorting_options.title', 'Title'),
    },
    {
      key: 'type',
      value: 'type',
      text: t('album_filter.sorting_options.type', 'Kind'),
    },
  ]

  return (
    <>
      {authToken() && (
        <FavoritesCheckbox
          onlyFavorites={onlyFavorites}
          setOnlyFavorites={setOnlyFavorites}
        />
      )}
      <SortByLabel>{t('album_filter.sort_by', 'Sort by')}</SortByLabel>
      <Dropdown
        selection
        options={sortingOptions}
        defaultValue={
          sortingOptions.find(e => e.value === ordering.orderBy)?.value ||
          sortingOptions[0].value
        }
        onChange={(e, data) => {
          setOrdering({ orderBy: data.value })
        }}
      />
      <OrderDirectionButton icon basic onClick={onChangeOrderDirection}>
        <Icon
          name={'arrow ' + (ordering.orderDirection === 'ASC' ? 'up' : 'down')}
        />
      </OrderDirectionButton>
    </>
  )
}

AlbumFilter.propTypes = {
  onlyFavorites: PropTypes.bool,
  setOnlyFavorites: PropTypes.func,
  setOrdering: PropTypes.func,
  ordering: PropTypes.object,
}

export default AlbumFilter
