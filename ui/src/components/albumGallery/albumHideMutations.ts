import { gql, useMutation } from '@apollo/client'
import { hideAlbum, hideAlbumVariables } from './__generated__/hideAlbum'

const HIDE_ALBUM_MUTATION = gql`
  mutation hideAlbum($albumId: ID!, $hidden: Boolean!) {
    hideAlbum(albumId: $albumId, hidden: $hidden) {
      id
      viewerHidden
    }
  }
`

export const useHideAlbumMutation = (refetchQueries?: string[]) =>
  useMutation<hideAlbum, hideAlbumVariables>(HIDE_ALBUM_MUTATION, {
    refetchQueries,
    // Without this, a rejected mutation is an unhandled promise rejection.
    // The global Apollo error link already toasts it, and the optimistic
    // response above is rolled back automatically on failure.
    onError: () => undefined,
  })

export const toggleAlbumHidden = (
  hideAlbumFn: ReturnType<typeof useHideAlbumMutation>[0],
  albumId: string,
  currentlyHidden: boolean
) =>
  hideAlbumFn({
    variables: { albumId, hidden: !currentlyHidden },
    optimisticResponse: {
      hideAlbum: {
        __typename: 'Album',
        id: albumId,
        viewerHidden: !currentlyHidden,
      },
    },
  })
