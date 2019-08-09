import React from 'react'
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

export default class PresentView extends React.Component {
  render() {
    const { image, presenting } = this.props

    if (!image || !presenting) {
      return null
    }

    return (
      <PresentContainer>
        <PreventScroll />
        <Query query={imageQuery} variables={{ id: image }}>
          {({ loading, error, data }) => {
            if (loading) return 'Loading...'
            if (error) return error

            const { photo } = data

            console.log(photo)

            return (
              <div>
                <PresentImage
                  src={photo && photo.original.url}
                  onLoad={this.props.imageLoaded && this.props.imageLoaded()}
                />
              </div>
            )
          }}
        </Query>
      </PresentContainer>
    )
  }
}
