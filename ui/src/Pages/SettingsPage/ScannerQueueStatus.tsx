import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import { InputLabelTitle } from './SettingsPage'
import { ScannerJobStatus } from '../../__generated__/globalTypes'
import {
  scannerQueueStatusQuery,
  scannerQueueStatusQuery_scannerQueueStatus,
} from './__generated__/scannerQueueStatusQuery'

const SCANNER_QUEUE_STATUS_QUERY = gql`
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

const albumBreadcrumb = (
  item: scannerQueueStatusQuery_scannerQueueStatus
): string => {
  const ancestors = item.album.path
    .slice()
    .reverse()
    .map(a => a.title)
  return [...ancestors, item.album.title].join(' / ')
}

const ScannerQueueStatus = () => {
  const { t } = useTranslation()
  const { data } = useQuery<scannerQueueStatusQuery>(
    SCANNER_QUEUE_STATUS_QUERY,
    { pollInterval: 2000 }
  )

  const items = data?.scannerQueueStatus ?? []
  if (items.length === 0) return null

  const running = items.filter(i => i.status === ScannerJobStatus.RUNNING)
  const queued = items.filter(i => i.status === ScannerJobStatus.QUEUED)

  return (
    <div>
      <InputLabelTitle>
        {t('settings.scanner.queue_status.title', 'Scan queue')}
      </InputLabelTitle>
      <ul className="text-sm mt-2 max-h-64 overflow-y-auto">
        {running.map(item => (
          <li
            key={item.album.id}
            className="flex justify-between gap-4 py-1 border-b border-gray-100 dark:border-dark-border2"
          >
            <span className="truncate">{albumBreadcrumb(item)}</span>
            <span className="text-green-600 shrink-0">
              {t('settings.scanner.queue_status.running', 'Running')}
            </span>
          </li>
        ))}
        {queued.map(item => (
          <li
            key={item.album.id}
            className="flex justify-between gap-4 py-1 border-b border-gray-100 dark:border-dark-border2"
          >
            <span className="truncate">{albumBreadcrumb(item)}</span>
            <span className="text-gray-400 shrink-0">
              {t('settings.scanner.queue_status.queued', 'Queued')}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}

export default ScannerQueueStatus
