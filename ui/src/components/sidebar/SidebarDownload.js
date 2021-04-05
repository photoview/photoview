import React from 'react'
import PropTypes from 'prop-types'
import { Table } from 'semantic-ui-react'
import styled from 'styled-components'
import { MessageState } from '../messages/Messages'
import { useLazyQuery, gql } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { useTranslation } from 'react-i18next'

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

const formatBytes = t => bytes => {
  if (bytes == 0) return '0 Byte'
  const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)))
  const count = Math.round(bytes / Math.pow(1024, i), 2)

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

const downloadMedia = t => async url => {
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

  const filename = url.match(/[^/]*$/)[0]

  downloadBlob(blob, filename)
}

const downloadMediaShowProgress = t => async response => {
  const totalBytes = Number(response.headers.get('content-length'))
  const reader = response.body.getReader()
  let data = new Uint8Array(totalBytes)

  let canceled = false
  const onDismiss = () => {
    canceled = true
    reader.cancel('Download canceled by user')
  }

  const notifKey = Math.random().toString(26)
  MessageState.add({
    key: notifKey,
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
      key: notifKey,
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
    key: notifKey,
    type: 'progress',
    props: {
      header: 'Downloading photo completed',
      content: `The photo has been downloaded`,
      percent: 100,
      positive: true,
    },
  })

  setTimeout(() => {
    MessageState.removeKey(notifKey)
  }, 2000)

  const content = new Blob([data.buffer], {
    type: response.headers.get('content-type'),
  })

  return content
}

const downloadBlob = async (blob, filename) => {
  let objectUrl = window.URL.createObjectURL(blob)

  let anchor = document.createElement('a')
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

const SidebarDownload = ({ photo }) => {
  const { t } = useTranslation()
  if (!photo || !photo.id) return null

  const [
    loadPhotoDownloads,
    { called, loading, data },
  ] = useLazyQuery(SIDEBAR_DOWNLOAD_QUERY, { variables: { mediaId: photo.id } })

  let downloads = []

  if (called) {
    if (!loading) {
      downloads = data && data.media.downloads
    }
  } else {
    if (!photo.downloads) {
      loadPhotoDownloads()
    } else {
      downloads = photo.downloads
    }
  }

  const extractExtension = url => {
    return url.split(/[#?]/)[0].split('.').pop().trim().toLowerCase()
  }

  const download = downloadMedia(t)
  const bytes = formatBytes(t)
  let downloadRows = downloads.map(x => (
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
