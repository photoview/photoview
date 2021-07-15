/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { NotificationType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL subscription operation: notificationSubscription
// ====================================================

export interface notificationSubscription_notification {
  __typename: 'Notification'
  key: string
  type: NotificationType
  header: string
  content: string
  progress: number | null
  positive: boolean
  negative: boolean
  /**
   * Time in milliseconds before the notification will close
   */
  timeout: number | null
}

export interface notificationSubscription {
  notification: notificationSubscription_notification
}
