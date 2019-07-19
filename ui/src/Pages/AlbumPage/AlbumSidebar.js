import React, { Component } from 'react'
import styled from 'styled-components'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'

const photoQuery = gql`
  query sidebarPhoto($id: ID) {
    photo(id: $id) {
      title
      original {
        path
        width
        height
      }
    }
  }
`

const RightSidebar = styled.div`
  height: 100%;
  width: 500px;
  position: fixed;
  right: 0;
  top: 60px;
  background-color: white;
  padding: 12px;
  border-left: 1px solid #eee;
`

const PreviewImage = styled.img`
  width: 100%;
  height: 333px;
  object-fit: contain;
`

const Name = styled.div`
  text-align: center;
  font-size: 16px;
`

class AlbumSidebar extends Component {
  render() {
    const { imageId } = this.props

    if (!imageId) {
      return <RightSidebar />
    }

    return (
      <RightSidebar>
        <Query query={photoQuery} variables={{ id: imageId }}>
          {({ loading, error, data }) => {
            if (loading) return 'Loading...'
            if (error) return error

            const { photo } = data

            return (
              <div>
                <PreviewImage src={photo.original.path} />
                <Name>{photo.title}</Name>
              </div>
            )
          }}
        </Query>
      </RightSidebar>
    )
  }
}

export default AlbumSidebar
