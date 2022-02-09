import React, { useState, useEffect } from 'react'
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
import { authToken } from '../../helpers/authentication'

const RESET_ALBUM_COVER_MUTATION = gql`
  mutation resetAlbumCover($albumID: ID!) {
    resetAlbumCover(albumID: $albumID) {
      id
      thumbnail {
        id
        thumbnail {
          url
        }
      }
    }
  }
`
const SET_ALBUM_COVER_MUTATION = gql`
  mutation setAlbumCover($coverID: ID!) {
    setAlbumCover(coverID: $coverID) {
      id
      thumbnail {
        id
        thumbnail {
          url
        }
      }
    }
  }
`

type SidebarPhotoCoverProps = {
  cover_id: string
}

export const SidebarPhotoCover = ({ cover_id }: SidebarPhotoCoverProps) => {
  const { t } = useTranslation()

  const [setAlbumCover] = useMutation<setAlbumCover, setAlbumCoverVariables>(
    SET_ALBUM_COVER_MUTATION,
    {
      variables: {
        coverID: cover_id,
      },
    }
  )

  const [buttonDisabled, setButtonDisabled] = useState(false)

  useEffect(() => {
    setButtonDisabled(false)
  }, [cover_id])

  // hide when not authenticated
  if (!authToken()) {
    return null
  }

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.album_cover', 'Album cover')}
      </SidebarSectionTitle>
      <div>
        <table className="border-collapse w-full">
          <tfoot>
            <tr className="text-left border-gray-100 dark:border-dark-border2 border-b border-t">
              <td colSpan={2} className="pl-4 py-2">
                <button
                  className="disabled:opacity-50 text-green-500 font-bold uppercase text-xs"
                  disabled={buttonDisabled}
                  onClick={() => {
                    setButtonDisabled(true),
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

  const [resetAlbumCover] = useMutation<
    resetAlbumCover,
    resetAlbumCoverVariables
  >(RESET_ALBUM_COVER_MUTATION, {
    variables: {
      albumID: id,
    },
  })

  const [buttonDisabled, setButtonDisabled] = useState(false)

  useEffect(() => {
    setButtonDisabled(false)
  }, [id])

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.album_cover', 'Album cover')}
      </SidebarSectionTitle>
      <div>
        <table className="border-collapse w-full">
          <tfoot>
            <tr className="text-left border-gray-100 dark:border-dark-border2 border-b border-t">
              <td colSpan={2} className="pl-4 py-2">
                <button
                  className="disabled:opacity-50 text-red-500 font-bold uppercase text-xs"
                  disabled={buttonDisabled}
                  onClick={() => {
                    setButtonDisabled(true),
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
