import React from 'react'
import { useMutation, gql } from '@apollo/client'
import { useTranslation } from 'react-i18next'

import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'

import {
  setAlbumCover,
  setAlbumCoverVariables,
} from './__generated__/setAlbumCover'
import {
  resetAlbumCover,
  resetAlbumCoverVariables,
} from './__generated__/resetAlbumCover'

const RESET_ALBUM_COVER_MUTATION = gql`
  mutation resetAlbumCover($albumID: ID!) {
    resetAlbumCover(albumID: $albumID) {
      id
    }
  }
`
const SET_ALBUM_COVER_MUTATION = gql`
  mutation setAlbumCover($coverID: ID!) {
    setAlbumCover(coverID: $coverID) {
      coverID
    }
  }
`

type SidebarPhotoCoverProps = {
  cover_id: string
}

export const SidebarPhotoCover = ({ cover_id }: SidebarPhotoCoverProps) => {
  const { t } = useTranslation()

  const [setAlbumCover, { loading }] = useMutation<
    setAlbumCover,
    setAlbumCoverVariables
  >(SET_ALBUM_COVER_MUTATION, {
    variables: {
      coverID: cover_id,
    },
  })

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.cover_photo', 'Album cover')}
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
                    setAlbumCover({
                      variables: {
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

  const [resetAlbumCover, { loading }] = useMutation<
    resetAlbumCover,
    resetAlbumCoverVariables
  >(RESET_ALBUM_COVER_MUTATION, {
    variables: {
      albumID: id,
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
                    resetAlbumCover({
                      variables: {
                        albumID: id,
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
