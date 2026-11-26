import React from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import { InputLabelTitle } from './SettingsPage'
import { ScannerJobStatus } from '../../__generated__/globalTypes'
import {
  scannerQueueStatusQuery,
  scannerQueueStatusQuery_scannerQueueStatus,
} from './__generated__/scannerQueueStatusQuery'
import {
  cancelScanJobMutation,
  cancelScanJobMutationVariables,
} from './__generated__/cancelScanJobMutation'
import { cancelAllScanJobsMutation } from './__generated__/cancelAllScanJobsMutation'
import { ReactComponent as DismissIcon } from '../../components/messages/icons/dismissIcon.svg'
import { Button } from '../../primitives/form/Input'

export const SCANNER_QUEUE_STATUS_QUERY = gql`
  query scannerQueueStatusQuery {
    scannerQueueStatus {
      status
      album {
        id
        title
        path {
          id
          title
        }
      }
    }
  }
`

const CANCEL_SCAN_JOB_MUTATION = gql`
  mutation cancelScanJobMutation($albumId: ID!) {
    cancelScanJob(albumId: $albumId)
  }
`

export const CANCEL_ALL_SCAN_JOBS_MUTATION = gql`
  mutation cancelAllScanJobsMutation {
    cancelAllScanJobs
  }
`

const albumBreadcrumb = (
  item: scannerQueueStatusQuery_scannerQueueStatus
): string => {
  const ancestors = item.album.path
    .slice()
    .reverse()
    .map(a => a.title)
  return [...ancestors, item.album.title].join(' / ')
}

type QueueRowProps = {
  item: scannerQueueStatusQuery_scannerQueueStatus
  statusLabel: string
  statusClassName: string
}

const QueueRow = ({ item, statusLabel, statusClassName }: QueueRowProps) => {
  const { t } = useTranslation()
  const [cancelScanJob, { loading }] = useMutation<
    cancelScanJobMutation,
    cancelScanJobMutationVariables
  >(CANCEL_SCAN_JOB_MUTATION, {
    // The global Apollo error link already shows a toast; this only
    // consumes the rejected promise so it isn't left unhandled.
    onError: () => undefined,
  })

  return (
    <li className="flex justify-between items-center gap-4 py-1 border-b border-gray-100 dark:border-dark-border2">
      <span className="truncate">{albumBreadcrumb(item)}</span>
      <span className="flex items-center gap-2 shrink-0">
        <span className={statusClassName}>{statusLabel}</span>
        <button
          title={t('settings.scanner.queue_status.cancel', 'Cancel')}
          aria-label={t('settings.scanner.queue_status.cancel', 'Cancel')}
          disabled={loading}
          onClick={() =>
            cancelScanJob({ variables: { albumId: item.album.id } })
          }
          className="p-1 disabled:opacity-40"
        >
          <DismissIcon className="w-[10px] h-[10px] text-gray-500 dark:text-gray-300" />
        </button>
      </span>
    </li>
  )
}

export const ScannerQueueStatus = () => {
  const { t } = useTranslation()
  const { data, refetch } = useQuery<scannerQueueStatusQuery>(
    SCANNER_QUEUE_STATUS_QUERY,
    { pollInterval: 2000 }
  )
  const [cancelAllScanJobs, { loading: cancellingAll }] =
    useMutation<cancelAllScanJobsMutation>(CANCEL_ALL_SCAN_JOBS_MUTATION, {
      onCompleted: () => refetch(),
      onError: () => undefined,
    })

  const items = data?.scannerQueueStatus ?? []
  if (items.length === 0) return null

  const running = items.filter(i => i.status === ScannerJobStatus.RUNNING)
  const queued = items.filter(i => i.status === ScannerJobStatus.QUEUED)

  return (
    <div>
      <div className="flex justify-between items-center">
        <InputLabelTitle>
          {t('settings.scanner.queue_status.title', 'Scan queue')}
        </InputLabelTitle>
        <Button onClick={() => cancelAllScanJobs()} disabled={cancellingAll}>
          {t('settings.scanner.queue_status.cancel_all', 'Cancel all')}
        </Button>
      </div>
      <ul className="text-sm mt-2 max-h-64 overflow-y-auto">
        {running.map(item => (
          <QueueRow
            key={item.album.id}
            item={item}
            statusLabel={t('settings.scanner.queue_status.running', 'Running')}
            statusClassName="text-green-600"
          />
        ))}
        {queued.map(item => (
          <QueueRow
            key={item.album.id}
            item={item}
            statusLabel={t('settings.scanner.queue_status.queued', 'Queued')}
            statusClassName="text-gray-400"
          />
        ))}
      </ul>
    </div>
  )
}

export default ScannerQueueStatus
