/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: sidebarDownloadQuery
// ====================================================

export interface sidebarDownloadQuery_media_downloads_mediaUrl {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
  /**
   * Width of the image in pixels
   */
  width: number
  /**
   * Height of the image in pixels
   */
  height: number
  /**
   * The file size of the resource in bytes
   */
  fileSize: number
}

export interface sidebarDownloadQuery_media_downloads {
  __typename: 'MediaDownload'
  /**
   * A description of the role of the media file
   */
  title: string
  mediaUrl: sidebarDownloadQuery_media_downloads_mediaUrl
}

export interface sidebarDownloadQuery_media {
  __typename: 'Media'
  id: string
  /**
   * A list of different versions of files for this media that can be downloaded by the user
   */
  downloads: sidebarDownloadQuery_media_downloads[]
}

export interface sidebarDownloadQuery {
  /**
   * Get media by id, user must own the media or be admin.
   * If valid tokenCredentials are provided, the media may be retrived without further authentication
   */
  media: sidebarDownloadQuery_media
}

export interface sidebarDownloadQueryVariables {
  mediaId: string
}
