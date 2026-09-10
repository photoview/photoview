import { Dialog } from '@headlessui/react'
import { useQuery } from '@apollo/client'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import AlbumTree from './AlbumTree'
import { ReactComponent as AlbumTreeIcon } from '../album/icons/album-tree.svg'
import { authToken } from '../../helpers/authentication'
import { ALBUM_TREE_PREFERENCE_QUERY } from '../layout/Layout'
import { layoutAlbumTreePreferenceQuery } from '../layout/__generated__/layoutAlbumTreePreferenceQuery'

// Lets small screens reach the album tree as a full-screen overlay, since the
// persistent sidebar (Layout.tsx) is desktop-only (hidden below the `lg`
// breakpoint) for lack of space.
const MobileAlbumTreeButton = () => {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const token = authToken()
  const { data } = useQuery<layoutAlbumTreePreferenceQuery>(
    ALBUM_TREE_PREFERENCE_QUERY,
    { skip: !token }
  )
  const showAlbumTree = data?.myUserPreferences.showAlbumTree ?? true

  if (!showAlbumTree) return null

  return (
    <>
      <button
        type="button"
        title={t('album_tree.mobile_open', 'Browse album tree')}
        aria-label={t('album_tree.mobile_open', 'Browse album tree')}
        className="lg:hidden bg-gray-50 h-[30px] align-top px-2 py-1 rounded border border-gray-200 focus:outline-none focus:border-blue-300 text-[#8b8b8b] hover:bg-gray-100 hover:text-[#777] dark:bg-dark-input-bg dark:border-dark-input-border dark:text-dark-input-text dark:focus:border-blue-300"
        onClick={() => setOpen(true)}
      >
        <AlbumTreeIcon />
      </button>
      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        // Higher than Sidebar's z-[110] (ui/src/components/sidebar/Sidebar.tsx),
        // so this overlay isn't rendered behind it and left unclickable when
        // both are open on a small screen.
        className="fixed z-[115] inset-0 lg:hidden"
      >
        <Dialog.Overlay className="fixed inset-0 bg-black opacity-30" />
        <div
          className="fixed inset-y-0 right-0 w-full max-w-xs bg-white dark:bg-dark-bg shadow-md overflow-y-auto"
          onClick={e => {
            if ((e.target as HTMLElement).closest('a')) setOpen(false)
          }}
        >
          <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-dark-border">
            <Dialog.Title className="text-lg">
              {t('album_tree.label', 'Album tree')}
            </Dialog.Title>
            <button
              type="button"
              aria-label={t('general.close', 'Close')}
              onClick={() => setOpen(false)}
              className="text-2xl leading-none px-2 text-gray-500 hover:text-gray-800 dark:hover:text-gray-200"
            >
              &times;
            </button>
          </div>
          <AlbumTree />
        </div>
      </Dialog>
    </>
  )
}

export default MobileAlbumTreeButton
