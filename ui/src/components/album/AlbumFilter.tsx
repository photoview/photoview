import React from 'react'
import { authToken } from '../../helpers/authentication'
import { useTranslation } from 'react-i18next'
import { OrderDirection } from '../../__generated__/globalTypes'
import { MediaOrdering, SetOrderingFn } from '../../hooks/useOrderingParams'
import Checkbox from '../../primitives/form/Checkbox'

import { ReactComponent as SortingIcon } from './icons/sorting.svg'
import { ReactComponent as DirectionIcon } from './icons/direction-arrow.svg'

import Dropdown from '../../primitives/form/Dropdown'
import classNames from 'classnames'

export type FavoriteCheckboxProps = {
  onlyFavorites: boolean
  setOnlyFavorites(favorites: boolean): void
}

export const FavoritesCheckbox = ({
  onlyFavorites,
  setOnlyFavorites,
}: FavoriteCheckboxProps) => {
  const { t } = useTranslation()

  return (
    <Checkbox
      className="mb-1"
      label={t('album_filter.only_favorites', 'Show only favorites')}
      checked={onlyFavorites}
      onChange={e => setOnlyFavorites(e.target.checked)}
    />
  )
}

type SortingOptionsProps = {
  ordering?: MediaOrdering
  setOrdering?: SetOrderingFn
}

const SortingOptions = ({ setOrdering, ordering }: SortingOptionsProps) => {
  const { t } = useTranslation()

  const changeOrderDirection = () => {
    if (setOrdering && ordering) {
      setOrdering({
        orderDirection:
          ordering.orderDirection == OrderDirection.ASC
            ? OrderDirection.DESC
            : OrderDirection.ASC,
      })
    }
  }

  const changeOrderBy = (value: string) => {
    if (setOrdering) {
      setOrdering({ orderBy: value })
    }
  }

  const sortingOptions = [
    {
      value: 'date_shot',
      label: t('album_filter.sorting_options.date_shot', 'Date shot'),
    },
    {
      value: 'updated_at',
      label: t('album_filter.sorting_options.date_imported', 'Date imported'),
    },
    {
      value: 'title',
      label: t('album_filter.sorting_options.title', 'Title'),
    },
    {
      value: 'type',
      label: t('album_filter.sorting_options.type', 'Kind'),
    },
  ]

  return (
    <fieldset>
      <legend id="filter_group_sort-label" className="inline-block mb-1">
        <SortingIcon
          className="inline-block align-baseline mr-1 mt-1"
          aria-hidden="true"
        />
        <span>{t('album_filter.sort', 'Sort')}</span>
      </legend>
      <div>
        <Dropdown
          aria-labelledby="filter_group_sort-label"
          setSelected={changeOrderBy}
          value={ordering?.orderBy || undefined}
          items={sortingOptions}
        />
        <button
          title="Sort direction"
          aria-label="Sort direction"
          className={classNames(
            'bg-gray-50 h-[30px] align-top px-2 py-1 rounded ml-2 border border-gray-200 focus:outline-none focus:border-blue-300 text-[#8b8b8b] hover:bg-gray-100 hover:text-[#777]',
            'dark:bg-dark-input-bg dark:border-dark-input-border dark:text-dark-input-text dark:focus:border-blue-300',
            { 'flip-y': ordering?.orderDirection == OrderDirection.ASC }
          )}
          onClick={changeOrderDirection}
        >
          <DirectionIcon />
          <span className="sr-only">
            {ordering?.orderDirection == OrderDirection.ASC
              ? 'ascending'
              : 'descending'}
          </span>
        </button>
      </div>
    </fieldset>
  )
}

type AlbumFilterProps = {
  onlyFavorites: boolean
  setOnlyFavorites?(favorites: boolean): void
  ordering?: MediaOrdering
  setOrdering?: SetOrderingFn
}

const AlbumFilter = ({
  onlyFavorites,
  setOnlyFavorites,
  setOrdering,
  ordering,
}: AlbumFilterProps) => {
  return (
    <div className="flex items-end gap-4 flex-wrap mb-4">
      <SortingOptions ordering={ordering} setOrdering={setOrdering} />
      {authToken() && setOnlyFavorites && (
        <FavoritesCheckbox
          onlyFavorites={onlyFavorites}
          setOnlyFavorites={setOnlyFavorites}
        />
      )}
    </div>
  )
}

export default AlbumFilter
