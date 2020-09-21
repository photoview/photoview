import gql from 'graphql-tag'
import React, { useRef, useState } from 'react'
import { useQuery, useMutation } from 'react-apollo'
import { Input, Loader } from 'semantic-ui-react'
import { InputLabelTitle, InputLabelDescription } from './SettingsPage'

const CONCURRENT_WORKERS_QUERY = gql`
  query concurrentWorkersQuery {
    siteInfo {
      concurrentWorkers
    }
  }
`

const SET_CONCURRENT_WORKERS_MUTATION = gql`
  mutation setConcurrentWorkers($workers: Int!) {
    setScannerConcurrentWorkers(workers: $workers)
  }
`

const ScannerConcurrentWorkers = () => {
  const workerAmountQuery = useQuery(CONCURRENT_WORKERS_QUERY, {
    onCompleted(data) {
      setWorkerAmount(data.siteInfo.concurrentWorkers)
      workerAmountServerValue.current = data.siteInfo.concurrentWorkers
    },
  })

  const [setWorkersMutation, workersMutationData] = useMutation(
    SET_CONCURRENT_WORKERS_MUTATION
  )

  const workerAmountServerValue = useRef(null)
  const [workerAmount, setWorkerAmount] = useState('')

  const updateWorkerAmount = workerAmount => {
    if (workerAmountServerValue.current != workerAmount) {
      workerAmountServerValue.current = workerAmount
      setWorkersMutation({
        variables: {
          workers: workerAmount,
        },
      })
    }
  }

  return (
    <div style={{ marginTop: 32 }}>
      <label htmlFor="scanner_concurrent_workers_field">
        <InputLabelTitle>Scanner Concurrent Workers</InputLabelTitle>
        <InputLabelDescription>
          The maximum amount of scanner jobs that is allowed to run at once
        </InputLabelDescription>
      </label>
      <Input
        disabled={workerAmountQuery.loading || workersMutationData.loading}
        type="number"
        min="1"
        max="24"
        id="scanner_concurrent_workers_field"
        value={workerAmount}
        onChange={(_, { value }) => {
          setWorkerAmount(value)
        }}
        onBlur={() => updateWorkerAmount(workerAmount)}
        onKeyDown={({ key }) =>
          key == 'Enter' && updateWorkerAmount(workerAmount)
        }
      />
      <Loader
        active={workerAmountQuery.loading || workersMutationData.loading}
        inline
        size="small"
        style={{ marginLeft: 16 }}
      />
    </div>
  )
}

export default ScannerConcurrentWorkers
