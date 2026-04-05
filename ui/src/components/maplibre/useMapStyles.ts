import { gql, useQuery } from '@apollo/client'
import { mapStylesQuery } from './__generated__/mapStylesQuery'

export const DEFAULT_STYLE_LIGHT = 'https://tiles.openfreemap.org/styles/positron'
export const DEFAULT_STYLE_DARK = 'https://tiles.openfreemap.org/styles/dark'

const MAP_STYLES_QUERY = gql`
  query mapStylesQuery {
    siteInfo {
      mapStyleLight
      mapStyleDark
    }
  }
`

const useMapStyles = () => {
  const { data } = useQuery<mapStylesQuery>(MAP_STYLES_QUERY)

  return {
    mapStyleLight: data?.siteInfo.mapStyleLight ?? DEFAULT_STYLE_LIGHT,
    mapStyleDark: data?.siteInfo.mapStyleDark ?? DEFAULT_STYLE_DARK,
  }
}

export default useMapStyles
