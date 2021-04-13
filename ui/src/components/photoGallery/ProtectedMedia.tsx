import React, { DetailedHTMLProps, ImgHTMLAttributes } from 'react'
import { isNil } from '../../helpers/utils'

const isNativeLazyLoadSupported = 'loading' in HTMLImageElement.prototype
const placeholder =
  'data:image/gif;base64,R0lGODlhAQABAPAAAAAAAAAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=='

const getProtectedUrl = (url?: string) => {
  if (url == undefined) return undefined

  const imgUrl = new URL(url, location.origin)

  const tokenRegex = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
  if (tokenRegex) {
    const token = tokenRegex[1]
    imgUrl.searchParams.set('token', token)
  }

  return imgUrl.href
}

export interface ProtectedImageProps
  extends Omit<
    DetailedHTMLProps<ImgHTMLAttributes<HTMLImageElement>, HTMLImageElement>,
    'src'
  > {
  src?: string
  key?: string
  lazyLoading?: boolean
  loaded?: boolean
}

/**
 * An image that needs authorization to load
 * Set lazyLoading to true if you want the image to be loaded once it enters the viewport
 * Native lazy load via HTMLImageElement.loading attribute will be preferred if it is supported by the browser,
 * otherwise IntersectionObserver will be used.
 */
export const ProtectedImage = ({
  src,
  key,
  lazyLoading,
  loaded,
  ...props
}: ProtectedImageProps) => {
  const lazyLoadProps: { 'data-src'?: string; loading?: 'lazy' | 'eager' } = {}

  if (!isNativeLazyLoadSupported && lazyLoading) {
    lazyLoadProps['data-src'] = getProtectedUrl(src)
  }

  if (isNativeLazyLoadSupported && lazyLoading) {
    lazyLoadProps.loading = 'lazy'
  }

  const imgSrc: string =
    lazyLoading && !isNativeLazyLoadSupported
      ? placeholder
      : getProtectedUrl(src) || placeholder

  const loadedProp =
    loaded !== undefined ? { loaded: loaded.toString() } : undefined

  return (
    <img
      key={key}
      {...props}
      {...lazyLoadProps}
      {...loadedProp}
      src={imgSrc}
      crossOrigin="use-credentials"
    />
  )
}

export interface ProtectedVideoProps_Media {
  __typename: 'Media'
  id: string
  thumbnail?: null | {
    __typename: 'MediaURL'
    url: string
  }
  videoWeb?: null | {
    __typename: 'MediaURL'
    url: string
  }
}

export interface ProtectedVideoProps {
  media: ProtectedVideoProps_Media
}

export const ProtectedVideo = ({ media, ...props }: ProtectedVideoProps) => {
  if (isNil(media.videoWeb)) {
    console.error('ProetctedVideo called with media.videoWeb = null')
    return null
  }

  return (
    <video
      {...props}
      controls
      key={media.id}
      crossOrigin="use-credentials"
      poster={getProtectedUrl(media.thumbnail?.url)}
    >
      <source src={getProtectedUrl(media.videoWeb.url)} type="video/mp4" />
    </video>
  )
}
