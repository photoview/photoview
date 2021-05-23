import React from 'react'
import { authToken } from '../helpers/authentication'
import { useTranslation } from 'react-i18next'
import { OrderDirection } from '../../__generated__/globalTypes'
import { MediaOrdering, SetOrderingFn } from '../hooks/useOrderingParams'

type FavoriteCheckboxProps = {
  onlyFavorites: boolean
  setOnlyFavorites(favorites: boolean): void
}

export const FavoritesCheckbox = ({
  onlyFavorites,
  setOnlyFavorites,
}: FavoriteCheckboxProps) => {
  const { t } = useTranslation()

  return (
    <label>
      <input
        type="checkbox"
        checked={onlyFavorites}
        onChange={e => setOnlyFavorites(e.target.checked)}
      />
      <span>{t('album_filter.only_favorites', 'Show only favorites')}</span>
    </label>
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

  const changeOrderBy = (e: React.ChangeEvent<HTMLSelectElement>) => {
    if (setOrdering) {
      setOrdering({ orderBy: e.target.value })
    }
  }

  const sortingOptions = [
    {
      value: 'date_shot',
      text: t('album_filter.sorting_options.date_shot', 'Date shot'),
    },
    {
      value: 'updated_at',
      text: t('album_filter.sorting_options.date_imported', 'Date imported'),
    },
    {
      value: 'title',
      text: t('album_filter.sorting_options.title', 'Title'),
    },
    {
      value: 'type',
      text: t('album_filter.sorting_options.type', 'Kind'),
    },
  ]

  return (
    <>
      {authToken() && setOnlyFavorites && (
        <FavoritesCheckbox
          onlyFavorites={onlyFavorites}
          setOnlyFavorites={setOnlyFavorites}
        />
      )}
      <span>{t('album_filter.sort_by', 'Sort by')}</span>

      <select onChange={changeOrderBy} value={ordering?.orderBy || undefined}>
        {sortingOptions.map(x => (
          <option key={x.value} value={x.value}>
            {x.text}
          </option>
        ))}
      </select>

      <button onClick={changeOrderDirection}>{ordering?.orderDirection}</button>
    </>
  )
}

export default AlbumFilter
