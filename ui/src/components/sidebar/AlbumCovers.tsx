import React from 'react'
import { useMutation, gql } from '@apollo/client'
import { useTranslation } from 'react-i18next'

import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'

import {
  setAlbumCoverID,
  setAlbumCoverIDVariables,
} from './__generated__/setAlbumCoverID'

const SET_ALBUM_COVER_ID_MUTATION = gql`
  mutation setAlbumCoverID($albumID: ID!, $coverID: Int!) {
    setAlbumCoverID(albumID: $albumID, coverID: $coverID) {
      id
      coverID
    }
  }
`

type SidebarPhotoCoverProps = {
  id: string
  cover_id: string
}

export const SidebarPhotoCover = ({ id, cover_id }: SidebarPhotoCoverProps) => {
  const { t } = useTranslation()

  const [setAlbumCoverID, { loading }] = useMutation<
    setAlbumCoverID,
    setAlbumCoverIDVariables
  >(SET_ALBUM_COVER_ID_MUTATION, {
    variables: {
      albumID: id,
      coverID: cover_id,
    },
  })

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.cover_photo', 'Cover Photo')}
      </SidebarSectionTitle>
      <div>
        <table className="border-collapse w-full">
          <tfoot>
            <tr className="text-left border-gray-100 border-b border-t">
              <td colSpan={2} className="pl-4 py-2">
                <button
                  className="text-green-500 font-bold uppercase text-xs"
                  disabled={loading}
                  onClick={() => {
                    setAlbumCoverID({
                      variables: {
                        albumID: id,
                        coverID: cover_id,
                      },
                    })
                  }}
                >
                  <span>
                    {t('sidebar.album.set_cover', 'Set as album cover photo')}
                  </span>
                </button>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </SidebarSection>
  )
}

type SidebarAlbumCoverProps = {
  id: string
}

export const SidebarAlbumCover = ({ id }: SidebarAlbumCoverProps) => {
  const { t } = useTranslation()

  const [setAlbumCoverID, { loading }] = useMutation<
    setAlbumCoverID,
    setAlbumCoverIDVariables
  >(SET_ALBUM_COVER_ID_MUTATION, {
    variables: {
      albumID: id,
      coverID: '-1',
    },
  })

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.album_cover', 'Album cover')}
      </SidebarSectionTitle>
      <div>
        <table className="border-collapse w-full">
          <tfoot>
            <tr className="text-left border-gray-100 border-b border-t">
              <td colSpan={2} className="pl-4 py-2">
                <button
                  className="text-red-500 font-bold uppercase text-xs"
                  disabled={loading}
                  onClick={() => {
                    setAlbumCoverID({
                      variables: {
                        albumID: id,
                        coverID: '-1',
                      },
                    })
                  }}
                >
                  <span>
                    {t('sidebar.album.reset_cover', 'Reset cover photo')}
                  </span>
                </button>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </SidebarSection>
  )
}
