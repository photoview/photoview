import { gql } from '@apollo/client'
import React, { useRef, useState } from 'react'
import { useMutation, useQuery } from '@apollo/client'
import { Checkbox, Dropdown, Input, Loader } from 'semantic-ui-react'
import { InputLabelDescription, InputLabelTitle } from './SettingsPage'
import { useTranslation } from 'react-i18next'

const SCAN_INTERVAL_QUERY = gql`
  query scanIntervalQuery {
    siteInfo {
      periodicScanInterval
    }
  }
`

const SCAN_INTERVAL_MUTATION = gql`
  mutation changeScanIntervalMutation($interval: Int!) {
    setPeriodicScanInterval(interval: $interval)
  }
`

const timeUnits = [
  {
    value: 'second',
    multiplier: 1,
  },
  {
    value: 'minute',
    multiplier: 60,
  },
  {
    value: 'hour',
    multiplier: 60 * 60,
  },
  {
    value: 'day',
    multiplier: 60 * 60 * 24,
  },
  {
    value: 'month',
    multiplier: 60 * 60 * 24 * 30,
  },
]

const convertToSeconds = ({ value, unit }) => {
  return parseInt(value * timeUnits.find(x => x.value == unit).multiplier)
}

const convertToAppropriateUnit = ({ value, unit }) => {
  if (value == 0) {
    return {
      unit: 'second',
      value: 0,
    }
  }

  const seconds = convertToSeconds({ value, unit })

  let resultingUnit = timeUnits.first
  for (const unit of timeUnits) {
    if (
      seconds / unit.multiplier >= 1 &&
      (seconds / unit.multiplier) % 1 == 0
    ) {
      resultingUnit = unit
    } else {
      break
    }
  }

  return {
    unit: resultingUnit.value,
    value: seconds / resultingUnit.multiplier,
  }
}

const PeriodicScanner = () => {
  const { t } = useTranslation()

  const [enablePeriodicScanner, setEnablePeriodicScanner] = useState(false)
  const [scanInterval, setScanInterval] = useState({
    value: 4,
    unit: 'minute',
  })

  const scanIntervalServerValue = useRef(null)

  const scanIntervalQuery = useQuery(SCAN_INTERVAL_QUERY, {
    onCompleted(data) {
      const queryScanInterval = data.siteInfo.periodicScanInterval

      if (queryScanInterval == 0) {
        setScanInterval({
          unit: 'second',
          value: '',
        })
      } else {
        setScanInterval(
          convertToAppropriateUnit({
            unit: 'second',
            value: queryScanInterval,
          })
        )
      }

      setEnablePeriodicScanner(queryScanInterval > 0)
    },
  })

  const [
    setScanIntervalMutation,
    { loading: scanIntervalMutationLoading },
  ] = useMutation(SCAN_INTERVAL_MUTATION)

  const onScanIntervalCheckboxChange = checked => {
    setEnablePeriodicScanner(checked)

    onScanIntervalUpdate(checked ? scanInterval : { value: 0, unit: 'second' })
  }

  const onScanIntervalUpdate = scanInterval => {
    const seconds = convertToSeconds(scanInterval)

    if (scanIntervalServerValue.current != seconds) {
      setScanIntervalMutation({
        variables: {
          interval: convertToSeconds(scanInterval),
        },
      })
      scanIntervalServerValue.current = seconds
    }
  }

  const scanIntervalUnits = [
    {
      key: 'second',
      text: t('settings.periodic_scanner.interval_unit.seconds', 'Seconds'),
      value: 'second',
    },
    {
      key: 'minute',
      text: t('settings.periodic_scanner.interval_unit.minutes', 'Minutes'),
      value: 'minute',
    },
    {
      key: 'hour',
      text: t('settings.periodic_scanner.interval_unit.hour', 'Hour'),
      value: 'hour',
    },
    {
      key: 'day',
      text: t('settings.periodic_scanner.interval_unit.days', 'Days'),
      value: 'day',
    },
    {
      key: 'month',
      text: t('settings.periodic_scanner.interval_unit.months', 'Months'),
      value: 'month',
    },
  ]

  return (
    <>
      <h3>{t('settings.periodic_scanner.title', 'Periodic scanner')}</h3>

      <div style={{ margin: '12px 0' }}>
        <Checkbox
          label={t(
            'settings.periodic_scanner.checkbox_label',
            'Enable periodic scanner'
          )}
          disabled={scanIntervalQuery.loading}
          checked={enablePeriodicScanner}
          onChange={(_, { checked }) => onScanIntervalCheckboxChange(checked)}
        />
      </div>

      {enablePeriodicScanner && (
        <>
          <label htmlFor="periodic_scan_field">
            <InputLabelTitle>
              {t(
                'settings.periodic_scanner.field.label',
                'Periodic scan interval'
              )}
            </InputLabelTitle>
            <InputLabelDescription>
              {t(
                'settings.periodic_scanner.field.description',
                'How often the scanner should perform automatic scans of all users'
              )}
            </InputLabelDescription>
          </label>
          <Input
            label={
              <Dropdown
                onChange={(_, { value }) => {
                  const newScanInterval = {
                    ...scanInterval,
                    unit: value,
                  }

                  setScanInterval(newScanInterval)
                  onScanIntervalUpdate(newScanInterval)
                }}
                value={scanInterval.unit}
                options={scanIntervalUnits}
              />
            }
            onBlur={() => onScanIntervalUpdate(scanInterval)}
            onKeyDown={({ key }) =>
              key == 'Enter' && onScanIntervalUpdate(scanInterval)
            }
            loading={scanIntervalQuery.loading}
            labelPosition="right"
            style={{ maxWidth: 300 }}
            id="periodic_scan_field"
            value={scanInterval.value}
            onChange={(_, { value }) => {
              setScanInterval(x => ({
                ...x,
                value,
              }))
            }}
          />
        </>
      )}
      <Loader
        active={scanIntervalQuery.loading || scanIntervalMutationLoading}
        inline
        size="small"
        style={{ marginLeft: 16 }}
      />
    </>
  )
}

export default PeriodicScanner
