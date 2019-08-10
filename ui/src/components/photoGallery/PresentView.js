import React, { useState } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import ProtectedImage from './ProtectedImage'

const PresentContainer = styled.div`
  position: fixed;
  width: 100vw;
  height: 100vh;
  background-color: black;
  color: white;
  top: 0;
  left: 0;
  z-index: 100;
`

const PreventScroll = createGlobalStyle`
  body {
    /* height: 100vh !important; */
    overflow: hidden;
  }
`

const imageQuery = gql`
  query presentImage($id: ID) {
    photo(id: $id) {
      title
      original {
        width
        height
        url
      }
    }
  }
`

const PresentImage = styled(ProtectedImage)`
  width: 100vw;
  height: 100vh;
  object-fit: contain;
  object-position: center;
`

const PresentView = ({
  image,
  presenting,
  thumbnail,
  imageLoaded: imageLoadedCallback,
}) => {
  if (!image || !presenting) {
    return null
  }

  return (
    <PresentContainer>
      <PreventScroll />
      <Query query={imageQuery} variables={{ id: image }}>
        {({ loading, error, data }) => {
          if (error) return error

          let original = null
          if (!loading) {
            const { photo } = data
            original = (
              <PresentImage
                src={photo && photo.original.url}
                onLoad={imageLoadedCallback && imageLoadedCallback()}
              />
            )
          }

          return (
            <div>
              <PresentImage src={thumbnail} />
              {original}
            </div>
          )
        }}
      </Query>
    </PresentContainer>
  )
}

export default PresentView
