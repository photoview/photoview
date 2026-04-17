import React, { useRef, useState } from 'react'
import { useQuery, useMutation, gql } from '@apollo/client'
import { SectionTitle, InputLabelTitle, InputLabelDescription } from './SettingsPage'
import { useTranslation } from 'react-i18next'
import { TextField } from '../../primitives/form/Input'
import Checkbox from '../../primitives/form/Checkbox'
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
  mutation setMapStyleLight($url: String) {
    setMapStyleLight(url: $url)
  }
`

export const SET_MAP_STYLE_DARK_MUTATION = gql`
  mutation setMapStyleDark($url: String) {
    setMapStyleDark(url: $url)
  }
`

const MapStyleSettings = () => {
  const { t } = useTranslation()

  const lightServerValue = useRef<string | null>(null)
  const darkServerValue = useRef<string | null>(null)
  const [lightUrl, setLightUrl] = useState('')
  const [darkUrl, setDarkUrl] = useState('')
  const [customizeStyle, setCustomizeStyle] = useState(false)

  const stylesQuery = useQuery<mapStyleSettingsQuery>(MAP_STYLES_QUERY, {
    onCompleted(data) {
      const light = data.siteInfo.mapStyleLight
      const dark = data.siteInfo.mapStyleDark
      lightServerValue.current = light
      darkServerValue.current = dark
      setLightUrl(light ?? '')
      setDarkUrl(dark ?? '')
      setCustomizeStyle(light != null || dark != null)
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

  const updateLight = (url: string | null) => {
    if (lightServerValue.current !== url) {
      lightServerValue.current = url
      setLightMutation({ variables: { url } })
    }
  }

  const updateDark = (url: string | null) => {
    if (darkServerValue.current !== url) {
      darkServerValue.current = url
      setDarkMutation({ variables: { url } })
    }
  }

  const handleToggleCustomize = (checked: boolean) => {
    setCustomizeStyle(checked)
    if (!checked) {
      setLightUrl('')
      setDarkUrl('')
      updateLight(null)
      updateDark(null)
    }
  }

  return (
    <div>
      <SectionTitle>
        {t('settings.map_styles.title', 'Map Styles')}
      </SectionTitle>

      <Checkbox
        label={t(
          'settings.map_styles.customize_label',
          'Customize map style'
        )}
        disabled={stylesQuery.loading}
        checked={customizeStyle}
        onChange={e => handleToggleCustomize(e.target.checked)}
        className="mb-4"
      />

      {!customizeStyle && (
        <InputLabelDescription>
          {t(
            'settings.map_styles.builtin_description',
            'Using the built-in map style with hillshading. Enable customization to use your own style URLs.'
          )}
        </InputLabelDescription>
      )}

      {customizeStyle && (
        <>
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
            onBlur={() => updateLight(lightUrl || null)}
            onKeyDown={e => e.key === 'Enter' && updateLight(lightUrl || null)}
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
            onBlur={() => updateDark(darkUrl || null)}
            onKeyDown={e => e.key === 'Enter' && updateDark(darkUrl || null)}
          />
        </>
      )}
    </div>
  )
}

export default MapStyleSettings
