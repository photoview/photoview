import { gql, useLazyQuery } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import { NotificationType } from '../../__generated__/globalTypes'
import { authToken } from '../../helpers/authentication'
import { TranslationFn } from '../../localization'
import { MessageState } from '../messages/Messages'
import { MediaSidebarMedia } from './MediaSidebar/MediaSidebar'
import React, { useRef, useState } from 'react'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import SidebarTable from './SidebarTable'
import { ReactComponent as ShareIcon } from './icons/shareNativeIcon.svg'
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
      throw new Error(`invalid byte value: ${bytes}`)
  }
}

const fetchMediaResponse = async (url: string): Promise<Response> => {
  const imgUrl = new URL(
    `${import.meta.env.BASE_URL}${url}`.replace(/\/\//g, '/'),
    location.origin
  )

  if (authToken() == null) {
    // Get share token if not authorized
    const token = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
    if (token) {
      imgUrl.searchParams.set('token', token[1])
    }
  }

  return fetch(imgUrl.href, {
    credentials: 'include',
  })
}

export const fetchMediaBlob =
  (t: TranslationFn) =>
  async (url: string): Promise<Blob | null | undefined> => {
    const response = await fetchMediaResponse(url)

    // An error body (401/403/404) is still a readable body - without this it
    // would be handed to the progress reader or saved as if it were media.
    if (!response.ok) {
      console.error(`Failed to fetch media: ${response.status}`)
      return null
    }

    if (response.headers.has('content-length')) {
      return downloadMediaShowProgress(t)(response)
    }

    return response.blob()
  }

// Like fetchMediaBlob, but never shows the download-progress notification —
// dismissing that notification cancels the fetch, which isn't appropriate
// for the share flow.
export const fetchMediaBlobQuiet = async (url: string): Promise<Blob> => {
  const response = await fetchMediaResponse(url)
  if (!response.ok) {
    throw new Error(`Failed to fetch media: ${response.status}`)
  }
  return response.blob()
}

const downloadMedia = (t: TranslationFn) => async (url: string) => {
  const blob = await fetchMediaBlob(t)(url)

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

const downloadMediaShowProgress =
  (t: TranslationFn) => async (response: Response) => {
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
      type: NotificationType.Progress,
      onDismiss,
      props: {
        header: 'Downloading photo',
        content: `Starting download`,
        percent: 0,
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
        type: NotificationType.Progress,
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
      type: NotificationType.Progress,
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

const downloadBlob = (blob: Blob, filename: string) => {
  const objectUrl = window.URL.createObjectURL(blob)

  const anchor = document.createElement('a')
  document.body.appendChild(anchor)

  anchor.href = objectUrl
  anchor.download = filename
  anchor.click()

  anchor.remove()

  window.URL.revokeObjectURL(objectUrl)
}

type SidebarDownloadTableRow = {
  title: string
  url: string
  width: number
  height: number
  fileSize: number
}

type SidebarDownloadTableProps = {
  rows: SidebarDownloadTableRow[]
}

const SidebarDownloadTable = ({ rows }: SidebarDownloadTableProps) => {
  const { t } = useTranslation()

  const extractExtension = (url: string) => {
    const urlMatch = url.split(/[#?]/)
    if (urlMatch == null) return

    return urlMatch[0].split('.').pop()?.trim().toLowerCase()
  }

  const download = downloadMedia(t)
  const bytes = formatBytes(t)
  const downloadRows = rows.map(x => (
    <SidebarTable.Row key={x.url} onClick={() => download(x.url)} tabIndex={0}>
      <td className="pl-4 py-2">{`${x.title}`}</td>
      <td className="py-2">{`${x.width} x ${x.height}`}</td>
      <td className="py-2">{`${bytes(x.fileSize)}`}</td>
      <td className="pr-4 py-2">{extractExtension(x.url)}</td>
    </SidebarTable.Row>
  ))

  return (
    <SidebarTable.Table>
      <SidebarTable.Head>
        <SidebarTable.HeadRow>
          <th className="w-2/6 pl-4 py-2">
            {t('sidebar.download.table_columns.name', 'Name')}
          </th>
          <th className="w-2/6 py-2">
            {t('sidebar.download.table_columns.dimensions', 'Dimensions')}
          </th>
          <th className="w-1/6 py-2">
            {t('sidebar.download.table_columns.file_size', 'Size')}
          </th>
          <th className="w-1/6 pr-4 py-2">
            {t('sidebar.download.table_columns.file_type', 'Type')}
          </th>
        </SidebarTable.HeadRow>
      </SidebarTable.Head>
      <tbody>{downloadRows}</tbody>
    </SidebarTable.Table>
  )
}

const canNativeShare = () =>
  typeof navigator !== 'undefined' && typeof navigator.share === 'function'

const pickShareRow = (rows: SidebarDownloadTableRow[]) =>
  rows.find(x => x.title == 'Web optimized video') ??
  rows.find(x => x.title == 'Original') ??
  rows.find(x => x.title == 'Large') ??
  rows[0]

type SidebarShareMediaButtonProps = {
  media: MediaSidebarMedia
  rows: SidebarDownloadTableRow[]
}

export const SidebarShareMediaButton = ({
  media,
  rows,
}: SidebarShareMediaButtonProps) => {
  const { t } = useTranslation()
  const [sharing, setSharing] = useState(false)
  const [retry, setRetry] = useState(false)
  // A file already prepared for a share that the browser then refused. Keeping
  // it means the retry needs no download and so stays inside its activation.
  const preparedFile = useRef<{ url: string; file: File } | null>(null)

  const row = pickShareRow(rows)

  if (!canNativeShare() || row == null) return null

  const share = async () => {
    setSharing(true)
    // Sharing a link is the fallback whenever the OS share sheet can't take
    // the file itself (desktop browsers without file support). Decide that
    // before downloading anything: a browser exposing share() without
    // canShare() gives no way to know it takes files, and a large or failing
    // download would otherwise delay a share it was never needed for.
    const shareLink = () =>
      navigator.share({
        title: media.title ?? undefined,
        url: location.href,
      })

    try {
      if (!navigator.canShare) {
        await shareLink()
        return
      }

      let file = preparedFile.current?.file
      if (preparedFile.current?.url !== row.url) {
        const blob = await fetchMediaBlobQuiet(row.url)

        const filename = row.url.match(/[^/]*$/)?.[0] ?? media.title ?? 'photo'
        file = new File([blob], filename, { type: blob.type })
      }
      if (file == null) return

      // Kept before the canShare check rather than after it: the link fallback
      // below shares the download that may have cost us the activation, so it
      // can fail the same way. Holding the file from here on means the retry
      // skips that download whichever of the two branches it lands in.
      preparedFile.current = { url: row.url, file }

      if (!navigator.canShare({ files: [file] })) {
        await shareLink()
        preparedFile.current = null
        setRetry(false)
        return
      }

      await navigator.share({ files: [file], title: media.title ?? undefined })
      preparedFile.current = null
      setRetry(false)
    } catch (err) {
      if ((err as Error)?.name === 'AbortError') {
        preparedFile.current = null
        setRetry(false)
        return
      }

      console.error('Native share failed', err)

      // A download slow enough to outlast the user activation leaves the share
      // sheet refusing to open, and the link fallback below has no activation
      // left either - so the tap would end in nothing at all. The file is
      // ready now, though, and the kept copy means a second tap opens the
      // sheet straight away. Asking for that tap only here keeps the common
      // case at one.
      if ((err as Error)?.name === 'NotAllowedError' && preparedFile.current) {
        setRetry(true)

        return
      }

      // The file couldn't be prepared or the sheet refused it for some other
      // reason, but the link is still shareable, so try that rather than
      // ending in nothing.
      try {
        await shareLink()
      } catch (fallbackErr) {
        if ((fallbackErr as Error)?.name !== 'AbortError') {
          console.error('Link share failed too', fallbackErr)
        }
      }
    } finally {
      setSharing(false)
    }
  }

  return (
    <div className="pl-4 py-2">
      <button
        className="text-green-500 font-bold uppercase text-xs disabled:opacity-50"
        disabled={sharing}
        onClick={share}
      >
        <ShareIcon className="inline-block mr-2" />
        <span>
          {retry
            ? t('sidebar.download.share_again', 'Tap share again')
            : t('sidebar.download.share', 'Share')}
        </span>
      </button>
    </div>
  )
}

type SidebarMediaDownladProps = {
  media: MediaSidebarMedia
}

const SidebarMediaDownload = ({ media }: SidebarMediaDownladProps) => {
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

  const downloadRows = downloads.map<SidebarDownloadTableRow>(x => ({
    title: x.title,
    url: x.mediaUrl.url,
    width: x.mediaUrl.width,
    height: x.mediaUrl.height,
    fileSize: x.mediaUrl.fileSize,
  }))

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.download.title', 'Download')}
      </SidebarSectionTitle>

      <SidebarDownloadTable rows={downloadRows} />
      <SidebarShareMediaButton media={media} rows={downloadRows} />
    </SidebarSection>
  )
}

export default SidebarMediaDownload
