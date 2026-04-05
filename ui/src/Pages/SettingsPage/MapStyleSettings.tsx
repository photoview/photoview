import React, { useRef, useState } from 'react'
import { useQuery, useMutation, gql } from '@apollo/client'
import { SectionTitle, InputLabelTitle, InputLabelDescription } from './SettingsPage'
import { useTranslation } from 'react-i18next'
import { TextField } from '../../primitives/form/Input'
import { mapStyleSettingsQuery } from './__generated__/mapStyleSettingsQuery'
import {
  setMapStyleLight,
  setMapStyleLightVariables,
} from './__generated__/setMapStyleLight'
import {
  setMapStyleDark,
  setMapStyleDarkVariables,
} from './__generated__/setMapStyleDark'

export const MAP_STYLES_QUERY = gql`
  query mapStyleSettingsQuery {
    siteInfo {
      mapStyleLight
      mapStyleDark
    }
  }
`

export const SET_MAP_STYLE_LIGHT_MUTATION = gql`
  mutation setMapStyleLight($url: String!) {
    setMapStyleLight(url: $url)
  }
`

export const SET_MAP_STYLE_DARK_MUTATION = gql`
  mutation setMapStyleDark($url: String!) {
    setMapStyleDark(url: $url)
  }
`

const MapStyleSettings = () => {
  const { t } = useTranslation()

  const lightServerValue = useRef<string | null>(null)
  const darkServerValue = useRef<string | null>(null)
  const [lightUrl, setLightUrl] = useState('')
  const [darkUrl, setDarkUrl] = useState('')

  const stylesQuery = useQuery<mapStyleSettingsQuery>(MAP_STYLES_QUERY, {
    onCompleted(data) {
      setLightUrl(data.siteInfo.mapStyleLight)
      lightServerValue.current = data.siteInfo.mapStyleLight
      setDarkUrl(data.siteInfo.mapStyleDark)
      darkServerValue.current = data.siteInfo.mapStyleDark
    },
  })

  const [setLightMutation, lightMutationData] = useMutation<
    setMapStyleLight,
    setMapStyleLightVariables
  >(SET_MAP_STYLE_LIGHT_MUTATION)

  const [setDarkMutation, darkMutationData] = useMutation<
    setMapStyleDark,
    setMapStyleDarkVariables
  >(SET_MAP_STYLE_DARK_MUTATION)

  const updateLight = (url: string) => {
    if (lightServerValue.current != url) {
      lightServerValue.current = url
      setLightMutation({ variables: { url } })
    }
  }

  const updateDark = (url: string) => {
    if (darkServerValue.current != url) {
      darkServerValue.current = url
      setDarkMutation({ variables: { url } })
    }
  }

  return (
    <div>
      <SectionTitle>
        {t('settings.map_styles.title', 'Map Styles')}
      </SectionTitle>

      <label htmlFor="map_style_light_field">
        <InputLabelTitle>
          {t('settings.map_styles.light_title', 'Light mode style URL')}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.map_styles.light_description',
            'MapLibre style URL used when the UI is in light mode'
          )}
        </InputLabelDescription>
      </label>
      <TextField
        disabled={stylesQuery.loading || lightMutationData.loading}
        id="map_style_light_field"
        value={lightUrl}
        onChange={e => setLightUrl(e.target.value)}
        onBlur={() => updateLight(lightUrl)}
        onKeyDown={e => e.key === 'Enter' && updateLight(lightUrl)}
      />

      <label htmlFor="map_style_dark_field">
        <InputLabelTitle>
          {t('settings.map_styles.dark_title', 'Dark mode style URL')}
        </InputLabelTitle>
        <InputLabelDescription>
          {t(
            'settings.map_styles.dark_description',
            'MapLibre style URL used when the UI is in dark mode'
          )}
        </InputLabelDescription>
      </label>
      <TextField
        disabled={stylesQuery.loading || darkMutationData.loading}
        id="map_style_dark_field"
        value={darkUrl}
        onChange={e => setDarkUrl(e.target.value)}
        onBlur={() => updateDark(darkUrl)}
        onKeyDown={e => e.key === 'Enter' && updateDark(darkUrl)}
      />
    </div>
  )
}

export default MapStyleSettings
