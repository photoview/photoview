import React from 'react'
import PropTypes from 'prop-types'
import { Table } from 'semantic-ui-react'
import styled from 'styled-components'
import { MessageState } from '../messages/Messages'
import { useLazyQuery, gql } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { useTranslation } from 'react-i18next'
import { TranslationFn } from '../../localization'
import { MediaSidebarMedia } from './MediaSidebar'
import {
  sidebarDownloadQuery,
  sidebarDownloadQueryVariables,
  sidebarDownloadQuery_media_downloads,
} from './__generated__/sidebarDownloadQuery'

export const SIDEBAR_DOWNLOAD_QUERY = gql`
  query sidebarDownloadQuery($mediaId: ID!) {
    media(id: $mediaId) {
      id
      downloads {
        title
        mediaUrl {
          url
          width
          height
          fileSize
        }
      }
    }
  }
`

const formatBytes = (t: TranslationFn) => (bytes: number) => {
  if (bytes == 0)
    return t('sidebar.download.filesize.byte', '{{count}} Byte', { count: 0 })

  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const count = Math.round(bytes / Math.pow(1024, i))

  switch (i) {
    case 0:
      // i18next-extract-mark-plural-next-line
      return t('sidebar.download.filesize.byte', '{{count}} Byte', { count })
    case 1:
      return t('sidebar.download.filesize.kilo_byte', '{{count}} KB', { count })
    case 2:
      return t('sidebar.download.filesize.mega_byte', '{{count}} MB', { count })
    case 3:
      return t('sidebar.download.filesize.giga_byte', '{{count}} GB', { count })
    case 4:
      return t('sidebar.download.filesize.tera_byte', '{{count}} TB', { count })
    default:
      return new Error(`invalid byte value: ${bytes}`)
  }
}

const downloadMedia = (t: TranslationFn) => async (url: string) => {
  const imgUrl = new URL(url, location.origin)

  if (authToken() == null) {
    // Get share token if not authorized
    const token = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
    if (token) {
      imgUrl.searchParams.set('token', token[1])
    }
  }

  const response = await fetch(imgUrl.href, {
    credentials: 'include',
  })

  let blob = null
  if (response.headers.has('content-length')) {
    blob = await downloadMediaShowProgress(t)(response)
  } else {
    blob = await response.blob()
  }

  if (blob == null) {
    console.log('Blob is null canceling')
    return
  }

  const filenameMatch = url.match(/[^/]*$/)

  if (filenameMatch == null) {
    console.error('Could not extract filename', url)
    return
  }

  const filename = filenameMatch[0]
  downloadBlob(blob, filename)
}

const downloadMediaShowProgress = (t: TranslationFn) => async (
  response: Response
) => {
  const totalBytes = Number(response.headers.get('content-length'))
  const reader = response.body?.getReader()
  const data = new Uint8Array(totalBytes)

  if (reader == null) {
    throw new Error('Download reader is null')
  }

  let canceled = false
  const onDismiss = () => {
    canceled = true
    reader.cancel('Download canceled by user')
  }

  const notifyKey = Math.random().toString(26)
  MessageState.add({
    key: notifyKey,
    type: 'progress',
    onDismiss,
    props: {
      header: 'Downloading photo',
      content: `Starting download`,
      progress: 0,
    },
  })

  let receivedBytes = 0
  let result
  do {
    result = await reader.read()

    if (canceled) break

    if (result.value) data.set(result.value, receivedBytes)

    receivedBytes += result.value ? result.value.length : 0

    MessageState.add({
      key: notifyKey,
      type: 'progress',
      onDismiss,
      props: {
        header: 'Downloading photo',
        percent: (receivedBytes / totalBytes) * 100,
        content: `${formatBytes(t)(receivedBytes)} of ${formatBytes(t)(
          totalBytes
        )} bytes downloaded`,
      },
    })
  } while (!result.done)

  if (canceled) {
    return
  }

  MessageState.add({
    key: notifyKey,
    type: 'progress',
    props: {
      header: 'Downloading photo completed',
      content: `The photo has been downloaded`,
      percent: 100,
      positive: true,
    },
  })

  setTimeout(() => {
    MessageState.removeKey(notifyKey)
  }, 2000)

  const content = new Blob([data.buffer], {
    type: response.headers.get('content-type') || undefined,
  })

  return content
}

const downloadBlob = async (blob: Blob, filename: string) => {
  const objectUrl = window.URL.createObjectURL(blob)

  const anchor = document.createElement('a')
  document.body.appendChild(anchor)

  anchor.href = objectUrl
  anchor.download = filename
  anchor.click()

  anchor.remove()

  window.URL.revokeObjectURL(objectUrl)
}

const DownloadTableRow = styled(Table.Row)`
  cursor: pointer;
`

type SidebarDownladProps = {
  media: MediaSidebarMedia
}

const SidebarDownload = ({ media }: SidebarDownladProps) => {
  const { t } = useTranslation()
  if (!media || !media.id) return null

  const [loadPhotoDownloads, { called, loading, data }] = useLazyQuery<
    sidebarDownloadQuery,
    sidebarDownloadQueryVariables
  >(SIDEBAR_DOWNLOAD_QUERY, { variables: { mediaId: media.id } })

  let downloads: sidebarDownloadQuery_media_downloads[] = []

  if (called) {
    if (!loading) {
      downloads = (data && data.media.downloads) || []
    }
  } else {
    if (!media.downloads) {
      loadPhotoDownloads()
    } else {
      downloads = media.downloads
    }
  }

  const extractExtension = (url: string) => {
    const urlMatch = url.split(/[#?]/)
    if (urlMatch == null) return

    return urlMatch[0].split('.').pop()?.trim().toLowerCase()
  }

  const download = downloadMedia(t)
  const bytes = formatBytes(t)
  const downloadRows = downloads.map(x => (
    <DownloadTableRow
      key={x.mediaUrl.url}
      onClick={() => download(x.mediaUrl.url)}
    >
      <Table.Cell>{`${x.title}`}</Table.Cell>
      <Table.Cell>{`${x.mediaUrl.width} x ${x.mediaUrl.height}`}</Table.Cell>
      <Table.Cell>{`${bytes(x.mediaUrl.fileSize)}`}</Table.Cell>
      <Table.Cell>{extractExtension(x.mediaUrl.url)}</Table.Cell>
    </DownloadTableRow>
  ))

  return (
    <div style={{ marginBottom: 24 }}>
      <h2>{t('sidebar.download.title', 'Download')}</h2>

      <Table selectable singleLine compact>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>
              {t('sidebar.download.table_columns.name', 'Name')}
            </Table.HeaderCell>
            <Table.HeaderCell>
              {t('sidebar.download.table_columns.dimensions', 'Dimensions')}
            </Table.HeaderCell>
            <Table.HeaderCell>
              {t('sidebar.download.table_columns.file_size', 'Size')}
            </Table.HeaderCell>
            <Table.HeaderCell>
              {t('sidebar.download.table_columns.file_type', 'Type')}
            </Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>{downloadRows}</Table.Body>
      </Table>
    </div>
  )
}

SidebarDownload.propTypes = {
  photo: PropTypes.object,
}

export default SidebarDownload
