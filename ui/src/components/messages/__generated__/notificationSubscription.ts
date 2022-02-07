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
  /**
   * A key used to identify the notification, new notification updates with the same key, should replace the old notifications
   */
  key: string
  type: NotificationType
  /**
   * The text for the title of the notification
   */
  header: string
  /**
   * The text for the body of the notification
   */
  content: string
  /**
   * A value between 0 and 1 when the notification type is `Progress`
   */
  progress: number | null
  /**
   * Whether or not the message of the notification is positive, the UI might reflect this with a green color
   */
  positive: boolean
  /**
   * Whether or not the message of the notification is negative, the UI might reflect this with a red color
   */
  negative: boolean
  /**
   * Time in milliseconds before the notification should close
   */
  timeout: number | null
}

export interface notificationSubscription {
  notification: notificationSubscription_notification
}
