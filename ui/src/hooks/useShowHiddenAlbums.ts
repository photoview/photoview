import { gql, useQuery } from '@apollo/client'
import { authToken } from '../helpers/authentication'
import { showHiddenAlbumsPreferenceQuery } from './__generated__/showHiddenAlbumsPreferenceQuery'

export const SHOW_HIDDEN_ALBUMS_PREFERENCE_QUERY = gql`
  query showHiddenAlbumsPreferenceQuery {
    myUserPreferences {
      id
      showHiddenAlbums
    }
  }
`

// Whether the logged-in user wants personally-hidden albums revealed
// (dimmed, with a click-to-unhide affordance) instead of excluded from
// navigation and search.
const useShowHiddenAlbums = (): boolean => {
  const token = authToken()
  const { data } = useQuery<showHiddenAlbumsPreferenceQuery>(
    SHOW_HIDDEN_ALBUMS_PREFERENCE_QUERY,
    { skip: !token }
  )

  return !!token && (data?.myUserPreferences.showHiddenAlbums ?? false)
}

export default useShowHiddenAlbums
