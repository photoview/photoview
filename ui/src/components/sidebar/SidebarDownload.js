import React from 'react'
import PropTypes from 'prop-types'
import { Menu, Dropdown, Button } from 'semantic-ui-react'
import { MessageState } from '../messages/Messages'
import { Query, useLazyQuery } from 'react-apollo'
import gql from 'graphql-tag'
import download from 'downloadjs'

const downloadQuery = gql`
  query sidebarDownloadQuery($photoId: Int!) {
    photo(id: $photoId) {
      id
      downloads {
        title
        url
        width
        height
      }
    }
  }
`

function formatBytes(bytes) {
  var sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  if (bytes == 0) return '0 Byte'
  var i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)))
  return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i]
}

const downloadPhoto = async url => {
  const imgUrl = new URL(url)
  let headers = {
    Authorization: `Bearer ${localStorage.getItem('token')}`,
  }

  if (localStorage.getItem('token') == null) {
    // Get share token if not authorized
    const token = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
    if (token) {
      imgUrl.searchParams.set('token', token[1])
    }

    headers = {}
  }

  const response = await fetch(imgUrl.href, {
    headers,
  })

  const totalBytes = Number(response.headers.get('content-length'))
  console.log(totalBytes)

  if (totalBytes == 0) {
    MessageState.add({
      key: Math.random().toString(26),
      type: 'message',
      props: {
        header: 'Error downloading photo',
        content: `Could not get size of photo from server`,
        negative: true,
      },
    })
    return
  }

  const notifKey = Math.random().toString(26)
  MessageState.add({
    key: notifKey,
    type: 'progress',
    props: {
      header: 'Downloading photo',
      content: `Starting download`,
      progress: 0,
    },
  })

  const reader = response.body.getReader()
  let data = new Uint8Array(totalBytes)

  let receivedBytes = 0
  let result
  do {
    result = await reader.read()

    if (result.value) data.set(result.value, receivedBytes)

    receivedBytes += result.value ? result.value.length : 0

    MessageState.add({
      key: notifKey,
      type: 'progress',
      props: {
        header: 'Downloading photo',
        percent: (receivedBytes / totalBytes) * 100,
        content: `${formatBytes(receivedBytes)} of ${formatBytes(
          totalBytes
        )} bytes downloaded`,
      },
    })
  } while (!result.done)

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
  const filename = url.match(/[^/]*$/)[0]

  download(content, filename)
}

const SidebarDownload = ({ photo }) => {
  if (!photo || !photo.id) return null

  const [
    loadPhotoDownloads,
    { called, loading, data },
  ] = useLazyQuery(downloadQuery, { variables: { photoId: photo.id } })

  let downloads = []

  if (called) {
    if (!loading) {
      downloads = data && data.photo.downloads
    }
  } else {
    if (!photo.downloads) {
      loadPhotoDownloads()
    } else {
      downloads = photo.downloads
    }
  }

  let buttons = downloads.map(x => (
    <Button
      style={{ marginTop: 4 }}
      key={x.url}
      onClick={() => downloadPhoto(x.url)}
    >
      {`${x.title} (${x.width} x ${x.height})`}
    </Button>
  ))

  return (
    <div style={{ marginBottom: 24 }}>
      <h2>Download</h2>
      <div>{buttons}</div>
    </div>
  )
}

SidebarDownload.propTypes = {
  photo: PropTypes.object,
}

export default SidebarDownload
