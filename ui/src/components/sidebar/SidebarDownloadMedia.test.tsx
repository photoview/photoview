import { describe, it, vi, beforeEach, expect } from 'vitest'
import { NotificationType } from '../../__generated__/globalTypes'
import { downloadMediaShowProgress } from './SidebarDownloadMedia'

vi.mock('../messages/Messages', () => {
  const MessageState = {
    add: vi.fn(),
    removeKey: vi.fn(),
  }

  return {
    MessageState,
  }
})

vi.mock('Response', () => {
  return {
    body: {
      getReader: vi.fn(),
    },
    headers: {
      get: vi.fn(),
    },
  }
})

import { MessageState } from '../messages/Messages'
import { Message } from '../messages/SubscriptionsHook'

describe('downloadMediaShowProgress', () => {
  const t = (s: string) => s

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call MessageState.add and return when reader is null', async () => {
    const response = {
      body: {
        getReader: () => null,
      },
      headers: {
        get: () => null,
      },
    } as never

    await downloadMediaShowProgress(t)(response)
    // eslint-disable-next-line @typescript-eslint/unbound-method
    expect(MessageState.add).toHaveBeenCalledWith(
      expect.objectContaining({
        type: NotificationType.Close,
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        props: expect.objectContaining({
          negative: true,
          header: 'Downloading media failed',
        }),
      } as Message)
    )
  })

  it('should call MessageState.add and return when totalBytes <= 0', async () => {
    const response = {
      body: {
        getReader: () => null,
      },
      headers: {
        get: () => '0',
      },
    } as never

    await downloadMediaShowProgress(t)(response)
    // eslint-disable-next-line @typescript-eslint/unbound-method
    expect(MessageState.add).toHaveBeenCalledWith(
      expect.objectContaining({
        type: NotificationType.Close,
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        props: expect.objectContaining({
          negative: true,
          header: 'Downloading media failed',
        }),
      } as Message)
    )
  })

  it('should call MessageState.add and return when totalBytes is NaN', async () => {
    const response = {
      body: {
        getReader: () => null,
      },
      headers: {
        get: () => 'not-a-number',
      },
    } as never

    await downloadMediaShowProgress(t)(response)
    // eslint-disable-next-line @typescript-eslint/unbound-method
    expect(MessageState.add).toHaveBeenCalledWith(
      expect.objectContaining({
        type: NotificationType.Close,
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        props: expect.objectContaining({
          negative: true,
          header: 'Downloading media failed',
        }),
      } as Message)
    )
  })
})
