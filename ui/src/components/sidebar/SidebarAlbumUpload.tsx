import React, { useRef, useState } from 'react'
import { useApolloClient } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import urlJoin from 'url-join'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import { Button } from '../../primitives/form/Input'
import { API_ENDPOINT } from '../../apolloClient'
import { MessageState } from '../messages/Messages'
import { NotificationType } from '../../__generated__/globalTypes'

type UploadFileResult = {
  path: string
  status: 'ok' | 'rejected' | 'error'
  reason?: string
}

type UploadResponse = {
  results: UploadFileResult[]
}

type SidebarAlbumUploadProps = {
  albumId: string
}

const relativePathOf = (file: File): string =>
  // webkitRelativePath is only set when the file was picked via a
  // directory input, and mirrors the path chosen relative to that folder.
  (file as File & { webkitRelativePath?: string }).webkitRelativePath ||
  file.name

const SidebarAlbumUpload = ({ albumId }: SidebarAlbumUploadProps) => {
  const { t } = useTranslation()
  const client = useApolloClient()
  const [uploading, setUploading] = useState(false)
  const filesInputRef = useRef<HTMLInputElement>(null)
  const folderInputRef = useRef<HTMLInputElement | null>(null)

  const upload = (fileList: FileList | null) => {
    if (!fileList || fileList.length === 0) return

    const formData = new FormData()
    for (const file of Array.from(fileList)) {
      formData.append(relativePathOf(file), file)
    }

    const notifyKey = `upload-${Date.now()}`
    setUploading(true)

    const xhr = new XMLHttpRequest()
    xhr.open('POST', urlJoin(API_ENDPOINT, '/upload', albumId))
    xhr.withCredentials = true

    xhr.upload.onprogress = event => {
      if (!event.lengthComputable) return
      MessageState.add({
        key: notifyKey,
        type: NotificationType.Progress,
        onDismiss: () => xhr.abort(),
        props: {
          header: t('sidebar.album.upload.uploading', 'Uploading'),
          percent: (event.loaded / event.total) * 100,
          content: `${Math.round(event.loaded / 1024)} / ${Math.round(
            event.total / 1024
          )} KB`,
        },
      })
    }

    // Takes an already-translated header, rather than a key, so every t()
    // call site below is a literal i18next-parser can extract - a dynamic
    // key here silently drops these strings from the extracted
    // translations.
    const finish = (header: string, content = '') => {
      setUploading(false)
      MessageState.add({
        key: notifyKey,
        type: NotificationType.Message,
        props: { header, content },
      })
      setTimeout(() => MessageState.removeKey(notifyKey), 4000)
    }

    xhr.onload = () => {
      if (xhr.status !== 200) {
        finish(
          t('sidebar.album.upload.failed', 'Upload failed'),
          xhr.responseText
        )
        return
      }

      let rejected: UploadFileResult[] | null = null
      try {
        const parsed = JSON.parse(xhr.responseText) as UploadResponse
        if (Array.isArray(parsed.results)) {
          rejected = parsed.results.filter(r => r.status !== 'ok')
        }
      } catch {
        // Handled below, same as a well-formed body without results.
      }

      if (rejected === null) {
        // A 200 whose body we can't read tells us nothing about which files
        // were stored - reporting "Upload complete" here would hide files
        // the server actually rejected. Some may still have landed, so the
        // cache invalidation below has to run either way.
        finish(
          t('sidebar.album.upload.failed', 'Upload failed'),
          xhr.responseText
        )
      } else if (rejected.length > 0) {
        finish(
          t(
            'sidebar.album.upload.finished_with_errors',
            'Upload finished with errors'
          ),
          rejected.map(r => `${r.path}: ${r.reason || r.status}`).join('\n')
        )
      } else {
        finish(t('sidebar.album.upload.complete', 'Upload complete'))
      }

      // Media/sub-albums arrive asynchronously once the server-side rescan
      // finishes, so just invalidate the cache rather than trying to show
      // them immediately.
      client.refetchQueries({ include: ['albumQuery'] })
    }

    xhr.onerror = () => {
      finish(t('sidebar.album.upload.failed', 'Upload failed'))
    }

    // abort() fires the abort event, not error - onload never runs either,
    // so without this the upload buttons would stay disabled forever after
    // dismissing the progress toast.
    xhr.onabort = () => {
      finish(t('sidebar.album.upload.cancelled', 'Upload cancelled'))
    }

    xhr.send(formData)
  }

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.upload.title', 'Upload')}
      </SidebarSectionTitle>
      <div className="mx-4 flex gap-2 flex-wrap">
        <input
          ref={filesInputRef}
          type="file"
          multiple
          accept="image/*,video/*"
          className="hidden"
          onChange={e => {
            upload(e.target.files)
            e.target.value = ''
          }}
        />
        <Button
          disabled={uploading}
          onClick={() => filesInputRef.current?.click()}
        >
          {t('sidebar.album.upload.select_files', 'Select files')}
        </Button>

        {/* webkitdirectory/directory are non-standard attributes not in
            React's typed InputHTMLAttributes, so they're set imperatively. */}
        <input
          type="file"
          multiple
          className="hidden"
          onChange={e => {
            upload(e.target.files)
            e.target.value = ''
          }}
          ref={el => {
            folderInputRef.current = el
            if (el) {
              el.setAttribute('webkitdirectory', '')
              el.setAttribute('directory', '')
            }
          }}
        />
        <Button
          disabled={uploading}
          onClick={() => folderInputRef.current?.click()}
        >
          {t('sidebar.album.upload.select_folder', 'Select folder')}
        </Button>
      </div>
    </SidebarSection>
  )
}

export default SidebarAlbumUpload
