import React, { useState } from 'react'

import {
  Button,
  Checkbox,
  Dropdown,
  Icon,
  Input,
  Loader,
} from 'semantic-ui-react'
import { useMutation, useQuery } from 'react-apollo'
import gql from 'graphql-tag'
import styled from 'styled-components'
import { SectionTitle } from './SettingsPage'

const SCAN_MUTATION = gql`
  mutation scanAllMutation {
    scanAll {
      success
      message
    }
  }
`

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

const InputLabelTitle = styled.p`
  font-size: 1.1em;
  font-weight: 600;
  margin: 1em 0 0 !important;
`

const InputLabelDescription = styled.p`
  font-size: 0.9em;
  margin: 0 0 0.5em !important;
`

const ScannerSection = () => {
  const [startScanner, { called }] = useMutation(SCAN_MUTATION)

  const [enablePeriodicScanner, setEnablePeriodicScanner] = useState(false)
  const [scanInterval, setScanInterval] = useState({
    value: 4,
    unit: 'minute',
  })

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

    setScanIntervalMutation({
      variables: {
        interval: checked ? convertToSeconds(scanInterval) : 0,
      },
    })
  }

  const onScanIntervalUpdate = scanInterval => {
    setScanIntervalMutation({
      variables: {
        interval: convertToSeconds(scanInterval),
      },
    })
  }

  const scanIntervalUnits = [
    {
      key: 'second',
      text: 'Seconds',
      value: 'second',
    },
    {
      key: 'minute',
      text: 'Minutes',
      value: 'minute',
    },
    {
      key: 'hour',
      text: 'Hours',
      value: 'hour',
    },
    {
      key: 'day',
      text: 'Days',
      value: 'day',
    },
    {
      key: 'month',
      text: 'Months',
      value: 'month',
    },
  ]

  return (
    <div>
      <SectionTitle nospace>Scanner</SectionTitle>
      <Button
        icon
        labelPosition="left"
        onClick={() => {
          startScanner()
        }}
        disabled={called}
      >
        <Icon name="sync" />
        Scan All
      </Button>

      <h3>Periodic scanner</h3>

      <div style={{ margin: '12px 0' }}>
        <Checkbox
          label="Enable periodic scanner"
          disabled={scanIntervalQuery.loading}
          checked={enablePeriodicScanner}
          onChange={(_, { checked }) => onScanIntervalCheckboxChange(checked)}
        />
      </div>

      {enablePeriodicScanner && (
        <>
          <label htmlFor="periodic_scan_field">
            <InputLabelTitle>Periodic scan interval</InputLabelTitle>
            <InputLabelDescription>
              How often the scanner should perform automatic scans of all users
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
    </div>
  )
}

export default ScannerSection
