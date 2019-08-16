import React, { useEffect, useState } from 'react'
import PropTypes from 'prop-types'
import styled, { createGlobalStyle } from 'styled-components'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import ProtectedImage from './ProtectedImage'

export const PresentContainer = ({ children, ...otherProps }) => {
  const StyledContainer = styled.div`
    position: fixed;
    width: 100vw;
    height: 100vh;
    background-color: black;
    color: white;
    top: 0;
    left: 0;
    z-index: 100;
  `

  return (
    <StyledContainer {...otherProps}>
      <PreventScroll />
      {children}
    </StyledContainer>
  )
}

PresentContainer.propTypes = {
  children: PropTypes.element,
}

const PreventScroll = createGlobalStyle`
  body {
    /* height: 100vh !important; */
    overflow: hidden;
  }
`

const imageQuery = gql`
  query presentImage($id: ID!) {
    photo(id: $id) {
      id
      title
      original {
        width
        height
        url
      }
    }
  }
`

const StyledPhoto = styled(ProtectedImage)`
  width: 100vw;
  height: 100vh;
  object-fit: contain;
  object-position: center;
`

export const PresentPhoto = ({
  photo,
  thumbnail,
  imageLoaded,
  photoId,
  ...otherProps
}) => {
  let [originalPhoto, setOriginalPhoto] = useState(null)
  useEffect(() => {
    if (!photoId) return

    function loadOriginalPhoto() {
      const originalPhoto = (
        <Query query={imageQuery} variables={{ id: photoId }}>
          {({ loading, error, data }) => {
            if (error) {
              alert(error)
              return null
            }

            if (data && data.photo) {
              const photo = data.photo

              return (
                <StyledPhoto
                  style={{ display: 'none' }}
                  src={photo.original.url}
                  onLoad={e => {
                    e.target.style.display = 'initial'
                    imageLoaded && imageLoaded()
                  }}
                />
              )
            }

            return null
          }}
        </Query>
      )

      setOriginalPhoto(originalPhoto)
    }

    const timeoutHandle = setTimeout(loadOriginalPhoto, 500)

    return function cleanup() {
      clearTimeout(timeoutHandle)
    }
  }, [])

  return (
    <div {...otherProps}>
      {originalPhoto}
      <StyledPhoto src={thumbnail} />
    </div>
  )
}

PresentPhoto.propTypes = {
  photo: PropTypes.object,
  thumbnail: PropTypes.string,
  imageLoaded: PropTypes.func,
  photoId: PropTypes.string,
}
