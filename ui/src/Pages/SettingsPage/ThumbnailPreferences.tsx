import { gql } from '@apollo/client'
import React, { useRef, useState } from 'react'
import { useMutation, useQuery } from '@apollo/client'
import { InputLabelDescription, InputLabelTitle } from './SettingsPage'
import { useTranslation } from 'react-i18next'
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
  mutation setThumbnailMethodMutation($method: Int!) {
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
        'settings.thumbnails.downsample_method.nearest_neighbor',
        'Nearest Neighbor'
      ),
      value: 0,
    },
    {
      label: t('settings.thumbnails.downsample_method.box', 'Box'),
      value: 1,
    },
    {
      label: t('settings.thumbnails.downsample_method.linear', 'Linear'),
      value: 2,
    },
    {
      label: t(
        'settings.thumbnails.downsample_method.mitchell_netravali',
        'Mitchell-Netravali'
      ),
      value: 3,
    },
    {
      label: t(
        'settings.thumbnails.downsample_method.catmull_rom',
        'Catmull-Rom'
      ),
      value: 4,
    },
    {
      label: t('settings.thumbnails.downsample_method.Lanczos', 'Lanczos'),
      value: 5,
    },
  ]

  // const [enablePeriodicScanner, setEnablePeriodicScanner] = useState(false)
  // const [thumbnailMethod, setThumbnailMethod] = useState({
  //   value: 0,
  //   unit: TimeUnit.Second,
  // })
  //
  // const scanIntervalServerValue = useRef<number | null>(null)
  //
  // const scanIntervalQuery = useQuery<scanIntervalQuery>(SCAN_INTERVAL_QUERY, {
  //   onCompleted(data) {
  //     const queryScanInterval = data.siteInfo.periodicScanInterval
  //
  //     if (queryScanInterval == 0) {
  //       setScanInterval({
  //         unit: TimeUnit.Second,
  //         value: 0,
  //       })
  //     } else {
  //       setScanInterval(
  //         convertToAppropriateUnit({
  //           unit: TimeUnit.Second,
  //           value: queryScanInterval,
  //         })
  //       )
  //     }
  //
  //     setEnablePeriodicScanner(queryScanInterval > 0)
  //   },
  // })
  //
  // const [setScanIntervalMutation, { loading: scanIntervalMutationLoading }] =
  //   useMutation<
  //     changeScanIntervalMutation,
  //     changeScanIntervalMutationVariables
  //   >(SCAN_INTERVAL_MUTATION)
  //
  // const onScanIntervalCheckboxChange = (checked: boolean) => {
  //   setEnablePeriodicScanner(checked)
  //
  //   onScanIntervalUpdate(
  //     checked ? scanInterval : { value: 0, unit: TimeUnit.Second }
  //   )
  // }
  //
  // const onScanIntervalUpdate = (scanInterval: TimeValue) => {
  //   const seconds = convertToSeconds(scanInterval)
  //
  //   if (scanIntervalServerValue.current != seconds) {
  //     setScanIntervalMutation({
  //       variables: {
  //         interval: seconds,
  //       },
  //     })
  //     scanIntervalServerValue.current = seconds
  //   }
  // }

  return (
    <>
      <div className="mt-4">
        <label htmlFor="thumbnail_method_field">
          <InputLabelTitle>
            {t('settings.thumbnails.field.label', 'Downsampling method')}
          </InputLabelTitle>
          <InputLabelDescription>
            {t(
              'settings.thumbnails.field.description',
              'The filter to use when generating thumbnails'
            )}
          </InputLabelDescription>
        </label>
        <div className="flex gap-2">
          <Dropdown
            aria-label="Method"
            items={methodItems}
            selected={downsampleMethod}
            setSelected={value => {
              setDownsampleMethod(value)
              updateDownsampleMethod(value)
            }}
          />
        </div>
      </div>
      <Loader
        active={downsampleMethodQuery.loading || downsampleMutationData.loading}
        size="small"
        style={{ marginLeft: 16 }}
      />
    </>
  )
}

export default ThumbnailPreferences

// <h3 className="font-semibold text-lg mt-4 mb-2">
//   {t('settings.thumbnails.title', 'Thumbnail preferences')}
// </h3>
