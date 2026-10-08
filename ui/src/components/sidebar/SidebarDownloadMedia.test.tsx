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

import { MessageState } from '../messages/Messages'

describe('downloadMediaShowProgress', () => {
  const t = (s: string) => s

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it.each([
    {
      scenario: 'reader is null',
      contentLength: null,
    },
    {
      scenario: 'totalBytes <= 0',
      contentLength: '0',
    },
    {
      scenario: 'totalBytes is NaN',
      contentLength: 'not-a-number',
    },
  ])(
    'should call MessageState.add and return when $scenario',
    async ({ contentLength }) => {
      const response = {
        body: {
          getReader: () => null,
        },
        headers: {
          get: () => contentLength,
        },
      } as unknown as Pick<Response, 'body' | 'headers'>

      await downloadMediaShowProgress(t)(response as Response)

      // eslint-disable-next-line @typescript-eslint/unbound-method
      expect(MessageState.add).toHaveBeenCalledWith(
        expect.objectContaining({
          type: NotificationType.Close,
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          props: expect.objectContaining({
            negative: true,
            header: 'Downloading media failed',
          }),
        })
      )
    }
  )

  it('should progress if content length is not null', async () => {
    const response = {
      body: {
        getReader: () => ({
          read: () => ({ done: true }),
        }),
      },
      headers: {
        get: () => '100',
      },
    } as unknown as Pick<Response, 'body' | 'headers'>

    await downloadMediaShowProgress(t)(response as Response)

    // eslint-disable-next-line @typescript-eslint/unbound-method
    expect(MessageState.add).toHaveBeenCalledWith(
      expect.objectContaining({
        type: NotificationType.Progress,
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        props: expect.objectContaining({
          header: 'Downloading photo',
        }),
      })
    )
  })
})
