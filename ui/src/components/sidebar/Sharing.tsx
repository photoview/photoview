import React, { useEffect, useState } from 'react'
import {
  useMutation,
  useQuery,
  gql,
  useLazyQuery,
  DocumentNode,
} from '@apollo/client'
import copy from 'copy-to-clipboard'
import { useTranslation } from 'react-i18next'
import { Popover } from '@headlessui/react'
import {
  sidebareDeleteShare,
  sidebareDeleteShareVariables,
} from './__generated__/sidebareDeleteShare'
import {
  sidebarPhotoAddShare,
  sidebarPhotoAddShareVariables,
} from './__generated__/sidebarPhotoAddShare'
import {
  sidebarAlbumAddShare,
  sidebarAlbumAddShareVariables,
} from './__generated__/sidebarAlbumAddShare'
import {
  sidebarGetPhotoShares,
  sidebarGetPhotoSharesVariables,
  sidebarGetPhotoShares_media_shares,
} from './__generated__/sidebarGetPhotoShares'
import {
  sidebarGetAlbumShares,
  sidebarGetAlbumSharesVariables,
  sidebarGetAlbumShares_album_shares,
} from './__generated__/sidebarGetAlbumShares'
import { authToken } from '../../helpers/authentication'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'

import { ReactComponent as LinkIcon } from './icons/shareLinkIcon.svg'
import { ReactComponent as CopyIcon } from './icons/shareCopyIcon.svg'
import { ReactComponent as DeleteIcon } from './icons/shareDeleteIcon.svg'
import { ReactComponent as MoreIcon } from './icons/shareMoreIcon.svg'
import { ReactComponent as AddIcon } from './icons/shareAddIcon.svg'
import Checkbox from '../../primitives/form/Checkbox'
import { TextField } from '../../primitives/form/Input'
import styled from 'styled-components'
import {
  sidebarProtectShare,
  sidebarProtectShareVariables,
} from './__generated__/sidebarProtectShare'

const SHARE_PHOTO_QUERY = gql`
  query sidebarGetPhotoShares($id: ID!) {
    media(id: $id) {
      id
      shares {
        id
        token
        hasPassword
      }
    }
  }
`

const SHARE_ALBUM_QUERY = gql`
  query sidebarGetAlbumShares($id: ID!) {
    album(id: $id) {
      id
      shares {
        id
        token
        hasPassword
      }
    }
  }
`

const ADD_MEDIA_SHARE_MUTATION = gql`
  mutation sidebarPhotoAddShare($id: ID!, $password: String, $expire: Time) {
    shareMedia(mediaId: $id, password: $password, expire: $expire) {
      token
    }
  }
`

const ADD_ALBUM_SHARE_MUTATION = gql`
  mutation sidebarAlbumAddShare($id: ID!, $password: String, $expire: Time) {
    shareAlbum(albumId: $id, password: $password, expire: $expire) {
      token
    }
  }
`

const PROTECT_SHARE_MUTATION = gql`
  mutation sidebarProtectShare($token: String!, $password: String) {
    protectShareToken(token: $token, password: $password) {
      token
      hasPassword
    }
  }
`

const DELETE_SHARE_MUTATION = gql`
  mutation sidebareDeleteShare($token: String!) {
    deleteShareToken(token: $token) {
      token
    }
  }
`

export const ArrowPopoverPanel = styled.div.attrs({
  className:
    'absolute -top-3 bg-white dark:bg-dark-bg rounded shadow-md border border-gray-200 dark:border-dark-border z-10',
})<{ width: number; flipped?: boolean }>`
  width: ${({ width }) => width}px;

  ${({ flipped }) =>
    flipped
      ? `
      left: 32px;
        `
      : `
      right: 24px;
    `}

  &::after {
    content: '';
    position: absolute;
    top: 18px;
    width: 8px;
    height: 14px;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 8 14'%3E%3Cpolyline stroke-width='1' stroke='%23E2E2E2' fill='%23FFFFFF' points='1 0 7 7 1 14'%3E%3C/polyline%3E%3C/svg%3E");

    ${({ flipped }) =>
      flipped
        ? `
      left: -7px;
      transform: rotate(180deg);
        `
        : `
      right: -7px;
    `}
  }
`

type MorePopoverSectionPasswordProps = {
  share: sidebarGetAlbumShares_album_shares
  query: DocumentNode
  id: string
}

const MorePopoverSectionPassword = ({
  share,
  query,
  id,
}: MorePopoverSectionPasswordProps) => {
  const [addingPassword, setAddingPassword] = useState(false)
  const activated = addingPassword || share.hasPassword

  const [passwordInputValue, setPasswordInputValue] = useState(
    share.hasPassword ? '**********' : ''
  )
  const [passwordHidden, setPasswordHidden] = useState(share.hasPassword)

  const [setPassword, { loading: setPasswordLoading }] = useMutation<
    sidebarProtectShare,
    sidebarProtectShareVariables
  >(PROTECT_SHARE_MUTATION, {
    refetchQueries: [{ query: query, variables: { id } }],
    onCompleted: data => {
      hidePassword(data.protectShareToken.hasPassword)
    },
    // refetchQueries: [{ query: query, variables: { id } }],
    variables: {
      token: share.token,
    },
  })

  const hidePassword = (hide: boolean) => {
    if (hide) {
      setPasswordInputValue('**********')
    }

    if (passwordHidden && !hide) {
      setPasswordInputValue('')
    }

    setPasswordHidden(hide)
  }

  const checkboxChange = () => {
    const enable = !activated
    setAddingPassword(enable)
    if (!enable) {
      setPassword({
        variables: {
          token: share.token,
          password: null,
        },
      })
      setPasswordInputValue('')
    }
  }

  const updatePasswordAction = () => {
    if (!passwordHidden && passwordInputValue != '') {
      setPassword({
        variables: {
          token: share.token,
          password: passwordInputValue,
        },
      })
    }
  }

  return (
    <div className="px-4 py-2">
      <Checkbox
        label="Password protected"
        checked={activated}
        onChange={checkboxChange}
      />
      <TextField
        disabled={!activated}
        type={passwordHidden ? 'password' : 'text'}
        value={passwordInputValue}
        className="mt-2 w-full"
        onKeyDown={event => {
          if (
            event.shiftKey ||
            event.altKey ||
            event.ctrlKey ||
            event.metaKey ||
            event.key == 'Enter' ||
            event.key == 'Tab' ||
            event.key == 'Escape'
          ) {
            return
          }

          hidePassword(false)
        }}
        onChange={event => {
          setPasswordInputValue(event.target.value)
        }}
        action={updatePasswordAction}
        loading={setPasswordLoading}
      />
    </div>
  )
}

type MorePopoverProps = {
  id: string
  query: DocumentNode
  share: sidebarGetAlbumShares_album_shares
}

const MorePopover = ({ id, share, query }: MorePopoverProps) => {
  const { t } = useTranslation()

  return (
    <Popover className="relative">
      <Popover.Button
        className="align-middle p-1 ml-2"
        title={t('sidebar.sharing.more', 'More')}
      >
        <MoreIcon />
      </Popover.Button>

      <Popover.Panel>
        <ArrowPopoverPanel width={260}>
          <MorePopoverSectionPassword id={id} share={share} query={query} />
          <div className="px-4 py-2 border-t border-gray-200 dark:border-dark-border mt-2 mb-2">
            <Checkbox label="Expiration date" />
            <TextField className="mt-2 w-full" />
          </div>
        </ArrowPopoverPanel>
      </Popover.Panel>
    </Popover>
  )
}

type SidebarShareAlbumProps = {
  id: string
}

export const SidebarAlbumShare = ({ id }: SidebarShareAlbumProps) => {
  const { t } = useTranslation()

  const {
    loading: queryLoading,
    error: sharesError,
    data: sharesData,
  } = useQuery<sidebarGetAlbumShares, sidebarGetAlbumSharesVariables>(
    SHARE_ALBUM_QUERY,
    { variables: { id } }
  )

  const [shareAlbum, { loading: mutationLoading }] = useMutation<
    sidebarAlbumAddShare,
    sidebarAlbumAddShareVariables
  >(ADD_ALBUM_SHARE_MUTATION, {
    refetchQueries: [{ query: SHARE_ALBUM_QUERY, variables: { id } }],
  })

  const loading = queryLoading || mutationLoading

  if (sharesError) {
    return <div>Error: {sharesError.message}</div>
  }

  if (loading) {
    return <div>{t('general.loading.shares', 'Loading shares...')}</div>
  }

  return (
    <SidebarShare
      id={id}
      isPhoto={false}
      loading={loading}
      shares={sharesData?.album.shares}
      shareItem={shareAlbum}
    />
  )
}
type SidebarSharePhotoProps = {
  id: string
}

export const SidebarPhotoShare = ({ id }: SidebarSharePhotoProps) => {
  const { t } = useTranslation()

  const [
    loadShares,
    { loading: queryLoading, error: sharesError, data: sharesData },
  ] = useLazyQuery<sidebarGetPhotoShares, sidebarGetPhotoSharesVariables>(
    SHARE_PHOTO_QUERY
  )

  const [sharePhoto, { loading: mutationLoading }] = useMutation<
    sidebarPhotoAddShare,
    sidebarPhotoAddShareVariables
  >(ADD_MEDIA_SHARE_MUTATION, {
    refetchQueries: [{ query: SHARE_PHOTO_QUERY, variables: { id } }],
  })

  useEffect(() => {
    if (authToken()) {
      loadShares({
        variables: {
          id,
        },
      })
    }
  }, [])

  const loading = queryLoading || mutationLoading

  if (sharesError) {
    return <div>Error: {sharesError.message}</div>
  }

  if (loading) {
    return <div>{t('general.loading.shares', 'Loading shares...')}</div>
  }

  return (
    <SidebarShare
      id={id}
      isPhoto={true}
      loading={loading}
      shares={sharesData?.media.shares}
      shareItem={sharePhoto}
    />
  )
}

type SidebarShareProps = {
  id: string
  isPhoto: boolean
  loading: boolean
  shares?: sidebarGetPhotoShares_media_shares[]
  shareItem: (item: { variables: { id: string } }) => Promise<unknown>
}

const SidebarShare = ({
  loading,
  shares,
  isPhoto,
  id,
  shareItem,
}: SidebarShareProps) => {
  const { t } = useTranslation()

  const query = isPhoto ? SHARE_PHOTO_QUERY : SHARE_ALBUM_QUERY

  const [deleteShare] = useMutation<
    sidebareDeleteShare,
    sidebareDeleteShareVariables
  >(DELETE_SHARE_MUTATION, {
    refetchQueries: [{ query: query, variables: { id } }],
  })

  if (shares === undefined) {
    return null
  }

  const optionsRows = shares.map(share => (
    <tr
      key={share.token}
      className="border-gray-100 dark:border-dark-border2 border-b border-t"
    >
      <td className="pl-4 py-2 w-full">
        <span className="text-[#585858] dark:text-[#C0C3C4] mr-2">
          <LinkIcon className="inline-block mr-2" />
          <span className="text-xs uppercase font-bold">
            {t('sidebar.sharing.public_link', 'Public Link') + ' '}
          </span>
        </span>
        <span className="text-sm">{share.token}</span>
      </td>
      <td className="pr-6 py-2 whitespace-nowrap text-[#5C6A7F] dark:text-[#7599ca] flex">
        <button
          className="align-middle p-1 ml-2"
          title={t('sidebar.sharing.copy_link', 'Copy Link')}
          onClick={() => {
            copy(`${location.origin}/share/${share.token}`)
          }}
        >
          <CopyIcon />
        </button>
        <button
          onClick={() => {
            deleteShare({ variables: { token: share.token } })
          }}
          className="align-middle p-1 ml-2 hover:text-red-600 focus:text-red-600"
          title={t('sidebar.sharing.delete', 'Delete')}
        >
          <DeleteIcon />
        </button>
        <MorePopover share={share} id={id} query={query} />

        {/* <ShareItemMoreDropdown share={share} id={id} isPhoto={isPhoto} /> */}
      </td>
    </tr>
  ))

  if (optionsRows.length == 0) {
    optionsRows.push(
      <tr
        key="no-shares"
        className="border-gray-100 dark:border-dark-border2 border-b border-t"
      >
        <td
          colSpan={2}
          className="pl-4 py-2 italic text-gray-600 dark:text-gray-300"
        >
          {t('sidebar.sharing.no_shares_found', 'No shares found')}
        </td>
      </tr>
    )
  }

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.sharing.title', 'Sharing options')}
      </SidebarSectionTitle>
      <div>
        <table className="border-collapse w-full">
          <tbody>{optionsRows}</tbody>
          <tfoot>
            <tr className="text-left border-gray-100 dark:border-dark-border2 border-b border-t">
              <td colSpan={2} className="pl-4 py-2">
                <button
                  className="text-green-500 font-bold uppercase text-xs"
                  disabled={loading}
                  onClick={() => {
                    shareItem({
                      variables: {
                        id,
                      },
                    })
                  }}
                >
                  <AddIcon className="inline-block mr-2" />
                  <span>{t('sidebar.sharing.add_share', 'Add shares')}</span>
                </button>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </SidebarSection>
  )
}
