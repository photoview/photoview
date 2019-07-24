import React, { Component } from 'react'
import styled from 'styled-components'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import { SidebarItem } from './SidebarItem'

const photoQuery = gql`
  query sidebarPhoto($id: ID) {
    photo(id: $id) {
      title
      original {
        url
        width
        height
      }
      exif {
        camera
        maker
        lens
        dateShot {
          formatted
        }
        fileSize
        exposure
        aperture
        iso
        focalLength
        flash
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
  margin-bottom: 12px;
`

const exifNameLookup = {
  camera: 'Camera',
  maker: 'Maker',
  lens: 'Lens',
  dateShot: 'Date Shot',
  fileSize: 'File Size',
  exposure: 'Exposure',
  aperture: 'Aperture',
  iso: 'ISO',
  focalLength: 'Focal Length',
  flash: 'Flash',
}

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

            let exifItems = []

            if (photo.exif) {
              let exifKeys = Object.keys(photo.exif).filter(
                x => !!photo.exif[x] && x != '__typename'
              )

              let exif = exifKeys.reduce(
                (prev, curr) => ({
                  ...prev,
                  [curr]: photo.exif[curr],
                }),
                {}
              )

              exif.dateShot = exif.dateShot.formatted

              exifItems = exifKeys.map(key => (
                <SidebarItem
                  key={key}
                  name={exifNameLookup[key]}
                  value={exif[key]}
                />
              ))
            }

            return (
              <div>
                <PreviewImage src={photo.original && photo.original.url} />
                <Name>{photo.title}</Name>
                <div>{exifItems}</div>
              </div>
            )
          }}
        </Query>
      </RightSidebar>
    )
  }
}

export default AlbumSidebar
