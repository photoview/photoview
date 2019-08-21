import React from 'react'
import PropTypes from 'prop-types'
import { Menu, Dropdown, Button } from 'semantic-ui-react'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'

const downloadQuery = gql`
  query sidebarDownloadQuery($photoId: ID!) {
    photo(id: $photoId) {
      id
      downloads {
        title
        url
      }
    }
  }
`

const downloadPhoto = async url => {
  const request = await fetch(url, {
    headers: {
      Authorization: `Bearer ${localStorage.getItem('token')}`,
    },
  })

  const content = await request.blob()
  const contentUrl = URL.createObjectURL(content)

  var downloadAnchor = document.createElement('a', contentUrl)
  downloadAnchor.setAttribute('href', contentUrl)
  downloadAnchor.setAttribute('download', url.match(/[^/]*$/)[0])
  downloadAnchor.click()
}

const SidebarDownload = ({ photoId }) => {
  if (!photoId) return null

  return (
    <div style={{ marginBottom: 24 }}>
      <h2>Download</h2>
      <Query query={downloadQuery} variables={{ photoId }}>
        {({ loading, error, data }) => {
          if (error) return <div>Error {error.message}</div>
          if (!data || !data.photo) return null

          let buttons = data.photo.downloads.map(x => (
            <Button key={x.url} onClick={() => downloadPhoto(x.url)}>
              {x.title}
            </Button>
          ))
          return <Button.Group>{buttons}</Button.Group>
        }}
      </Query>
    </div>
  )
}

SidebarDownload.propTypes = {
  photoId: PropTypes.string,
}

export default SidebarDownload
