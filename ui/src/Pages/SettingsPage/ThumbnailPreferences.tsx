import { gql } from '@apollo/client'
import React, { useRef, useState } from 'react'
import { useMutation, useQuery } from '@apollo/client'
import {
  SectionTitle,
  InputLabelDescription,
  InputLabelTitle,
} from './SettingsPage'
import { useTranslation } from 'react-i18next'
import { ThumbnailFilter } from '../../__generated__/globalTypes'
import { thumbnailMethodQuery } from './__generated__/thumbnailMethodQuery'
import {
  setThumbnailMethodMutation,
  setThumbnailMethodMutationVariables,
} from './__generated__/setThumbnailMethodMutation'
import Dropdown, { DropdownItem } from '../../primitives/form/Dropdown'
import Loader from '../../primitives/Loader'

export const THUMBNAIL_METHOD_QUERY = gql`
  query thumbnailMethodQuery {
    siteInfo {
      thumbnailMethod
    }
  }
`

export const SET_THUMBNAIL_METHOD_MUTATION = gql`
  mutation setThumbnailMethodMutation($method: ThumbnailFilter!) {
    setThumbnailDownsampleMethod(method: $method)
  }
`

const ThumbnailPreferences = () => {
  const { t } = useTranslation()

  const downsampleMethodServerValue = useRef<null | number>(null)
  const [downsampleMethod, setDownsampleMethod] = useState(0)

  const downsampleMethodQuery = useQuery<thumbnailMethodQuery>(
    THUMBNAIL_METHOD_QUERY,
    {
      onCompleted(data) {
        setDownsampleMethod(data.siteInfo.thumbnailMethod)
        downsampleMethodServerValue.current = data.siteInfo.thumbnailMethod
      },
    }
  )

  const [setDownsampleMutation, downsampleMutationData] = useMutation<
    setThumbnailMethodMutation,
    setThumbnailMethodMutationVariables
  >(SET_THUMBNAIL_METHOD_MUTATION)

  const updateDownsampleMethod = (downsampleMethod: number) => {
    if (downsampleMethodServerValue.current != downsampleMethod) {
      downsampleMethodServerValue.current = downsampleMethod
      setDownsampleMutation({
        variables: {
          method: downsampleMethod,
        },
      })
    }
  }

  const methodItems: DropdownItem[] = [
    {
      label: t(
        'settings.thumbnails.method.filter.nearest_neighbor',
        'Nearest Neighbor (default)'
      ),
      value: ThumbnailFilter.NearestNeighbor,
    },
    {
      label: t('settings.thumbnails.method.filter.box', 'Box'),
      value: ThumbnailFilter.Box,
    },
    {
      label: t('settings.thumbnails.method.filter.linear', 'Linear'),
      value: ThumbnailFilter.Linear,
    },
    {
      label: t(
        'settings.thumbnails.method.filter.mitchell_netravali',
        'Mitchell-Netravali'
      ),
      value: ThumbnailFilter.MitchellNetravali,
    },
    {
      label: t('settings.thumbnails.method.filter.catmull_rom', 'Catmull-Rom'),
      value: ThumbnailFilter.CatmullRom,
    },
    {
      label: t(
        'settings.thumbnails.method.filter.Lanczos',
        'Lanczos (highest quality)'
      ),
      value: ThumbnailFilter.Lanczos,
    },
  ]

  return (
    <div>
      <SectionTitle>
        {t('settings.thumbnails.title', 'Thumbnail preferences')}
      </SectionTitle>
      <label htmlFor="thumbnail_method_field">
        <InputLabelTitle>
          {t('settings.thumbnails.method.label', 'Downsampling method')}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.thumbnails.method.description',
            'The filter to use when generating thumbnails'
          )}
        </InputLabelDescription>
      </label>
      <Dropdown
        aria-label="Method"
        items={methodItems}
        selected={downsampleMethod}
        setSelected={value => {
          setDownsampleMethod(value)
          updateDownsampleMethod(value)
        }}
      />
      <Loader
        active={downsampleMethodQuery.loading || downsampleMutationData.loading}
        size="small"
        style={{ marginLeft: 16 }}
      />
    </div>
  )
}

export default ThumbnailPreferences
