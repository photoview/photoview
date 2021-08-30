import classNames from 'classnames'
import React, { DetailedHTMLProps, ImgHTMLAttributes } from 'react'
import { useRef } from 'react'
import { useState } from 'react'
import { useEffect } from 'react'
import { isNil } from '../../helpers/utils'

const isNativeLazyLoadSupported = 'loading' in document.createElement('img')
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
}

/**
 * An image that needs authorization to load
 * Set lazyLoading to true if you want the image to be loaded once it enters the viewport
 * Native lazy load via HTMLImageElement.loading attribute will be preferred if it is supported by the browser,
 * otherwise IntersectionObserver will be used.
 */
export const ProtectedImage = ({
  src,
  lazyLoading,
  ...props
}: ProtectedImageProps) => {
  const url = getProtectedUrl(src) || placeholder

  if (!lazyLoading) {
    return (
      <img {...props} src={url} loading="eager" crossOrigin="use-credentials" />
    )
  }

  if (!isNativeLazyLoadSupported) {
    return <FallbackLazyloadedImage src={url} {...props} />
  }

  // load with native lazy loading
  return (
    <img {...props} src={url} loading="lazy" crossOrigin="use-credentials" />
  )
}

interface FallbackLazyloadedImageProps
  extends Omit<
    DetailedHTMLProps<ImgHTMLAttributes<HTMLImageElement>, HTMLImageElement>,
    'src'
  > {
  src?: string
}

const FallbackLazyloadedImage = ({
  src,
  ...props
}: FallbackLazyloadedImageProps) => {
  const [inView, setInView] = useState(false)
  const imgRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const imgElm = imgRef.current
    if (isNil(imgElm) || inView) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setInView(true)
          observer.disconnect()
        }
      },
      {
        root: null,
        threshold: 0,
      }
    )

    observer.observe(imgElm)

    return () => {
      observer.disconnect()
    }
  }, [imgRef])

  if (inView) {
    return <img {...props} src={src} crossOrigin="use-credentials" />
  } else {
    return (
      <div
        ref={imgRef}
        className={classNames(props.className, 'bg-[#eee]')}
      ></div>
    )
  }
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
