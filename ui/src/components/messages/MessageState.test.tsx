import { renderHook, act } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import { useMessageState, MessageProvider } from './MessageState'
import { NotificationType } from '../../__generated__/globalTypes'

describe('MessageState', () => {
    let originalDateNow: () => number
    let clearIntervalSpy: any

    beforeEach(() => {
        originalDateNow = Date.now
        try {
            vi.useFakeTimers()
            clearIntervalSpy = vi.spyOn(window, 'clearInterval')
        } catch (error) {
            Date.now = originalDateNow
            throw error
        }
    })

    afterEach(() => {
        try {
            vi.useRealTimers()
            clearIntervalSpy.mockRestore()
        } finally {
            Date.now = originalDateNow
        }
    })

    const wrapper = ({ children }: { children: React.ReactNode }) => (
        <MessageProvider>{children}</MessageProvider>
    )

    it('should cleanup messages older than 24 hours', async () => {
        // Mock current time
        const now = 1643673600000 // 2022-02-01T00:00:00.000Z
        Date.now = vi.fn(() => now)

        const { result } = renderHook(() => useMessageState(), { wrapper })

        // Add messages with different timestamps
        await act(async () => {
            result.current.add({
                key: 'old',
                type: NotificationType.Message,
                props: {
                    header: 'Old Message',
                    content: 'This message is old'
                }
            })
        })

        // Advance time by 25 hours to trigger cleanup
        await act(async () => {
            Date.now = vi.fn(() => now + 25 * 60 * 60 * 1000)
            vi.runOnlyPendingTimers()
        })

        expect(result.current.messages).toHaveLength(0)
    })

    it('should handle concurrent message additions', async () => {
        const now = 1643673600000
        Date.now = vi.fn(() => now)

        const { result } = renderHook(() => useMessageState(), { wrapper })

        await act(async () => {
            // Add multiple messages concurrently
            result.current.add({
                key: 'msg1',
                type: NotificationType.Message,
                props: { header: 'Message 1', content: 'Content 1' }
            })
            result.current.add({
                key: 'msg2',
                type: NotificationType.Message,
                props: { header: 'Message 2', content: 'Content 2' }
            })
        })

        expect(result.current.messages).toHaveLength(2)
        expect(result.current.messages[0].key).toBe('msg1')
        expect(result.current.messages[1].key).toBe('msg2')
    })

    it('should add timestamp to new messages', async () => {
        const now = 1643673600000
        Date.now = vi.fn(() => now)

        const { result } = renderHook(() => useMessageState(), { wrapper })

        await act(async () => {
            result.current.add({
                key: 'test',
                type: NotificationType.Message,
                props: {
                    header: 'Test Message',
                    content: 'Test content'
                }
            })
        })

        expect(result.current.messages).toHaveLength(1)
        expect(result.current.messages[0].timestamp).toBe(now)
    })

    it('should clear interval on unmount', async () => {
        const { unmount } = renderHook(() => useMessageState(), { wrapper })

        await act(async () => {
            unmount()
        })

        expect(clearIntervalSpy).toHaveBeenCalled()
    })

    it('should run cleanup every hour', async () => {
        const now = 1643673600000
        Date.now = vi.fn(() => now)

        const { result } = renderHook(() => useMessageState(), { wrapper })

        await act(async () => {
            result.current.add({
                key: 'test',
                type: NotificationType.Message,
                props: {
                    header: 'Test Message',
                    content: 'Test content'
                }
            })
        })

        // Message should still be present
        expect(result.current.messages).toHaveLength(1)

        // Advance time by 23 hours
        await act(async () => {
            Date.now = vi.fn(() => now + 23 * 60 * 60 * 1000)
            vi.runOnlyPendingTimers()
        })

        // Message should still be present (24.5 hours old)
        expect(result.current.messages).toHaveLength(1)

        // Advance time by 2 more hours
        await act(async () => {
            Date.now = vi.fn(() => now + 25 * 60 * 60 * 1000)
            vi.runOnlyPendingTimers()
        })

        // Message should be removed (25.5 hours old)
        expect(result.current.messages).toHaveLength(0)
    })
})
