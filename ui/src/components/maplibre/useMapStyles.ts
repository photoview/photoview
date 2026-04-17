import { gql, useQuery } from '@apollo/client'
import type maplibregl from 'maplibre-gl'
import { mapStylesQuery } from './__generated__/mapStylesQuery'
import libertyStyle from './libertyStyle.json'

export type MapStyle = NonNullable<maplibregl.MapOptions['style']>

// Built-in style with hillshading, served locally
export const BUILTIN_STYLE = libertyStyle as unknown as MapStyle

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
    mapStyleLight: data?.siteInfo.mapStyleLight ?? BUILTIN_STYLE,
    mapStyleDark: data?.siteInfo.mapStyleDark ?? BUILTIN_STYLE,
  }
}

export default useMapStyles
