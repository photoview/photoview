import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { globalMessageHandler } from './globalMessageHandler'
import { NotificationType } from '../../__generated__/globalTypes'

// Reset the module between tests to ensure a clean singleton
vi.resetModules()

describe('GlobalMessageHandler', () => {
    const mockAdd = vi.fn()
    const mockRemoveKey = vi.fn()

    beforeEach(() => {
        vi.resetAllMocks()
        // Create a private copy of the handler to reset its state
        // @ts-ignore - accessing private property for testing
        globalMessageHandler.isInitialized = false
        // @ts-ignore - accessing private property for testing
        globalMessageHandler.messageState = null
        // @ts-ignore - accessing private property for testing
        globalMessageHandler.pendingMessages = []
    })

    afterEach(() => {
        vi.restoreAllMocks()
    })

    it('should queue messages when not initialized', () => {
        const testMessage = {
            key: 'test-key',
            type: NotificationType.Message,
            props: {
                header: 'Test Header',
                content: 'Test Content'
            }
        }

        globalMessageHandler.add(testMessage)

        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages).toHaveLength(1)
        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages[0]).toEqual(testMessage)
        expect(mockAdd).not.toHaveBeenCalled()
    })

    it('should process messages directly after initialization', () => {
        globalMessageHandler.initialize({ add: mockAdd, removeKey: mockRemoveKey })

        const testMessage = {
            key: 'test-key',
            type: NotificationType.Message,
            props: {
                header: 'Test Header',
                content: 'Test Content'
            }
        }

        globalMessageHandler.add(testMessage)

        expect(mockAdd).toHaveBeenCalledWith(testMessage)
        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages).toHaveLength(0)
    })

    it('should process pending messages upon initialization', () => {
        const testMessage1 = {
            key: 'test-key-1',
            type: NotificationType.Message,
            props: {
                header: 'Test Header 1',
                content: 'Test Content 1'
            }
        }

        const testMessage2 = {
            key: 'test-key-2',
            type: NotificationType.Message,
            props: {
                header: 'Test Header 2',
                content: 'Test Content 2'
            }
        }

        // Add messages before initialization
        globalMessageHandler.add(testMessage1)
        globalMessageHandler.add(testMessage2)

        // Initialize the handler
        globalMessageHandler.initialize({ add: mockAdd, removeKey: mockRemoveKey })

        // Verify pending messages were processed
        expect(mockAdd).toHaveBeenCalledTimes(2)
        expect(mockAdd).toHaveBeenCalledWith(testMessage1)
        expect(mockAdd).toHaveBeenCalledWith(testMessage2)
        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages).toHaveLength(0)
    })

    it('should remove messages by key when initialized', () => {
        globalMessageHandler.initialize({ add: mockAdd, removeKey: mockRemoveKey })

        globalMessageHandler.removeKey('test-key')

        expect(mockRemoveKey).toHaveBeenCalledWith('test-key')
    })

    it('should remove pending messages by key when not initialized', () => {
        const testMessage1 = {
            key: 'keep-this-message',
            type: NotificationType.Message,
            props: {
                header: 'Keep this',
                content: 'This should stay'
            }
        }

        const testMessage2 = {
            key: 'remove-this-message',
            type: NotificationType.Message,
            props: {
                header: 'Remove this',
                content: 'This should be removed'
            }
        }

        // Add messages before initialization
        globalMessageHandler.add(testMessage1)
        globalMessageHandler.add(testMessage2)

        // Remove one message
        globalMessageHandler.removeKey('remove-this-message')

        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages).toHaveLength(1)
        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages[0].key).toBe('keep-this-message')
    })

    it('should create and add error messages', () => {
        globalMessageHandler.initialize({ add: mockAdd, removeKey: mockRemoveKey })

        globalMessageHandler.addErrorMessage('Error Header', 'Error Content')

        expect(mockAdd).toHaveBeenCalledTimes(1)
        expect(mockAdd).toHaveBeenCalledWith(expect.objectContaining({
            type: NotificationType.Message,
            props: {
                negative: true,
                header: 'Error Header',
                content: 'Error Content'
            }
        }))
    })

    it('should queue error messages when not initialized', () => {
        globalMessageHandler.addErrorMessage('Error Header', 'Error Content')

        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages).toHaveLength(1)
        // @ts-ignore - accessing private property for testing
        expect(globalMessageHandler.pendingMessages[0]).toMatchObject({
            type: NotificationType.Message,
            props: {
                negative: true,
                header: 'Error Header',
                content: 'Error Content'
            }
        })
    })
})
